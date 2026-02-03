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

	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const RegistryCollectorName = "RegistryCollector"

var (
	registryStatus = typedDesc{
		desc:      newDescWithLables("", "registry_status", "Status of the registry", "name", "url", "type"),
		valueType: prometheus.GaugeValue,
	}
)

type RegistryCollector struct {
	regCtl registry.Controller
}

func NewRegistryCollector() *RegistryCollector {
	return &RegistryCollector{
		regCtl: registry.Ctl,
	}
}

func (rc *RegistryCollector) GetName() string {
	return RegistryCollectorName
}

func (rc *RegistryCollector) Describe(c chan<- *prometheus.Desc) {
	c <- registryStatus.Desc()
}

// Collect implements prometheus.Collector
func (rc *RegistryCollector) Collect(c chan<- prometheus.Metric) {
	for _, m := range rc.getRegistryStatus() {
		c <- m
	}
}

func (rc *RegistryCollector) getRegistryStatus() []prometheus.Metric {
	if CacheEnabled() {
		value, ok := CacheGet(RegistryCollectorName)
		if ok {
			return value.([]prometheus.Metric)
		}
	}

	result := []prometheus.Metric{}
	ctx := orm.Context()

	registries, err := rc.regCtl.List(ctx, nil)
	if err != nil {
		log.Errorf("failed to list registries: %v", err)
		return result
	}

	for _, reg := range registries {
		status := getHealthyValue(reg.Status)
		result = append(result, registryStatus.MustNewConstMetric(status, reg.Name, reg.URL, reg.Type))
	}

	if CacheEnabled() {
		CachePut(RegistryCollectorName, result)
	}
	return result
}

// Returns 1 for healthy, 0 for unhealthy or unknown status
func getHealthyValue(status string) float64 {
	if status == model.Healthy {
		return 1
	}
	return 0
}
