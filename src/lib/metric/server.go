package metric

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
