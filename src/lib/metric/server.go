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
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// NamespaceEnvKey is the metric namespace key in environment
	NamespaceEnvKey = "METRIC_NAMESPACE"
	// SubsystemEnvKey is the metric subsystem key in environment
	SubsystemEnvKey = "METRIC_SUBSYSTEM"
)

// ServeProm return a server to serve prometheus metrics
func ServeProm(path string, port int) {
	mux := http.NewServeMux()
	mux.Handle(path, promhttp.Handler())
	log.Infof("Prometheus metric server running on port %v", port)
	log.Errorf("Promethus metrcis server down with %s", http.ListenAndServe(fmt.Sprintf(":%v", port), mux))
}
