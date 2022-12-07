// Copyright 2021 loxilb Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connections

import (
	"fmt"
	"time"

	tk "github.com/loxilb-io/loxilib"

	"github.com/loxilb-io/loxilb/loxiexporter/flowexporter"
	"github.com/loxilb-io/loxilb/loxiexporter/flowexporter/priorityqueue"

	"github.com/loxilb-io/loxilb/loxiexporter/metrics"
)

type ConntrackConnectionStore struct {
	connDumper            ConnTrackDumper
	v4Enabled             bool
	v6Enabled             bool
	pollInterval          time.Duration
	connectUplinkToBridge bool
	connectionStore
}

func NewConntrackConnectionStore(
	connTrackDumper ConnTrackDumper,
	v4Enabled bool,
	v6Enabled bool,
	o *flowexporter.FlowExporterOptions,
) *ConntrackConnectionStore {
	return &ConntrackConnectionStore{
		connDumper:            connTrackDumper,
		v4Enabled:             v4Enabled,
		v6Enabled:             v6Enabled,
		pollInterval:          o.PollInterval,
		connectionStore:       NewConnectionStore(o),
		connectUplinkToBridge: o.ConnectUplinkToBridge,
	}
}

// Run enables the periodical polling of conntrack connections at a given flowPollInterval.
func (cs *ConntrackConnectionStore) Run(stopCh <-chan struct{}) {
	tk.LogIt(tk.LogInfo, "Starting conntrack polling\n")

	pollTicker := time.NewTicker(cs.pollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-stopCh:
			break
		case <-pollTicker.C:

			_, err := cs.Poll()
			if err != nil {
				// Not failing here as errors can be transient and could be resolved in future poll cycles.
				// TODO: Come up with a backoff/retry mechanism by increasing poll interval and adding retry timeout
				tk.LogIt(tk.LogError, "Error during conntrack poll cycle: %v\n", err)
			}
		}
	}
}

// Poll calls into conntrackDumper interface to dump conntrack flows. It returns the number of connections for each
// address family, as a slice. In dual-stack clusters, the slice will contain 2 values (number of IPv4 connections first,
// then number of IPv6 connections).
// TODO: As optimization, only poll invalid/closed connections during every poll, and poll the established connections right before the export.
func (cs *ConntrackConnectionStore) Poll() ([]int, error) {
	tk.LogIt(tk.LogInfo, "Polling conntrack\n")

	var connsLens []int
	var totalConns int
	var filteredConnsList []*flowexporter.Connection

	filteredConnsListPerZone, totalConnsPerZone, err := cs.connDumper.DumpFlows()
	if err != nil {
		return []int{}, err
	}
	totalConns += totalConnsPerZone
	filteredConnsList = append(filteredConnsList, filteredConnsListPerZone...)
	connsLens = append(connsLens, len(filteredConnsList))
	// Reset IsPresent flag for all connections in connection map before updating
	// the dumped flows information in connection map. If the connection does not
	// exist in conntrack table and has been exported, then we will delete it from
	// connection map. In addition, if the connection was not exported for a specific
	// time period, then we consider it to be stale and delete it.
	deleteIfStaleOrResetConn := func(key flowexporter.ConnectionKey, conn *flowexporter.Connection) error {
		if !conn.IsPresent {
			// Delete the connection if it is ready to delete or it was not exported
			// in the time period as specified by the stale connection timeout.
			if conn.ReadyToDelete || time.Since(conn.LastExportTime) >= cs.staleConnectionTimeout {
				if removedItem := cs.expirePriorityQueue.Remove(key); removedItem != nil {
					// In case ReadyToDelete is true, item should already have been removed from pq
					tk.LogIt(tk.LogInfo, "Conn removed from cs pq due to stale timeout\n",
						"key", key, "conn", removedItem.Conn)
				}
				if err := cs.deleteConnWithoutLock(key); err != nil {
					return err
				}
			}
		} else {
			conn.IsPresent = false
		}
		return nil
	}

	// Hold the lock until we verify whether the connection exist in conntrack table,
	// and finish updating the connection store.
	cs.AcquireConnStoreLock()

	if err := cs.ForAllConnectionsDoWithoutLock(deleteIfStaleOrResetConn); err != nil {
		cs.ReleaseConnStoreLock()
		return []int{}, err
	}

	// Update only the Connection store. IPFIX records are generated based on Connection store.
	for _, conn := range filteredConnsList {
		cs.AddOrUpdateConn(conn)
	}

	cs.ReleaseConnStoreLock()

	metrics.TotalConnectionsInConnTrackTable.Set(float64(totalConns))
	maxConns, err := cs.connDumper.GetMaxConnections()
	if err != nil {
		return []int{}, err
	}
	metrics.MaxConnectionsInConnTrackTable.Set(float64(maxConns))
	tk.LogIt(tk.LogInfo, "Conntrack polling successful\n")

	return connsLens, nil
}

// AddOrUpdateConn updates the connection if it is already present, i.e., update timestamp, counters etc.,
// or adds a new connection with the resolved K8s metadata.
func (cs *ConntrackConnectionStore) AddOrUpdateConn(conn *flowexporter.Connection) {
	conn.IsPresent = true
	connKey := flowexporter.NewConnectionKey(conn)

	existingConn, exists := cs.connections[connKey]
	if exists {
		existingConn.IsPresent = conn.IsPresent
		if flowexporter.IsConnectionDying(existingConn) {
			return
		}
		// Update the necessary fields that are used in generating flow records.
		// Can same 5-tuple flow get deleted and added to conntrack table? If so use ID.
		existingConn.StopTime = conn.StopTime
		existingConn.OriginalBytes = conn.OriginalBytes
		existingConn.OriginalPackets = conn.OriginalPackets
		existingConn.ReverseBytes = conn.ReverseBytes
		existingConn.ReversePackets = conn.ReversePackets
		existingConn.TCPState = conn.TCPState
		existingConn.IsActive = flowexporter.CheckConntrackConnActive(existingConn)
		if existingConn.IsActive {
			existingItem, exists := cs.expirePriorityQueue.KeyToItem[connKey]
			if !exists {
				// If the connKey:pqItem pair does not exist in the map, it shows the
				// conn was inactive, and was removed from PQ and map. Since it becomes
				// active again now, we create a new pqItem and add it to PQ and map.
				cs.expirePriorityQueue.WriteItemToQueue(connKey, existingConn)
			} else {
				cs.connectionStore.expirePriorityQueue.Update(existingItem, existingItem.ActiveExpireTime,
					time.Now().Add(cs.connectionStore.expirePriorityQueue.IdleFlowTimeout))
			}
		}
		tk.LogIt(tk.LogInfo, "loxilb flow updated connection\n", existingConn)
	} else {
		cs.fillPodInfo(conn)
		// if conn.Mark&openflow.ServiceCTMark.GetRange().ToNXRange().ToUint32Mask() == openflow.ServiceCTMark.GetValue() {
		// 	clusterIP := conn.DestinationServiceAddress.String()
		// 	svcPort := conn.DestinationServicePort
		// 	protocol, err := lookupServiceProtocol(conn.FlowKey.Protocol)
		// 	if err != nil {
		// 		klog.InfoS("Could not retrieve Service protocol", "error", err)
		// 	} else {
		// 		serviceStr := fmt.Sprintf("%s:%d/%s", clusterIP, svcPort, protocol)
		// 		cs.fillServiceInfo(conn, serviceStr)
		// 	}
		// }
		if conn.StartTime.IsZero() {
			conn.StartTime = time.Now()
			conn.StopTime = time.Now()
		}
		conn.LastExportTime = conn.StartTime
		metrics.TotalloxilbConnectionsInConnTrackTable.Inc()
		conn.IsActive = true
		// Add new loxilb connection to connection store and PQ.
		cs.connections[connKey] = conn
		cs.expirePriorityQueue.WriteItemToQueue(connKey, conn)
		tk.LogIt(tk.LogInfo, "New loxilb flow added connection\n", conn)
	}
}

func (cs *ConntrackConnectionStore) GetExpiredConns(expiredConns []flowexporter.Connection, currTime time.Time, maxSize int) ([]flowexporter.Connection, time.Duration) {
	cs.AcquireConnStoreLock()
	defer cs.ReleaseConnStoreLock()
	for i := 0; i < maxSize; i++ {
		pqItem := cs.connectionStore.expirePriorityQueue.GetTopExpiredItem(currTime)
		if pqItem == nil {
			break
		}
		expiredConns = append(expiredConns, *pqItem.Conn)
		if flowexporter.IsConnectionDying(pqItem.Conn) {
			// If a conntrack connection is in dying state or connection is not
			// in the conntrack table, we set the ReadyToDelete flag to true to
			// do the deletion later.
			pqItem.Conn.ReadyToDelete = true
		}
		if pqItem.IdleExpireTime.Before(currTime) {
			// No packets have been received during the idle timeout interval,
			// the connection is therefore considered inactive.
			pqItem.Conn.IsActive = false
		}
		cs.UpdateConnAndQueue(pqItem, currTime)
	}
	return expiredConns, cs.connectionStore.expirePriorityQueue.GetExpiryFromExpirePriorityQueue()
}

// deleteConnWithoutLock deletes the connection from the connection map given
// the connection key without grabbing the lock. Caller is expected to grab lock.
func (cs *ConntrackConnectionStore) deleteConnWithoutLock(connKey flowexporter.ConnectionKey) error {
	_, exists := cs.connections[connKey]
	if !exists {
		return fmt.Errorf("connection with key %v doesn't exist in map", connKey)
	}
	delete(cs.connections, connKey)
	metrics.TotalloxilbConnectionsInConnTrackTable.Dec()
	return nil
}

func (cs *ConntrackConnectionStore) GetPriorityQueue() *priorityqueue.ExpirePriorityQueue {
	return cs.connectionStore.expirePriorityQueue
}
