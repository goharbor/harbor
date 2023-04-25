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

package metric

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

// RegisterCollectors register all the common static collector
func RegisterCollectors() {
	prometheus.MustRegister([]prometheus.Collector{
		TotalInFlightGauge,
		TotalReqCnt,
		TotalReqDurSummary,
	}...)
}

var (
	// TotalInFlightGauge used to collect total in flight number
	TotalInFlightGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: os.Getenv(NamespaceEnvKey),
			Subsystem: os.Getenv(SubsystemEnvKey),
			Name:      "http_inflight_requests",
			Help:      "The total number of requests",
		},
	)

	// TotalReqCnt used to collect total request counter
	TotalReqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: os.Getenv(NamespaceEnvKey),
			Subsystem: os.Getenv(SubsystemEnvKey),
			Name:      "http_request_total",
			Help:      "The total number of requests",
		},
		[]string{"method", "code", "operation"},
	)

	// TotalReqDurSummary used to collect total request duration summaries
	TotalReqDurSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  os.Getenv(NamespaceEnvKey),
			Subsystem:  os.Getenv(SubsystemEnvKey),
			Name:       "http_request_duration_seconds",
			Help:       "The time duration of the requests",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method", "operation"})
)
