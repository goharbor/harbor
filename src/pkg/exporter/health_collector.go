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
	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/goharbor/harbor/src/lib/log"
)

const (
	healthURL           string = "/api/v2.0/health"
	healthCollectorName string = "HealthCollector"
)

var (
	harborHealth = typedDesc{
		desc:      newDesc("", "health", "Running status of Harbor"),
		valueType: prometheus.GaugeValue,
	}
	harborComponentsHealth = typedDesc{
		desc:      newDescWithLabels("", "up", "Running status of harbor component", "component"),
		valueType: prometheus.GaugeValue,
	}
)

// NewHealthCollect ...
func NewHealthCollect(cli *HarborClient) *HealthCollector {
	return &HealthCollector{
		HarborClient: cli,
	}
}

// HealthCollector is the Heartbeat
type HealthCollector struct {
	*HarborClient
}

// Describe implements prometheus.Collector
func (hc *HealthCollector) Describe(c chan<- *prometheus.Desc) {
	c <- harborHealth.Desc()
	c <- harborComponentsHealth.Desc()
}

// Collect implements prometheus.Collector
func (hc *HealthCollector) Collect(c chan<- prometheus.Metric) {
	for _, m := range hc.getHealthStatus() {
		c <- m
	}
}

// GetName returns the name of the health collector
func (hc *HealthCollector) GetName() string {
	return healthCollectorName
}

func (hc *HealthCollector) getHealthStatus() []prometheus.Metric {
	if CacheEnabled() {
		value, ok := CacheGet(healthCollectorName)
		if ok {
			return value.([]prometheus.Metric)
		}
	}
	result := []prometheus.Metric{}
	res, err := hbrCli.Get(healthURL)
	if err != nil {
		log.Errorf("request health info failed with err: %v", err)
		return result
	}
	defer res.Body.Close()
	var healthResponse responseHealth
	err = json.NewDecoder(res.Body).Decode(&healthResponse)
	if err != nil {
		log.Errorf("failed to decode res.Body into healthResponse, error: %v", err)
		return result
	}
	result = append(result, harborHealth.MustNewConstMetric(healthy(healthResponse.Status)))
	for _, v := range healthResponse.Components {
		result = append(result, harborComponentsHealth.MustNewConstMetric(healthy(v.Status), v.Name))
	}
	if CacheEnabled() {
		CachePut(healthCollectorName, result)
	}
	return result
}

type responseHealth struct {
	Status     string              `json:"status"`
	Components []responseComponent `json:"components"`
}

type responseComponent struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func healthy(h string) float64 {
	if h == "healthy" {
		return 1
	}
	return 0
}
