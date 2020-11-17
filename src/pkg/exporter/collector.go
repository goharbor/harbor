package exporter

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "harbor"
	subsystem = "exporter"
)

var (
	scrapeDuration = typedDesc{
		desc:      newDescWithLables(subsystem, "collector_duration_seconds", "Duration of a collector scrape", "collector"),
		valueType: prometheus.GaugeValue,
	}
	scrapeSuccess = typedDesc{
		desc:      newDescWithLables(subsystem, "collector_success", " Whether a collector succeeded.", "collector"),
		valueType: prometheus.GaugeValue,
	}
)

var hbrCli *HarborClient

// HarborClient is client for request harbor
type HarborClient struct {
	HarborScheme string
	HarborHost   string
	HarborPort   int
	*http.Client
}

func (hc HarborClient) harborURL(p string) url.URL {
	return url.URL{
		Scheme: hc.HarborScheme,
		Host:   fmt.Sprintf("%s:%d", hc.HarborHost, hc.HarborPort),
		Path:   p,
	}
}

// Get ...
func (hc HarborClient) Get(p string) (*http.Response, error) {
	hbrURL := hc.harborURL(p)
	res, err := http.Get(hbrURL.String())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// InitHarborClient initialize the harbor client
func InitHarborClient(hc *HarborClient) {
	hbrCli = hc
}

func newDesc(subsystem, name, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, nil, nil,
	)
}

func newDescWithLables(subsystem, name, help string, labels ...string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, name),
		help, labels, nil,
	)
}

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (d *typedDesc) MustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

func (d *typedDesc) Desc() *prometheus.Desc {
	return d.desc
}

// // ErrNoData indicates the collector found no data to collect, but had no other error.
// var ErrNoData = errors.New("collector returned no data")

// func IsNoDataError(err error) bool {
// 	return err == ErrNoData
// }
