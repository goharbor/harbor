package exporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/goharbor/harbor/src/lib/log"
)

// Opt is the config of Harbor exporter
type Opt struct {
	Port                   int
	MetricsPath            string
	ExporterMetricsEnabled bool
	MaxRequests            int
	TLSEnabled             bool
	Certificate            string
	Key                    string
}

// NewExporter creates a exporter for Harbor with the configuration
func NewExporter(opt *Opt) *Exporter {
	exporter := &Exporter{
		Opt:        opt,
		Collectors: make(map[string]prometheus.Collector),
	}
	exporter.RegisterColletor(healthCollectorName, NewHealthCollect(hbrCli))
	exporter.RegisterColletor(systemInfoCollectorName, NewSystemInfoCollector(hbrCli))
	exporter.RegisterColletor(ProjectCollectorName, NewProjectCollector())
	r := prometheus.NewRegistry()
	r.MustRegister(exporter)
	exporter.Server = newServer(opt, r)

	return exporter
}

// Exporter is struct for Harbor which can used to connection Harbor and collecting data
type Exporter struct {
	*http.Server

	Opt *Opt
	ctx context.Context

	Collectors map[string]prometheus.Collector
}

// RegisterColletor register a collector to expoter
func (e *Exporter) RegisterColletor(name string, c prometheus.Collector) error {
	if _, ok := e.Collectors[name]; ok {
		return errors.New("Collector name is already registered")
	}
	e.Collectors[name] = c
	log.Infof("collector %s registered ...", name)
	return nil
}

func newServer(opt *Opt, r *prometheus.Registry) *http.Server {
	exporterMux := http.NewServeMux()
	exporterMux.Handle(opt.MetricsPath, promhttp.Handler())
	exporterMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>Harbor Exporter</title></head>
		<body>
		<h1>Harbor Exporter</h1>
		<p><a href="` + opt.MetricsPath + `">Metrics</a></p>
		</body>
		</html>`))
	})

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", opt.Port),
		Handler: exporterMux,
	}
}

// Describe implements prometheus.Collector
func (e *Exporter) Describe(c chan<- *prometheus.Desc) {
	for _, v := range e.Collectors {
		v.Describe(c)
	}
}

// Collect implements prometheus.Collector
func (e *Exporter) Collect(c chan<- prometheus.Metric) {
	for _, v := range e.Collectors {
		v.Collect(c)
	}
}
