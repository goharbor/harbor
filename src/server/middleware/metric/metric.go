package metric

import (
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/metric"
)

func instrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now, url := time.Now(), r.URL.EscapedPath()

		metric.TotalInFlightGauge.WithLabelValues(url).Inc()
		defer metric.TotalInFlightGauge.WithLabelValues(url).Dec()
		rc := lib.NewResponseRecorder(w)

		next.ServeHTTP(rc, r)

		metric.TotalReqDurSummary.WithLabelValues(r.Method, url).Observe(time.Since(now).Seconds())
		metric.TotalReqCnt.WithLabelValues(r.Method, strconv.Itoa(rc.StatusCode), url).Inc()
	})
}

// Middleware returns a middleware for handling requests
func Middleware() func(http.Handler) http.Handler {
	if config.Metric().Enabled {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				next = instrumentHandler(next)
				next.ServeHTTP(rw, req)
			})
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(rw, req)
		})
	}
}
