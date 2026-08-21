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
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

// registryCollectorName ...
const registryCollectorName = "RegistryCollector"

var (
	registryInfo = typedDesc{
		desc:      newDescWithLabels("", "registry_info", "Information about the registry", "name", "description", "type", "url", "insecure", "status"),
		valueType: prometheus.GaugeValue,
	}
)

// NewRegistryCollector ...
func NewRegistryCollector() *RegistryCollector {
	return &RegistryCollector{
		regCtl: registry.Ctl,
	}
}

// RegistryCollector ...
type RegistryCollector struct {
	regCtl registry.Controller
}

// Describe implements prometheus.Collector
func (hrc RegistryCollector) Describe(c chan<- *prometheus.Desc) {
	c <- registryInfo.Desc()
}

// Collect implements prometheus.Collector
func (hrc RegistryCollector) Collect(c chan<- prometheus.Metric) {
	registryMetrics := hrc.getInformation()
	for _, metric := range registryMetrics {
		c <- metric
	}
}

// GetName returns the name of the registry collector
func (hrc RegistryCollector) GetName() string {
	return registryCollectorName
}

func (hrc RegistryCollector) getInformation() []prometheus.Metric {
	if CacheEnabled() {
		if cachedValue, ok := CacheGet(registryCollectorName); ok {
			return cachedValue.([]prometheus.Metric)
		}
	}

	var (
		result []prometheus.Metric
		ctx    = orm.Context()
	)

	regList, err := hrc.regCtl.List(ctx, q.New(q.KeyWords{}))
	if err != nil {
		log.Errorf("get public projects error: %v", err)
		return result
	}

	for _, r := range regList {
		// hrc.regCtl.IsHealthy(ctx, r)
		result = append(result, registryInfo.MustNewConstMetric(1,
			r.Name,
			r.Description,
			r.Type,
			r.URL,
			strconv.FormatBool(r.Insecure),
			r.Status,
		))
	}

	if CacheEnabled() {
		CachePut(registryCollectorName, result)
	}

	return result
}
