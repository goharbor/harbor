package utils

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusObserveTimeMilliseconds observes the elapsed time since start, in milliseconds,
// on the specified Prometheus Histogram.
func PrometheusObserveTimeMilliseconds(h prometheus.Histogram, start time.Time) {
	h.Observe(float64(time.Since(start).Nanoseconds()) / float64(time.Millisecond))
}
