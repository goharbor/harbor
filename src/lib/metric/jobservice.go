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

// RegisterJobServiceCollectors ...
func RegisterJobServiceCollectors() {
	prometheus.MustRegister([]prometheus.Collector{
		JobserviceInfo,
		JobserviceTotalTask,
		JobservieTaskProcessTimeSummary,
	}...)
}

var (
	// JobserviceInfo used for collect jobservice information
	JobserviceInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: os.Getenv(NamespaceEnvKey),
			Subsystem: os.Getenv(SubsystemEnvKey),
			Name:      "info",
			Help:      "the information of jobservice",
		},
		[]string{"node", "pool", "workers"},
	)
	// JobserviceTotalTask used for collect data
	JobserviceTotalTask = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: os.Getenv(NamespaceEnvKey),
			Subsystem: os.Getenv(SubsystemEnvKey),
			Name:      "task_total",
			Help:      "The number of processed tasks",
		},
		[]string{"type", "status"},
	)
	// JobservieTaskProcessTimeSummary used for instrument task running time
	JobservieTaskProcessTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  os.Getenv(NamespaceEnvKey),
			Subsystem:  os.Getenv(SubsystemEnvKey),
			Name:       "task_process_time_seconds",
			Help:       "The time duration of the task processing time",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"type", "status"})
)
