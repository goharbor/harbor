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
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestNewRegistryCollector(t *testing.T) {
	rc := NewRegistryCollector()
	assert.NotNil(t, rc)
	assert.NotNil(t, rc.regCtl)
}

func TestRegistryCollector_GetName(t *testing.T) {
	rc := NewRegistryCollector()
	assert.Equal(t, RegistryCollectorName, rc.GetName())
}

func TestRegistryCollector_Describe(t *testing.T) {
	rc := NewRegistryCollector()
	descChan := make(chan *prometheus.Desc, 1)

	go rc.Describe(descChan)
	desc := <-descChan

	assert.NotNil(t, desc)
	assert.Equal(t, registryStatus.Desc(), desc)
}

func TestRegistryCollector_Collect_WithCache(t *testing.T) {
	CacheInit(&Opt{
		CacheDuration: 60,
	})

	// Create test metrics and put them in cache
	data := []prometheus.Metric{
		prometheus.MustNewConstMetric(registryStatus.Desc(), prometheus.GaugeValue, 1, "docker-hub", "https://hub.docker.com", "docker-hub"),
		prometheus.MustNewConstMetric(registryStatus.Desc(), prometheus.GaugeValue, 0, "quay", "https://quay.io", "quay"),
	}
	CachePut(RegistryCollectorName, data)

	rc := NewRegistryCollector()
	metricChan := make(chan prometheus.Metric, 2)

	go rc.Collect(metricChan)

	metric1 := <-metricChan
	if !reflect.DeepEqual(metric1, data[0]) {
		t.Errorf("RegistryCollector.Collect() first metric = %v, want %v", metric1, data[0])
	}

	metric2 := <-metricChan
	if !reflect.DeepEqual(metric2, data[1]) {
		t.Errorf("RegistryCollector.Collect() second metric = %v, want %v", metric2, data[1])
	}
}

func TestGetHealthyValue(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   float64
	}{
		{
			name:   "healthy status",
			status: model.Healthy,
			want:   1,
		},
		{
			name:   "unhealthy status",
			status: model.Unhealthy,
			want:   0,
		},
		{
			name:   "unknown status",
			status: "unknown",
			want:   0,
		},
		{
			name:   "empty status",
			status: "",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getHealthyValue(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistryStatusMetric(t *testing.T) {
	// Create a metric
	metric := registryStatus.MustNewConstMetric(1, "test-registry", "https://test.registry.io", "docker-hub")

	// Write it to a DTO and verify
	dto := &dto.Metric{}
	err := metric.Write(dto)
	assert.NoError(t, err)
	assert.NotNil(t, dto.Gauge)
	assert.NotNil(t, dto.Gauge.Value)
	assert.Equal(t, float64(1), *dto.Gauge.Value)

	// Verify labels
	labels := dto.GetLabel()
	assert.Len(t, labels, 3)

	labelMap := make(map[string]string)
	for _, l := range labels {
		labelMap[l.GetName()] = l.GetValue()
	}
	assert.Equal(t, "test-registry", labelMap["name"])
	assert.Equal(t, "https://test.registry.io", labelMap["url"])
	assert.Equal(t, "docker-hub", labelMap["type"])
}
