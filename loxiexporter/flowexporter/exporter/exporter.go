// Copyright 2020 Antrea Authors
// Modified by 2022 NetLOX Inc
// Modified log : Remove K8s integration parts. And added log as like LoxiLB.
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

package exporter

import (
	"fmt"
	"hash/fnv"
	"os"
	"time"

	cmn "github.com/loxilb-io/loxilb/common"

	ipfixentities "github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/exporter"
	ipfixregistry "github.com/vmware/go-ipfix/pkg/registry"

	tk "github.com/loxilb-io/loxilib"

	"github.com/loxilb-io/loxilb/loxiexporter/flowexporter"
	"github.com/loxilb-io/loxilb/loxiexporter/flowexporter/connections"
	"github.com/loxilb-io/loxilb/loxiexporter/flowexporter/priorityqueue"

	"github.com/loxilb-io/loxilb/loxiexporter/metrics"

	"github.com/loxilb-io/loxilb/loxiexporter/ipfix"
)

// When initializing flowExporter, a slice is allocated with a fixed size to
// store expired connections. The advantage is every time we export, the connection
// store lock will only be held for a bounded time. The disadvantages are: 1. the
// constant is independent of actual number of expired connections 2. when the
// number of expired connections goes over the constant, the export can not be
// finished in a single round. It could be delayed by conntrack connections polling
// routine, which also acquires the connection store lock. The possible solution
// can be taking a fraction of the size of connection store to approximate the
// number of expired connections, while having a min and a max to handle edge cases,
// e.g. min(50 + 0.1 * connectionStore.size(), 200)
const maxConnsToExport = 64

var (
	IANAInfoElementsCommon = []string{
		"flowStartSeconds",
		"flowEndSeconds",
		"flowEndReason",
		"sourceTransportPort",
		"destinationTransportPort",
		"protocolIdentifier",
		"packetTotalCount",
		"octetTotalCount",
		"packetDeltaCount",
		"octetDeltaCount",
	}
	IANAInfoElementsIPv4 = append(IANAInfoElementsCommon, []string{"sourceIPv4Address", "destinationIPv4Address"}...)
	IANAInfoElementsIPv6 = append(IANAInfoElementsCommon, []string{"sourceIPv6Address", "destinationIPv6Address"}...)
	// IANAReverseInfoElements contain substring "reverse" which is an indication to get reverse element of go-ipfix library.
	IANAReverseInfoElements = []string{
		"reversePacketTotalCount",
		"reverseOctetTotalCount",
		"reversePacketDeltaCount",
		"reverseOctetDeltaCount",
	}
	loxilbInfoElementsCommon = []string{
		"sourcePodName",
		"sourcePodNamespace",
		"sourceNodeName",
		"destinationPodName",
		"destinationPodNamespace",
		"destinationNodeName",
		"destinationServicePort",
		"destinationServicePortName",
		"ingressNetworkPolicyName",
		"ingressNetworkPolicyNamespace",
		"ingressNetworkPolicyType",
		"ingressNetworkPolicyRuleName",
		"ingressNetworkPolicyRuleAction",
		"egressNetworkPolicyName",
		"egressNetworkPolicyNamespace",
		"egressNetworkPolicyType",
		"egressNetworkPolicyRuleName",
		"egressNetworkPolicyRuleAction",
		"tcpState",
		"flowType",
	}
	LoxiLBInfoElementsIPv4 = append(loxilbInfoElementsCommon, []string{"destinationClusterIPv4"}...)
	LoxiLBInfoElementsIPv6 = append(loxilbInfoElementsCommon, []string{"destinationClusterIPv6"}...)
)

type FlowExporter struct {
	conntrackConnStore     *connections.ConntrackConnectionStore
	denyConnStore          *connections.DenyConnectionStore
	process                ipfix.IPFIXExportingProcess
	elementsListv4         []ipfixentities.InfoElementWithValue
	elementsListv6         []ipfixentities.InfoElementWithValue
	ipfixSet               ipfixentities.Set
	numDataSetsSent        uint64 // used for unit tests.
	templateIDv4           uint16
	templateIDv6           uint16
	registry               ipfix.IPFIXRegistry
	v4Enabled              bool
	v6Enabled              bool
	exporterInput          exporter.ExporterInput
	nodeName               string
	conntrackPriorityQueue *priorityqueue.ExpirePriorityQueue
	denyPriorityQueue      *priorityqueue.ExpirePriorityQueue
	expiredConns           []flowexporter.Connection
}

func genObservationID(nodeName string) uint32 {
	h := fnv.New32()
	h.Write([]byte(nodeName))
	return h.Sum32()
}

func prepareExporterInputArgs(collectorAddr, collectorProto, nodeName string) exporter.ExporterInput {
	expInput := exporter.ExporterInput{}
	// Exporting process requires domain observation ID.
	expInput.ObservationDomainID = genObservationID(nodeName)

	expInput.CollectorAddress = collectorAddr
	if collectorProto == "tls" {
		expInput.IsEncrypted = true
		expInput.CollectorProtocol = "tcp"
	} else {
		expInput.IsEncrypted = false
		expInput.CollectorProtocol = collectorProto
	}

	return expInput
}

func RegisterApiHooks(hooks cmn.NetHookInterface) {
	connections.ApiHooks = hooks
}

func NewFlowExporter(v4Enabled, v6Enabled bool, o *flowexporter.FlowExporterOptions) (*FlowExporter, error) {
	// Initialize IPFIX registry
	registry := ipfix.NewIPFIXRegistry()
	registry.LoadRegistry()

	// Prepare input args for IPFIX exporting process.
	nodeName, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	expInput := prepareExporterInputArgs(o.FlowCollectorAddr, o.FlowCollectorProto, nodeName)

	connTrackDumper := connections.InitializeConnTrackDumper()
	denyConnStore := connections.NewDenyConnectionStore(o)
	conntrackConnStore := connections.NewConntrackConnectionStore(connTrackDumper, v4Enabled, v6Enabled, o)

	return &FlowExporter{
		conntrackConnStore:     conntrackConnStore,
		denyConnStore:          denyConnStore,
		registry:               registry,
		v4Enabled:              v4Enabled,
		v6Enabled:              v6Enabled,
		exporterInput:          expInput,
		ipfixSet:               ipfixentities.NewSet(false),
		nodeName:               nodeName,
		conntrackPriorityQueue: conntrackConnStore.GetPriorityQueue(),
		denyPriorityQueue:      denyConnStore.GetPriorityQueue(),
		expiredConns:           make([]flowexporter.Connection, 0, maxConnsToExport*2),
	}, nil
}

func (exp *FlowExporter) GetDenyConnStore() *connections.DenyConnectionStore {
	return exp.denyConnStore
}

func (exp *FlowExporter) Run(stopCh <-chan struct{}) {
	// Start the goroutine to periodically delete stale deny connections.
	go exp.denyConnStore.RunPeriodicDeletion(stopCh)

	// Start the goroutine to poll conntrack flows.
	go exp.conntrackConnStore.Run(stopCh)

	defaultTimeout := exp.conntrackPriorityQueue.ActiveFlowTimeout
	expireTimer := time.NewTimer(defaultTimeout)
	for {
		select {
		case <-stopCh:
			if exp.process != nil {
				exp.process.CloseConnToCollector()
			}
			expireTimer.Stop()
			return
		case <-expireTimer.C:
			if exp.process == nil {
				err := exp.initFlowExporter()
				if err != nil {
					tk.LogIt(tk.LogError, err.Error(), "Error when initializing flow exporter\n")
					// There could be other errors while initializing flow exporter
					// other than connecting to IPFIX collector, therefore closing
					// the connection and resetting the process.
					if exp.process != nil {
						exp.process.CloseConnToCollector()
						exp.process = nil
					}
					// Initializing flow exporter fails, will retry in next cycle.
					expireTimer.Reset(defaultTimeout)
					continue
				}
			}
			// Pop out the expired connections from the conntrack priority queue
			// and the deny priority queue, and send the data records.
			nextExpireTime, err := exp.sendFlowRecords()
			if err != nil {
				tk.LogIt(tk.LogError, err.Error(), "Error when sending expired flow records\n")
				// If there is an error when sending flow records because of intermittent
				// connectivity, we reset the connection to IPFIX collector and retry
				// in the next export cycle to reinitialize the connection and send flow records.
				exp.process.CloseConnToCollector()
				exp.process = nil
				expireTimer.Reset(defaultTimeout)
				continue
			}
			expireTimer.Reset(nextExpireTime)
		}
	}
}

func (exp *FlowExporter) sendFlowRecords() (time.Duration, error) {
	currTime := time.Now()
	var expireTime1, expireTime2 time.Duration
	exp.expiredConns, expireTime1 = exp.conntrackConnStore.GetExpiredConns(exp.expiredConns, currTime, maxConnsToExport)
	exp.expiredConns, expireTime2 = exp.denyConnStore.GetExpiredConns(exp.expiredConns, currTime, maxConnsToExport)
	// Select the shorter time out among two connection stores to do the next round of export.
	nextExpireTime := getMinTime(expireTime1, expireTime2)
	for i := range exp.expiredConns {
		// 여기서 ip nil이 뜨는데 뭔지 모르겠다 더 확인해봐야지
		if err := exp.exportConn(&exp.expiredConns[i]); err != nil {
			tk.LogIt(tk.LogError, err.Error(), "Error when sending expired flow record\n")
			return nextExpireTime, err
		}
	}
	// Clear expiredConns slice after exporting. Allocated memory is kept.
	exp.expiredConns = exp.expiredConns[:0]
	return nextExpireTime, nil
}

func (exp *FlowExporter) initFlowExporter() error {
	var err error
	if exp.exporterInput.CollectorProtocol == "tcp" {
		// TCP transport does not need any tempRefTimeout, so sending 0.
		// tempRefTimeout is the template refresh timeout, which specifies how often
		// the exporting process should send the template again.
		exp.exporterInput.TempRefTimeout = 0
	} else {
		// For UDP transport, hardcoding tempRefTimeout value as 1800s.
		exp.exporterInput.TempRefTimeout = 1800
	}
	expProcess, err := ipfix.NewIPFIXExportingProcess(exp.exporterInput)
	if err != nil {
		return fmt.Errorf("error when starting exporter: %v", err)
	}
	exp.process = expProcess
	if exp.v4Enabled {
		templateID := exp.process.NewTemplateID()
		exp.templateIDv4 = templateID
		sentBytes, err := exp.sendTemplateSet(false)
		if err != nil {
			return err
		}

		tk.LogIt(tk.LogInfo, "Initialized flow exporter for IPv4 flow records and sent %d bytes size of template record\n", sentBytes)
	}
	if exp.v6Enabled {
		templateID := exp.process.NewTemplateID()
		exp.templateIDv6 = templateID
		sentBytes, err := exp.sendTemplateSet(true)
		if err != nil {
			return err
		}
		tk.LogIt(tk.LogInfo, "Initialized flow exporter for IPv6 flow records and sent %d bytes size of template record\n", sentBytes)
	}
	metrics.ReconnectionsToFlowCollector.Inc()
	return nil
}

func (exp *FlowExporter) sendTemplateSet(isIPv6 bool) (int, error) {
	elements := make([]ipfixentities.InfoElementWithValue, 0)

	IANAInfoElements := IANAInfoElementsIPv4
	LoxiLBInfoElements := LoxiLBInfoElementsIPv4
	templateID := exp.templateIDv4
	if isIPv6 {
		IANAInfoElements = IANAInfoElementsIPv6
		LoxiLBInfoElements = LoxiLBInfoElementsIPv6
		templateID = exp.templateIDv6
	}
	for _, ie := range IANAInfoElements {
		element, err := exp.registry.GetInfoElement(ie, ipfixregistry.IANAEnterpriseID)
		if err != nil {
			return 0, fmt.Errorf("%s not present. returned error: %v", ie, err)
		}
		ieWithValue, err := ipfixentities.DecodeAndCreateInfoElementWithValue(element, nil)
		if err != nil {
			return 0, fmt.Errorf("error when creating information element: %v", err)
		}
		elements = append(elements, ieWithValue)
	}
	for _, ie := range IANAReverseInfoElements {
		element, err := exp.registry.GetInfoElement(ie, ipfixregistry.IANAReversedEnterpriseID)
		if err != nil {
			return 0, fmt.Errorf("%s not present. returned error: %v", ie, err)
		}
		ieWithValue, err := ipfixentities.DecodeAndCreateInfoElementWithValue(element, nil)
		if err != nil {
			return 0, fmt.Errorf("error when creating information element: %v", err)
		}
		elements = append(elements, ieWithValue)
	}
	for _, ie := range LoxiLBInfoElements {
		element, err := exp.registry.GetInfoElement(ie, ipfixregistry.AntreaEnterpriseID) // It sams as Antrea ID
		if err != nil {
			return 0, fmt.Errorf("information element %s is not present in LoxiLB registry", ie)
		}
		ieWithValue, err := ipfixentities.DecodeAndCreateInfoElementWithValue(element, nil)
		if err != nil {
			return 0, fmt.Errorf("error when creating information element: %v", err)
		}
		elements = append(elements, ieWithValue)
	}
	exp.ipfixSet.ResetSet()
	if err := exp.ipfixSet.PrepareSet(ipfixentities.Template, templateID); err != nil {
		return 0, err
	}
	err := exp.ipfixSet.AddRecord(elements, templateID)
	if err != nil {
		return 0, fmt.Errorf("error in adding record to template set: %v", err)
	}
	sentBytes, err := exp.process.SendSet(exp.ipfixSet)
	if err != nil {
		return 0, fmt.Errorf("error in IPFIX exporting process when sending template record: %v", err)
	}

	// Get all elements from template record.
	if !isIPv6 {
		exp.elementsListv4 = elements
	} else {
		exp.elementsListv6 = elements
	}

	return sentBytes, nil
}

func (exp *FlowExporter) addConnToSet(conn *flowexporter.Connection) error {
	exp.ipfixSet.ResetSet()

	eL := exp.elementsListv4
	templateID := exp.templateIDv4
	if conn.FlowKey.SourceAddress.To4() == nil {
		templateID = exp.templateIDv6
		eL = exp.elementsListv6
	}
	if err := exp.ipfixSet.PrepareSet(ipfixentities.Data, templateID); err != nil {
		return err
	}
	// Iterate over all infoElements in the list
	for i := range eL {
		ie := eL[i]
		switch ieName := ie.GetInfoElement().Name; ieName {
		case "flowStartSeconds":
			ie.SetUnsigned32Value(uint32(conn.StartTime.Unix()))
		case "flowEndSeconds":
			ie.SetUnsigned32Value(uint32(conn.StopTime.Unix()))
		case "flowEndReason":
			if flowexporter.IsConnectionDying(conn) {
				ie.SetUnsigned8Value(ipfixregistry.EndOfFlowReason)
			} else if conn.IsActive {
				ie.SetUnsigned8Value(ipfixregistry.ActiveTimeoutReason)
			} else {
				ie.SetUnsigned8Value(ipfixregistry.IdleTimeoutReason)
			}
		case "sourceIPv4Address":
			ie.SetIPAddressValue(conn.FlowKey.SourceAddress)
		case "destinationIPv4Address":
			ie.SetIPAddressValue(conn.FlowKey.DestinationAddress)
		case "sourceIPv6Address":
			ie.SetIPAddressValue(conn.FlowKey.SourceAddress)
		case "destinationIPv6Address":
			ie.SetIPAddressValue(conn.FlowKey.DestinationAddress)
		case "sourceTransportPort":
			ie.SetUnsigned16Value(conn.FlowKey.SourcePort)
		case "destinationTransportPort":
			ie.SetUnsigned16Value(conn.FlowKey.DestinationPort)
		case "protocolIdentifier":
			ie.SetUnsigned8Value(conn.FlowKey.Protocol)
		case "packetTotalCount":
			ie.SetUnsigned64Value(conn.OriginalPackets)
		case "octetTotalCount":
			ie.SetUnsigned64Value(conn.OriginalBytes)
		case "packetDeltaCount":
			deltaPkts := int64(conn.OriginalPackets) - int64(conn.PrevPackets)
			if deltaPkts < 0 {
				tk.LogIt(tk.LogInfo, "Packet delta count for connection should not be negative packet delta count %v\n", deltaPkts)
			}
			ie.SetUnsigned64Value(uint64(deltaPkts))
		case "octetDeltaCount":
			deltaBytes := int64(conn.OriginalBytes) - int64(conn.PrevBytes)
			if deltaBytes < 0 {
				tk.LogIt(tk.LogInfo, "Byte delta count for connection should not be negative byte delta count %v \n", deltaBytes)
			}
			ie.SetUnsigned64Value(uint64(deltaBytes))
		case "reversePacketTotalCount":
			ie.SetUnsigned64Value(conn.ReversePackets)
		case "reverseOctetTotalCount":
			ie.SetUnsigned64Value(conn.ReverseBytes)
		case "reversePacketDeltaCount":
			deltaPkts := int64(conn.ReversePackets) - int64(conn.PrevReversePackets)
			if deltaPkts < 0 {
				tk.LogIt(tk.LogInfo, "Packet delta count for connection should not be negative packet delta count %v\n", deltaPkts)
			}
			ie.SetUnsigned64Value(uint64(deltaPkts))
		case "reverseOctetDeltaCount":
			deltaBytes := int64(conn.ReverseBytes) - int64(conn.PrevReverseBytes)
			if deltaBytes < 0 {
				tk.LogIt(tk.LogInfo, "Byte delta count for connection should not be negative byte delta count %v\n", deltaBytes)
			}
			ie.SetUnsigned64Value(uint64(deltaBytes))
		// case "sourcePodNamespace":
		// 	ie.SetStringValue(conn.SourcePodNamespace)
		// case "sourcePodName":
		// 	ie.SetStringValue(conn.SourcePodName)
		// case "sourceNodeName":
		// 	// Add nodeName for only local pods whose pod names are resolved.
		// 	if conn.SourcePodName != "" {
		// 		ie.SetStringValue(exp.nodeName)
		// 	} else {
		// 		ie.SetStringValue("")
		// 	}
		// case "destinationPodNamespace":
		// 	ie.SetStringValue(conn.DestinationPodNamespace)
		// case "destinationPodName":
		// 	ie.SetStringValue(conn.DestinationPodName)
		// case "destinationNodeName":
		// 	// Add nodeName for only local pods whose pod names are resolved.
		// 	if conn.DestinationPodName != "" {
		// 		ie.SetStringValue(exp.nodeName)
		// 	} else {
		// 		ie.SetStringValue("")
		// 	}
		// case "destinationClusterIPv4":
		// 	if conn.DestinationServicePortName != "" {
		// 		ie.SetIPAddressValue(conn.DestinationServiceAddress)
		// 	} else {
		// 		// Sending dummy IP as IPFIX collector expects constant length of data for IP field.
		// 		// We should probably think of better approach as this involves customization of IPFIX collector to ignore
		// 		// this dummy IP address.
		// 		ie.SetIPAddressValue(net.IP{0, 0, 0, 0})
		// 	}
		// case "destinationClusterIPv6":
		// 	if conn.DestinationServicePortName != "" {
		// 		ie.SetIPAddressValue(conn.DestinationServiceAddress)
		// 	} else {
		// 		// Same as destinationClusterIPv4.
		// 		ie.SetIPAddressValue(net.ParseIP("::"))
		// 	}
		// case "destinationServicePort":
		// 	if conn.DestinationServicePortName != "" {
		// 		ie.SetUnsigned16Value(conn.DestinationServicePort)
		// 	} else {
		// 		ie.SetUnsigned16Value(uint16(0))
		// 	}
		// case "destinationServicePortName":
		// 	ie.SetStringValue(conn.DestinationServicePortName)
		// case "ingressNetworkPolicyName":
		// 	ie.SetStringValue(conn.IngressNetworkPolicyName)
		// case "ingressNetworkPolicyNamespace":
		// 	ie.SetStringValue(conn.IngressNetworkPolicyNamespace)
		// case "ingressNetworkPolicyType":
		// 	ie.SetUnsigned8Value(conn.IngressNetworkPolicyType)
		// case "ingressNetworkPolicyRuleName":
		// 	ie.SetStringValue(conn.IngressNetworkPolicyRuleName)
		// case "ingressNetworkPolicyRuleAction":
		// 	ie.SetUnsigned8Value(conn.IngressNetworkPolicyRuleAction)
		// case "egressNetworkPolicyName":
		// 	ie.SetStringValue(conn.EgressNetworkPolicyName)
		// case "egressNetworkPolicyNamespace":
		// 	ie.SetStringValue(conn.EgressNetworkPolicyNamespace)
		// case "egressNetworkPolicyType":
		// 	ie.SetUnsigned8Value(conn.EgressNetworkPolicyType)
		// case "egressNetworkPolicyRuleName":
		// 	ie.SetStringValue(conn.EgressNetworkPolicyRuleName)
		// case "egressNetworkPolicyRuleAction":
		// 	ie.SetUnsigned8Value(conn.EgressNetworkPolicyRuleAction)
		case "tcpState":
			ie.SetStringValue(conn.TCPState)
		case "flowType":
			ie.SetUnsigned8Value(exp.findFlowType(*conn))
		}
	}
	err := exp.ipfixSet.AddRecord(eL, templateID)
	if err != nil {
		return fmt.Errorf("error in adding record to data set: %v", err)
	}
	return nil
}

func (exp *FlowExporter) sendDataSet() (int, error) {
	fmt.Printf("exp.ipfixSet: %v\n", exp.ipfixSet)

	sentBytes, err := exp.process.SendSet(exp.ipfixSet)
	if err != nil {
		return 0, fmt.Errorf("error when sending data set: %v", err)
	}
	tk.LogIt(tk.LogInfo, "Data set sent successfully Bytes sent %v \n", sentBytes)

	return sentBytes, nil
}

func (exp *FlowExporter) findFlowType(conn flowexporter.Connection) uint8 {
	// TODO: support Pod-To-External flows in network policy only mode.
	// if exp.isNetworkPolicyOnly {
	// 	if conn.SourcePodName == "" || conn.DestinationPodName == "" {
	// 		return ipfixregistry.FlowTypeInterNode
	// 	}
	// 	return ipfixregistry.FlowTypeIntraNode
	// }

	// if exp.nodeRouteController == nil {
	// 	klog.Warningf("Can't find flowType without nodeRouteController")
	// 	return 0
	// }
	// if exp.nodeRouteController.IPInPodSubnets(conn.FlowKey.SourceAddress) {
	// 	if conn.Mark&openflow.ServiceCTMark.GetRange().ToNXRange().ToUint32Mask() == openflow.ServiceCTMark.GetValue() || exp.nodeRouteController.IPInPodSubnets(conn.FlowKey.DestinationAddress) {
	// 	if conn.SourcePodName == "" || conn.DestinationPodName == "" {
	// 		return ipfixregistry.FlowTypeInterNode
	// 	}
	// 	return ipfixregistry.FlowTypeIntraNode
	// 	}
	// 	return ipfixregistry.FlowTypeToExternal
	// }
	// We do not support External-To-Pod flows for now.
	//klog.Warningf("Source IP: %s doesn't exist in PodCIDRs", conn.FlowKey.SourceAddress.String())
	return 0
}

func (exp *FlowExporter) exportConn(conn *flowexporter.Connection) error {
	// TODO: more records per data set will be supported when go-ipfix supports size check when adding records
	if err := exp.addConnToSet(conn); err != nil {
		return err
	}
	if _, err := exp.sendDataSet(); err != nil {
		return err
	}
	exp.numDataSetsSent = exp.numDataSetsSent + 1
	tk.LogIt(tk.LogInfo, "Record for connection sent successfully flowKey %v connection %v\n", conn.FlowKey, conn)
	return nil
}

func getMinTime(t1, t2 time.Duration) time.Duration {
	if t1 <= t2 {
		return t1
	}
	return t2
}
