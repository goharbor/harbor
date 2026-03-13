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

package exporter

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/goharbor/harbor/src/pkg/metrics"
)

const (
	proxyCollectorName string = "ProxyCollector"
)

// NewProxyCollector creates a new proxy cache metrics collector
func NewProxyCollector() *ProxyCollector {
	return &ProxyCollector{}
}

// ProxyCollector collects proxy cache metrics
type ProxyCollector struct{}

// Describe implements prometheus.Collector
func (pc *ProxyCollector) Describe(c chan<- *prometheus.Desc) {
	metrics.RegistryRequestsTotal.Describe(c)
	metrics.ProxyUpstreamRequestsTotal.Describe(c)
}

// Collect implements prometheus.Collector
func (pc *ProxyCollector) Collect(c chan<- prometheus.Metric) {
	metrics.RegistryRequestsTotal.Collect(c)
	metrics.ProxyUpstreamRequestsTotal.Collect(c)
}

// GetName returns the name of the proxy collector
func (pc *ProxyCollector) GetName() string {
	return proxyCollectorName
}
