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
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/propagation"

	"github.com/goharbor/harbor/src/lib/log"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
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
	loc := "/server/middleware/log/log_test.go:41"
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
