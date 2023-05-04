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

package metric

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
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
	// ReferrersOperationID ...
	ReferrersOperationID = "v2_referrers"
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

func instrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric.TotalInFlightGauge.Inc()
		defer metric.TotalInFlightGauge.Dec()
		now, rc, op := time.Now(), lib.NewResponseRecorder(w), ""
		ctx := context.WithValue(r.Context(), contextOpIDKey{}, &op)
		next.ServeHTTP(rc, r.WithContext(ctx))
		if len(op) == 0 {
			// From swagger's perspective the operation of this legacy URL is unknown
			op = "unknown"
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
