package metric

import (
	"context"
	"github.com/goharbor/harbor/src/lib/config"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/metric"
)

// ContextOpIDKey ...
type contextOpIDKey struct{}

const (
	// CatalogOperationID ...
	CatalogOperationID = "v2_catalog"
	// ListTagOperationID ...
	ListTagOperationID = "v2_tags"
	// ManifestOperationID ...
	ManifestOperationID = "v2_manifest"
	// BlobsOperationID ...
	BlobsOperationID = "v2_blob"
	// BlobsUploadOperationID ...
	BlobsUploadOperationID = "v2_blob_upload"
	// OthersOperationID ...
	OthersOperationID = "v2_others"
)

// SetMetricOpID used to set operation ID for metrics
func SetMetricOpID(ctx context.Context, value string) {
	if config.Metric().Enabled {
		v := ctx.Value(contextOpIDKey{}).(*string)
		*v = value
	}
}

func isChartMuseumURL(url string) bool {
	return strings.HasPrefix(url, "/chartrepo/") || strings.HasPrefix(url, "/api/chartrepo/")
}

func instrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric.TotalInFlightGauge.Inc()
		defer metric.TotalInFlightGauge.Dec()
		now, rc, op := time.Now(), lib.NewResponseRecorder(w), ""
		ctx := context.WithValue(r.Context(), contextOpIDKey{}, &op)
		next.ServeHTTP(rc, r.WithContext(ctx))
		if len(op) == 0 {
			if isChartMuseumURL(r.URL.Path) {
				op = "chartmuseum"
			} else {
				// From swagger's perspective the operation of this legacy URL is unknown
				op = "unknown"
			}
		}
		metric.TotalReqDurSummary.WithLabelValues(r.Method, op).Observe(time.Since(now).Seconds())
		metric.TotalReqCnt.WithLabelValues(r.Method, strconv.Itoa(rc.StatusCode), op).Inc()
	})
}

func transparentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Middleware returns a middleware for handling requests
func Middleware() func(http.Handler) http.Handler {
	if config.Metric().Enabled {
		return instrumentHandler
	}
	return transparentHandler
}

// InjectOpIDMiddleware returns a middleware used for injecting operations ID
func InjectOpIDMiddleware(opID string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			SetMetricOpID(r.Context(), opID)
			next.ServeHTTP(w, r)
		})
	}
}
