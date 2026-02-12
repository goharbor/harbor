// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RegistryRequestsTotal counts total number of registry pull/head requests
	// received by the proxy cache
	RegistryRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "harbor",
			Subsystem: "core",
			Name:      "registry_requests_total",
			Help:      "Total number of registry pull/head requests received by proxy cache",
		},
		[]string{"project", "repo", "method"},
	)

	// ProxyUpstreamRequestsTotal counts requests that were proxied to upstream
	// (cache miss)
	ProxyUpstreamRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "harbor",
			Subsystem: "core",
			Name:      "proxy_upstream_requests_total",
			Help:      "Number of requests proxied to upstream registry (cache miss)",
		},
		[]string{"project", "upstream_url", "status"},
	)
)
