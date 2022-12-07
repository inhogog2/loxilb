//go:build linux
// +build linux

// Copyright 2020 Antrea Authors
// Modified by 2022 NetLOX Inc
// Modified log : Added log as like LoxiLB.
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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/ti-mo/conntrack"

	tk "github.com/loxilb-io/loxilib"

	cmn "github.com/loxilb-io/loxilb/common"

	"github.com/loxilb-io/loxilb/loxiexporter/flowexporter"
)

// connTrackSystem implements ConnTrackDumper. This is for linux kernel datapath.
var _ ConnTrackDumper = new(connTrackSystem)
var ApiHooks cmn.NetHookInterface

const (
	sysctlNet = "/proc/sys/net"
)

// GetSysctlNet returns the value for sysctl net.* settings
func GetSysctlNet(sysctl string) (int, error) {
	data, err := ioutil.ReadFile(path.Join(sysctlNet, sysctl))
	if err != nil {
		return -1, err
	}
	val, err := strconv.Atoi(strings.Trim(string(data), " \n"))
	if err != nil {
		return -1, err
	}
	return val, nil
}

type connTrackSystem struct {
	connTrack LoxiLBConnTrack
}

// TODO: detect the endianness of the system when initializing conntrack dumper to handle situations on big-endian platforms.
// All connection labels are required to store in little endian format in conntrack dumper.
func NewConnTrackSystem() *connTrackSystem {
	return &connTrackSystem{
		LoxiLBConnTrack{},
	}
}

// DumpFlows opens netlink connection and dumps all the flows in LoxiLB ZoneID of conntrack table.
func (ct *connTrackSystem) DumpFlows() ([]*flowexporter.Connection, int, error) {
	conns, err := ct.connTrack.DumpFlowsInCtZone()
	if err != nil {
		return nil, 0, fmt.Errorf("error when dumping flows from conntrack: %v\n", err)
	}

	filteredConns := filterLoxiLBConns(conns)
	tk.LogIt(tk.LogInfo, "No. of flow exporter considered flows in LoxiLB zoneID: %d\n", len(filteredConns))

	return filteredConns, len(conns), nil
}

// LoxiLBConnTrack interface helps for testing the code that contains the third party library functions ("github.com/ti-mo/conntrack")
type loxiLBConnTrack interface {
	Dial() error
	DumpFlowsInCtZone() ([]*flowexporter.Connection, error)
}

type LoxiLBConnTrack struct {
	LBConn *conntrack.Conn
}

func (nfct *LoxiLBConnTrack) DumpFlowsInCtZone() ([]*flowexporter.Connection, error) {
	conns, err := ApiHooks.NetCtInfoGet()
	if err != nil {
		return nil, err
	}
	loxiLBConns := make([]*flowexporter.Connection, len(conns))
	for i := range conns {
		conn := conns[i]
		loxiLBConns[i] = LoxiLBFlowConnection(conn)
	}

	tk.LogIt(tk.LogInfo, "Finished dumping -- total no. of flows in conntrack: %d\n", len(loxiLBConns))

	return loxiLBConns, nil
}

func LoxiLBFlowConnection(conn cmn.CtInfo) *flowexporter.Connection {
	tuple := flowexporter.Tuple{
		SourceAddress:      conn.Sip,
		DestinationAddress: conn.Dip,
		Protocol:           uint8(1), //conn.Proto,
		SourcePort:         conn.Sport,
		DestinationPort:    conn.Dport,
	}

	// Tuple to Key as hash function.
	hash := md5.Sum([]byte(fmt.Sprintf("%s%s%s%d%d", conn.Sip, conn.Dip, conn.Proto, conn.Sport, conn.Dport)))
	tupleKey := hex.EncodeToString(hash[:])

	// Assign all the applicable fields
	newConn := flowexporter.Connection{
		ID:              tupleKey,
		FlowKey:         tuple,
		OriginalPackets: conn.Pkts,
		OriginalBytes:   conn.Bytes,
		TCPState:        "",
	}
	if conn.Proto == "tcp" {
		newConn.TCPState = conn.CState
	}

	return &newConn
}

func (ct *connTrackSystem) GetMaxConnections() (int, error) {
	maxConns, err := GetSysctlNet("netfilter/nf_conntrack_max")
	return maxConns, err
}

// reference: https://github.com/torvalds/linux/blob/master/net/LoxiLB/nf_conntrack_proto_tcp.c#L51-L62
func stateToString(state uint8) string {
	stateList := []string{
		"NONE",
		"SYN_SENT",
		"SYN_RECV",
		"ESTABLISHED",
		"FIN_WAIT",
		"CLOSE_WAIT",
		"LAST_ACK",
		"TIME_WAIT",
		"CLOSE",
		"SYN_SENT2",
	}
	if state > uint8(9) { // invalid state number
		return ""
	}
	return stateList[state]
}
