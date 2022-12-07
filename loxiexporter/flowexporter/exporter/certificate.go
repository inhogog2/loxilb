// Copyright 2021 Antrea Authors
// Modified by 2022 NetLOX Inc
// Modified log : Change Name
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

const (
	CAConfigMapName       = "flow-aggregator-ca"
	CAConfigMapKey        = "ca.crt"
	CAConfigMapNamespace  = "flow-aggregator"
	ClientSecretNamespace = "flow-aggregator"
	// #nosec G101: false positive triggered by variable name which includes "Secret"
	ClientSecretName = "flow-aggregator-client-tls"
)
