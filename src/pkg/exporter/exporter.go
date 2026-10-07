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
	CacheDuration          int64
	CacheCleanInterval     int64
}

// NewExporter creates a exporter for Harbor with the configuration
func NewExporter(opt *Opt) *Exporter {
	exporter := &Exporter{
		Opt:        opt,
		collectors: make(map[string]prometheus.Collector),
	}
	if opt.CacheDuration > 0 {
		CacheInit(opt)
	}
	err := exporter.RegisterCollector(NewHealthCollect(hbrCli),
		NewSystemInfoCollector(hbrCli),
		NewRegistryCollector(),
		NewProjectCollector(),
		NewJobServiceCollector(),
		NewStatisticsCollector(),
	)
	if err != nil {
		log.Warningf("calling RegisterCollector() errored out, error: %v", err)
	}

	r := prometheus.NewRegistry()
	r.MustRegister(exporter)
	exporter.Server = newServer(opt, r)

	return exporter
}

// Exporter is struct for Harbor which can used to connection Harbor and collecting data
type Exporter struct {
	*http.Server
	Opt        *Opt
	collectors map[string]prometheus.Collector
}

// RegisterCollector register a collector to expoter
func (e *Exporter) RegisterCollector(collectors ...collector) error {
	for _, c := range collectors {
		name := c.GetName()
		if _, ok := e.collectors[name]; ok {
			return errors.New("collector name is already registered")
		}
		e.collectors[name] = c
		log.Infof("collector %s registered ...", name)
	}
	return nil
}

func newServer(opt *Opt, _ *prometheus.Registry) *http.Server {
	exporterMux := http.NewServeMux()
	exporterMux.Handle(opt.MetricsPath, promhttp.Handler())
	exporterMux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html>
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
	for _, v := range e.collectors {
		v.Describe(c)
	}
}

// Collect implements prometheus.Collector
func (e *Exporter) Collect(c chan<- prometheus.Metric) {
	for _, v := range e.collectors {
		v.Collect(c)
	}
}

func checkErr(err error, arg string) {
	if err == nil {
		return
	}

	log.Errorf("%s: %v", arg, err)
}
