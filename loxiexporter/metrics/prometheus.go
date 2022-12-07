// Copyright 2020 Antrea Authors
// Modified by 2022 NetLOX Inc
// Modified log : Remove K8s integration parts. And added log as like LoxiLB.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	tk "github.com/loxilb-io/loxilib"
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	metricNamespaceLoxiLB = "loxilb"
	metricSubsystemAgent  = "agent"
)

var (
	EgressNetworkPolicyRuleCount = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "egress_networkpolicy_rule_count",
			Help:           "Number of egress NetworkPolicy rules on local Node which are managed by the loxilb Agent.",
			StabilityLevel: metrics.STABLE,
		},
	)

	IngressNetworkPolicyRuleCount = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "ingress_networkpolicy_rule_count",
			Help:           "Number of ingress NetworkPolicy rules on local Node which are managed by the loxilb Agent.",
			StabilityLevel: metrics.STABLE,
		},
	)

	PodCount = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "local_pod_count",
			Help:           "Number of Pods on local Node which are managed by the loxilb Agent.",
			StabilityLevel: metrics.STABLE,
		},
	)

	NetworkPolicyCount = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "networkpolicy_count",
			Help:           "Number of NetworkPolicies on local Node which are managed by the loxilb Agent.",
			StabilityLevel: metrics.STABLE,
		},
	)

	TotalConnectionsInConnTrackTable = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "conntrack_total_connection_count",
			Help:           "Number of connections in the conntrack table. This metric gets updated at an interval specified by flowPollInterval, a configuration parameter for the Agent.",
			StabilityLevel: metrics.ALPHA,
		},
	)

	TotalloxilbConnectionsInConnTrackTable = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "conntrack_loxilb_connection_count",
			Help:           "Number of connections in the loxilb ZoneID of the conntrack table. This metric gets updated at an interval specified by flowPollInterval, a configuration parameter for the Agent.",
			StabilityLevel: metrics.ALPHA,
		},
	)

	TotalDenyConnections = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "denied_connection_count",
			Help:           "Number of denied connections detected by Flow Exporter deny connections tracking. This metric gets updated when a flow is rejected/dropped by network policy.",
			StabilityLevel: metrics.ALPHA,
		},
	)

	ReconnectionsToFlowCollector = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "flow_collector_reconnection_count",
			Help:           "Number of re-connections between Flow Exporter and flow collector. This metric gets updated whenever the connection is re-established between the Flow Exporter and the flow collector (e.g. the Flow Aggregator).",
			StabilityLevel: metrics.ALPHA,
		},
	)

	MaxConnectionsInConnTrackTable = metrics.NewGauge(
		&metrics.GaugeOpts{
			Namespace:      metricNamespaceLoxiLB,
			Subsystem:      metricSubsystemAgent,
			Name:           "conntrack_max_connection_count",
			Help:           "Size of the conntrack table. This metric gets updated at an interval specified by flowPollInterval, a configuration parameter for the Agent.",
			StabilityLevel: metrics.ALPHA,
		},
	)
)

func InitializePrometheusMetrics() {
	tk.LogIt(tk.LogInfo, "Initializing prometheus metrics\n")

	InitializePodMetrics()
	InitializeNetworkPolicyMetrics()
	InitializeConnectionMetrics()
}

func InitializePodMetrics() {
	if err := legacyregistry.Register(PodCount); err != nil {
		tk.LogIt(tk.LogError, err.Error(), "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_local_pod_count\n")
	}
}

func InitializeNetworkPolicyMetrics() {
	if err := legacyregistry.Register(EgressNetworkPolicyRuleCount); err != nil {
		tk.LogIt(tk.LogError, err.Error(), err, "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_egress_networkpolicy_rule_count\n")
	}

	if err := legacyregistry.Register(IngressNetworkPolicyRuleCount); err != nil {
		tk.LogIt(tk.LogError, err.Error(), err, "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_ingress_networkpolicy_rule_count\n")
	}

	if err := legacyregistry.Register(NetworkPolicyCount); err != nil {
		tk.LogIt(tk.LogError, err.Error(), err, "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_networkpolicy_count\n")
	}
}

func InitializeConnectionMetrics() {
	if err := legacyregistry.Register(TotalConnectionsInConnTrackTable); err != nil {
		tk.LogIt(tk.LogError, err.Error(), err, "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_conntrack_total_connection_count\n")
	}
	if err := legacyregistry.Register(TotalloxilbConnectionsInConnTrackTable); err != nil {
		tk.LogIt(tk.LogError, err.Error(), err, "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_conntrack_loxilb_connection_count\n")
	}
	if err := legacyregistry.Register(TotalDenyConnections); err != nil {
		tk.LogIt(tk.LogError, err.Error(), err, "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_denied_connection_count\n")
	}
	if err := legacyregistry.Register(ReconnectionsToFlowCollector); err != nil {
		tk.LogIt(tk.LogError, err.Error(), err, "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_flow_collector_reconnection_count\n")
	}
	if err := legacyregistry.Register(MaxConnectionsInConnTrackTable); err != nil {
		tk.LogIt(tk.LogError, err.Error(), err, "Failed to register metrics with Prometheus", "metrics", "loxilb_agent_conntrack_max_connection_count\n")
	}
}
