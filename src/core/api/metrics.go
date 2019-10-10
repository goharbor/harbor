package api

import (
	"contrib.go.opencensus.io/exporter/prometheus"
	"fmt"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"github.com/goharbor/harbor/src/common/utils/log"
)

func InitMetrics() *prometheus.Exporter{
	log.Info("InitMetrics")
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "demo",
	})
	if err != nil {
		log.Errorf("Failed to set prometheus NewExporter %v", err)
	}
	log.Info("Register Metrics")
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Errorf("Failed to register server views for HTTP metrics: %v", err)
	}
	return pe
}
