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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	_ "github.com/goharbor/harbor/src/pkg/auditext/event/login"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/propagation"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/lib/log"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
	_ "github.com/goharbor/harbor/src/pkg/auditext/event/config"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

type MiddlewareTestSuite struct {
	suite.Suite
}

func (s *MiddlewareTestSuite) TestTableMiddleware() {
	next := func(fields log.Fields) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.G(r.Context()).WithFields(fields).Info("this is message") // variable loc below refers to this line

			w.WriteHeader(http.StatusOK)
		})
	}
	loc := "/server/middleware/log/log_test.go:47"
	locPrefix := regexp.MustCompile(fmt.Sprintf(`\[([^\s]*)%s\]`, loc))

	type args struct {
		headers        map[string]string
		fields         map[string]any
		ctxTraceparent string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Dummy",
			args: args{
				headers: map[string]string{},
			},
			want: fmt.Sprintf("TIMESTAMP [INFO] [%s]%s", loc,
				"[TestCase=\"Dummy\"]: this is message\n"),
		},
		{
			name: "X-Request-ID",
			args: args{
				headers: map[string]string{
					"X-Request-ID": "fd6139e6-9092-4181-9220-42d3d48bf658",
				},
			},
			want: fmt.Sprintf("TIMESTAMP [INFO] [%s]%s", loc,
				"[TestCase=\"X-Request-ID\" requestID=\"fd6139e6-9092-4181-9220-42d3d48bf658\"]: this is message\n"),
		},
		{
			name: "X-Request-ID, field",
			args: args{
				headers: map[string]string{
					"X-Request-ID": "fd6139e6-9092-4181-9220-42d3d48bf658",
				},
				fields: log.Fields{"method": "GET"},
			},
			want: fmt.Sprintf("TIMESTAMP [INFO] [%s]%s", loc,
				"[TestCase=\"X-Request-ID, field\" method=\"GET\" requestID=\"fd6139e6-9092-4181-9220-42d3d48bf658\"]: this is message\n"),
		},
		{
			name: "Traceparent Header",
			args: args{
				headers: map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
			},
			want: fmt.Sprintf("TIMESTAMP [INFO] [%s]%s", loc,
				"[TestCase=\"Traceparent Header\" traceID=\"0af7651916cd43dd8448eb211c80319c\"]: this is message\n"),
		},
		{
			name: "Traceparent Context",
			args: args{
				headers:        map[string]string{},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
			},
			want: fmt.Sprintf("TIMESTAMP [INFO] [%s]%s", loc,
				"[TestCase=\"Traceparent Context\" traceID=\"80e1afed08e019fc1110464cfa66635c\"]: this is message\n"),
		},
		{
			name: "Traceparent Context+Header",
			args: args{
				headers: map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
			},
			want: fmt.Sprintf("TIMESTAMP [INFO] [%s]%s", loc,
				"[TestCase=\"Traceparent Context+Header\" traceID=\"80e1afed08e019fc1110464cfa66635c\"]: this is message\n"),
		},
		{
			name: "Traceparent Context+Header, X-Request-ID",
			args: args{
				headers: map[string]string{
					"X-Request-ID": "fd6139e6-9092-4181-9220-42d3d48bf658",
					"traceparent":  "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
			},
			want: fmt.Sprintf("TIMESTAMP [INFO] [%s]%s", loc,
				"[TestCase=\"Traceparent Context+Header, X-Request-ID\" requestID=\"fd6139e6-9092-4181-9220-42d3d48bf658\" traceID=\"80e1afed08e019fc1110464cfa66635c\"]: this is message\n"),
		},
		{
			name: "Traceparent Context+Header, X-Request-ID, field",
			args: args{
				headers: map[string]string{
					"X-Request-ID": "fd6139e6-9092-4181-9220-42d3d48bf658",
					"traceparent":  "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
				fields:         log.Fields{"method": "GET"},
			},
			want: fmt.Sprintf("TIMESTAMP [INFO] [%s]%s", loc,
				"[TestCase=\"Traceparent Context+Header, X-Request-ID, field\" method=\"GET\" requestID=\"fd6139e6-9092-4181-9220-42d3d48bf658\" traceID=\"80e1afed08e019fc1110464cfa66635c\"]: this is message\n"),
		},
	}

	origEnabled := tracelib.C.Enabled
	defer func() {
		tracelib.C.Enabled = origEnabled
	}()
	tracelib.C.Enabled = true

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			b := make([]byte, 0, 200)
			buf := bytes.NewBuffer(b)
			formatter := log.NewTextFormatter()
			formatter.SetTimeFormat("TIMESTAMP")
			logger := log.New(buf, formatter, log.InfoLevel, 3).WithField("TestCase", tt.name)
			ctx := log.WithLogger(context.Background(), logger)
			if tt.args.ctxTraceparent != "" {
				var prop propagation.TraceContext
				ctx = prop.Extract(ctx, propagation.MapCarrier{"traceparent": tt.args.ctxTraceparent})
			}

			req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil).WithContext(ctx)
			for h, v := range tt.args.headers {
				req.Header.Set(h, v)
			}
			rr := httptest.NewRecorder()

			Middleware()(next(tt.args.fields)).ServeHTTP(rr, req)

			line := string(removeSubmatch(locPrefix, buf.Bytes()))
			s.Equal(tt.want, line, tt.name)
		})
	}
}

func (s *MiddlewareTestSuite) TestRequestEntityTooLarge() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	largeBody := make([]byte, common.MaxAuditLogPayloadSize+1)
	req := httptest.NewRequest("POST", "/c/login", bytes.NewReader(largeBody))
	rr := httptest.NewRecorder()

	Middleware()(next).ServeHTTP(rr, req)

	s.Equal(http.StatusRequestEntityTooLarge, rr.Code)
	s.Equal("request body too large\n", rr.Body.String())
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &MiddlewareTestSuite{})
}

func removeSubmatch(matchRe *regexp.Regexp, line []byte) []byte {
	matches := matchRe.FindSubmatchIndex(line)
	if len(matches) < 4 {
		return line
	}

	return append(line[0:matches[2]], line[matches[3]:]...)
}

func TestRemoveSubmatch(t *testing.T) {
	loc := "/server/middleware/log/log_test.go:41"
	locPrefix := regexp.MustCompile(fmt.Sprintf(`\[([^\s]*)%s\]`, loc))

	line := `TIMESTAMP [INFO] [/github.com/goharbor/harbor/src/server/middleware/log/log_test.go:41][method="GET" requestID="fd6139e6-9092-4181-9220-42d3d48bf658" traceID="80e1afed08e019fc1110464cfa66635c"]: this is message`
	assert.Equal(t, `TIMESTAMP [INFO] [/server/middleware/log/log_test.go:41][method="GET" requestID="fd6139e6-9092-4181-9220-42d3d48bf658" traceID="80e1afed08e019fc1110464cfa66635c"]: this is message`,
		string(removeSubmatch(locPrefix, []byte(line))),
	)

	line = `TIMESTAMP [INFO] [/server/middleware/log/log_test.go:41][method="GET" requestID="fd6139e6-9092-4181-9220-42d3d48bf658" traceID="80e1afed08e019fc1110464cfa66635c"]: this is message`
	assert.Equal(t, `TIMESTAMP [INFO] [/server/middleware/log/log_test.go:41][method="GET" requestID="fd6139e6-9092-4181-9220-42d3d48bf658" traceID="80e1afed08e019fc1110464cfa66635c"]: this is message`,
		string(removeSubmatch(locPrefix, []byte(line))),
	)

}

const maxAuditBodySize = common.MaxAuditLogPayloadSize

type observedReadCloser struct {
	remaining int64
	bytesRead int64
}

func (r *observedReadCloser) Read(buffer []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	count := int64(len(buffer))
	if count > r.remaining {
		count = r.remaining
	}
	for i := int64(0); i < count; i++ {
		buffer[i] = 'x'
	}
	r.remaining -= count
	r.bytesRead += count
	return int(count), nil
}

func (*observedReadCloser) Close() error {
	return nil
}

func TestMiddlewareBoundsActualAuditedBody(t *testing.T) {
	for _, unknownLength := range []bool{false, true} {
		name := "known length"
		if unknownLength {
			name = "unknown length"
		}
		t.Run(name, func(t *testing.T) {
			body := &observedReadCloser{remaining: 2 * maxAuditBodySize}
			request := httptest.NewRequest(
				http.MethodPut,
				"/api/v2.0/configurations?source=test",
				body,
			)
			// httptest.NewRequest only infers ContentLength for in-memory readers
			request.ContentLength = 2 * maxAuditBodySize
			if unknownLength {
				request.ContentLength = -1
			}

			downstreamCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				downstreamCalled = true
				w.WriteHeader(http.StatusNoContent)
			})
			response := httptest.NewRecorder()

			Middleware()(next).ServeHTTP(response, request)

			if response.Code != http.StatusRequestEntityTooLarge {
				t.Errorf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
			}
			if downstreamCalled {
				t.Error("oversized audited body reached downstream")
			}
			if body.bytesRead > maxAuditBodySize+1 {
				t.Errorf("audit middleware read %d bytes, want at most %d", body.bytesRead, maxAuditBodySize+1)
			}
			if body.bytesRead <= maxAuditBodySize {
				t.Errorf("audit middleware read %d bytes, want enough to detect oversize", body.bytesRead)
			}
		})
	}
}

func TestMiddlewarePreservesAuditedBodyAtLimit(t *testing.T) {
	body := &observedReadCloser{remaining: maxAuditBodySize}
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v2.0/configurations?source=test",
		body,
	)

	downstreamCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		downstreamCalled = true
		read, err := io.Copy(io.Discard, request.Body)
		if err != nil {
			t.Errorf("read downstream body: %v", err)
		}
		if read != maxAuditBodySize {
			t.Errorf("downstream body size = %d, want %d", read, maxAuditBodySize)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	response := httptest.NewRecorder()

	Middleware()(next).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if !downstreamCalled {
		t.Error("audited body at the limit did not reach downstream")
	}
}

type disabledAuditResolver struct{}

func (disabledAuditResolver) Resolve(*commonevent.Metadata, *event.Event) error {
	return nil
}

func (disabledAuditResolver) PreCheck(context.Context, string, string) (bool, string) {
	return false, ""
}

func TestMiddlewareDoesNotReadDisabledAuditBody(t *testing.T) {
	commonevent.RegisterResolver(`^/__disabled_audit_test__$`, disabledAuditResolver{})
	defer commonevent.UnregisterResolver(`^/__disabled_audit_test__$`)
	body := &observedReadCloser{remaining: 2 * maxAuditBodySize}
	request := httptest.NewRequest(http.MethodPut, "/__disabled_audit_test__", body)

	next := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Body != body {
			t.Error("disabled audit event replaced the request body")
		}
		if body.bytesRead != 0 {
			t.Errorf("disabled audit event read %d bytes before downstream", body.bytesRead)
		}
		_, _ = io.Copy(io.Discard, request.Body)
		w.WriteHeader(http.StatusNoContent)
	})
	response := httptest.NewRecorder()

	Middleware()(next).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if body.bytesRead != 2*maxAuditBodySize {
		t.Errorf("downstream read %d bytes, want %d", body.bytesRead, 2*maxAuditBodySize)
	}
}

func TestMiddlewareDoesNotReadRegistryResolverCollisions(t *testing.T) {
	urls := []string{
		"/v2/project/repo/manifests/latest?x=/api/v2.0/configurations",
		"/v2/api/v2.0/configurations/blobs/uploads/session?digest=sha256:" +
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"/prefix/api/v2.0/configurations",
		"/api/v2X0/configurations",
	}

	for _, requestURL := range urls {
		t.Run(requestURL, func(t *testing.T) {
			body := &observedReadCloser{remaining: 2 * maxAuditBodySize}
			request := httptest.NewRequest(http.MethodPut, requestURL, body)
			downstreamCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				downstreamCalled = true
				if request.Body != body {
					t.Error("audit middleware replaced a registry request body")
				}
				if body.bytesRead != 0 {
					t.Errorf("audit middleware read %d registry body bytes", body.bytesRead)
				}
				read, err := io.Copy(io.Discard, request.Body)
				if err != nil {
					t.Errorf("read registry body: %v", err)
				}
				if read != 2*maxAuditBodySize {
					t.Errorf("registry body size = %d, want %d", read, 2*maxAuditBodySize)
				}
				w.WriteHeader(http.StatusNoContent)
			})
			response := httptest.NewRecorder()

			Middleware()(next).ServeHTTP(response, request)

			if response.Code != http.StatusNoContent {
				t.Errorf("status = %d, want %d", response.Code, http.StatusNoContent)
			}
			if !downstreamCalled {
				t.Error("registry request did not reach downstream")
			}
		})
	}
}
