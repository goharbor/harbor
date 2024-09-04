package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/suite"
)

type StatisticsCollectorTestSuite struct {
	suite.Suite
	collector *StatisticsCollector
}

func (c *StatisticsCollectorTestSuite) TestStatisticsCollector() {
	metrics := c.collector.getStatistics()
	c.Equalf(7, len(metrics), "statistics collector should return %d metrics", 7)
	c.testGaugeMetric(metrics[0], 2, "total repo amount mismatch")     // total repo amount
	c.testGaugeMetric(metrics[1], 1, "public repo amount mismatch")    // only one project is public so its single repo is public too
	c.testGaugeMetric(metrics[2], 1, "primate repo amount mismatch")   //
	c.testGaugeMetric(metrics[3], 3, "total project amount mismatch")  // including library, project by default
	c.testGaugeMetric(metrics[4], 2, "public project amount mismatch") // including library, project by default
	c.testGaugeMetric(metrics[5], 1, "private project amount mismatch")
	c.testGaugeMetric(metrics[6], 0, "total storage usage mismatch") // still zero
}

func (c *StatisticsCollectorTestSuite) getMetricDTO(m prometheus.Metric) *dto.Metric {
	d := &dto.Metric{}
	c.NoError(m.Write(d))
	return d
}

func (c *StatisticsCollectorTestSuite) testCounterMetric(m prometheus.Metric, value float64) {
	d := c.getMetricDTO(m)
	if !c.NotNilf(d, "write metric error") {
		return
	}
	if !c.NotNilf(d.Counter, "counter is nil") {
		return
	}
	if !c.NotNilf(d.Counter.Value, "counter value is nil") {
		return
	}
	c.Equalf(value, *d.Counter.Value, "expected counter value does not match: expected: %v actual: %v", value, *d.Counter.Value)
}

func (c *StatisticsCollectorTestSuite) testGaugeMetric(m prometheus.Metric, value float64, msg string) {
	d := c.getMetricDTO(m)
	if !c.NotNilf(d, "write metric error") {
		return
	}
	if !c.NotNilf(d.Gauge, "gauge is nil") {
		return
	}
	if !c.NotNilf(d.Gauge.Value, "gauge value is nil") {
		return
	}
	c.Equalf(value, *d.Gauge.Value, "%s expected: %v actual: %v", msg, value, *d.Gauge.Value)
}
