package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "harbor"
)

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

type collector interface {
	prometheus.Collector
	// Return the name of the collector
	GetName() string
}
