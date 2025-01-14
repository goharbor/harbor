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

package log

import (
	"io"
	"net/http"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/server/middleware"
)

// Middleware middleware which add logger to context
func Middleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		rid := r.Header.Get("X-Request-ID")
		if rid != "" {
			logger := log.G(r.Context())
			logger.Debugf("attach request id %s to the logger for the request %s %s", rid, r.Method, r.URL.Path)

			ctx := log.WithLogger(r.Context(), logger.WithFields(log.Fields{"requestID": rid}))
			r = r.WithContext(ctx)
		}

		traceID := tracelib.ExractTraceID(r)
		if traceID != "" {
			ctx := log.WithLogger(r.Context(), log.G(r.Context()).WithFields(log.Fields{"traceID": traceID}))
			r = r.WithContext(ctx)
		}

		e := &commonevent.Metadata{
			Ctx:           r.Context(),
			Username:      "unknown",
			RequestMethod: r.Method,
			RequestURL:    r.URL.String(),
		}
		if matched, resName := e.PreCheckMetadata(); matched {
			lib.NopCloseRequest(r)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			requestContent := string(body)
			if secCtx, ok := security.FromContext(r.Context()); ok {
				e.Username = secCtx.GetUsername()
			}
			rw := &ResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(rw, r)

			// Add information in the response
			e.ResourceName = resName
			e.RequestPayload = requestContent
			e.ResponseCode = rw.statusCode

			// Need to parse the Location header to get the resource ID on creating resource
			if e.RequestMethod == http.MethodPost {
				e.ResponseLocation = rw.header.Get("Location")
			}

			notification.AddEvent(e.Ctx, e, true)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// ResponseWriter wrapper to HTTP response to get the statusCode and response content
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	header     http.Header
}

// WriteHeader write header info
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Header get header info
func (rw *ResponseWriter) Header() http.Header {
	rw.header = rw.ResponseWriter.Header()
	return rw.ResponseWriter.Header()
}
