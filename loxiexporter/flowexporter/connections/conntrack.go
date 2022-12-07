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

package connections

import (
	"github.com/loxilb-io/loxilb/loxiexporter/flowexporter"
)

// InitializeConnTrackDumper initializes the ConnTrackDumper interface for different OS and datapath types.
func InitializeConnTrackDumper() ConnTrackDumper {
	connTrackDumper := NewConnTrackSystem()
	return connTrackDumper
}

func filterLoxiLBConns(conns []*flowexporter.Connection) []*flowexporter.Connection {
	filteredConns := conns[:0]
	for _, conn := range conns {
		filteredConns = append(filteredConns, conn)
	}
	return filteredConns
}
