//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package trace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel/propagation"
)

func TestHarborSpanNameFormatter(t *testing.T) {
	type args struct {
		operation string
		request   *http.Request
	}
	normalReq, err := http.NewRequest("GET", "http://10.192.168.1/api/v2.0/configuration", nil)
	if err != nil {
		t.Error(err)
	}
	normalReqWithHTTPS, err := http.NewRequest("GET", "https://10.192.168.1/api/v2.0/configuration", nil)
	if err != nil {
		t.Error(err)
	}
	patternReq := httptest.NewRequest(http.MethodGet, "/api/v2.0/projects/library", nil)
	patternReq.Pattern = "/api/v2.0/projects/{project_name}"
	methodPatternReq := httptest.NewRequest(http.MethodGet, "/api/v2.0/projects/library", nil)
	methodPatternReq.Pattern = "GET /api/v2.0/projects/{project_name}"
	cases := []struct {
		name string
		in   args
		want string
	}{
		{
			name: `normal`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", Method: http.MethodGet, URL: &url.URL{Scheme: "http", Path: "/api/v2.0/configuration"}}},
			want: "GET",
		},
		{
			name: `normal request `,
			in:   args{"sample", normalReq},
			want: "GET",
		},
		{
			name: `normal request with https `,
			in:   args{"sample", normalReqWithHTTPS},
			want: "GET",
		},
		{
			name: `post`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", Method: http.MethodPost, URL: &url.URL{Path: "/api/v2.0/projects"}}},
			want: "POST",
		},
		{
			name: `nil url`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", Method: http.MethodGet}},
			want: "GET",
		},
		{
			name: `pattern`,
			in:   args{"sample", patternReq},
			want: "GET /api/v2.0/projects/{project_name}",
		},
		{
			name: `pattern with method`,
			in:   args{"sample", methodPatternReq},
			want: "GET /api/v2.0/projects/{project_name}",
		},
		{
			name: `no method`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", URL: &url.URL{Path: "/api/v2.0/configuration"}}},
			want: "sample",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := HarborSpanNameFormatter(tt.in.operation, tt.in.request)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}

func TestExractTraceID(t *testing.T) {
	type args struct {
		headers        map[string]string
		ctxTraceparent string
		traceEnabled   bool
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
			want: "",
		},
		{
			name: "Traceparent Header, trace enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				traceEnabled: true,
			},
			want: "0af7651916cd43dd8448eb211c80319c",
		},
		{
			name: "Traceparent Header, trace not enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				traceEnabled: false,
			},
			want: "0af7651916cd43dd8448eb211c80319c",
		},
		{
			name: "Traceparent Header, invalid",
			args: args{
				headers: map[string]string{
					"traceparent": "INVALID",
				},
			},
			want: "",
		},
		{
			name: "Traceparent Context, trace enabled",
			args: args{
				headers:        map[string]string{},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
				traceEnabled:   true,
			},
			want: "80e1afed08e019fc1110464cfa66635c",
		},
		{
			name: "Traceparent Context, trace not enabled",
			args: args{
				headers:        map[string]string{},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
				traceEnabled:   false,
			},
			want: "",
		},
		{
			name: "Traceparent Context, invalid, trace enabled",
			args: args{
				headers:        map[string]string{},
				ctxTraceparent: "INVALID",
				traceEnabled:   true,
			},
			want: "",
		},
		{
			name: "Traceparent Context, invalid, trace not enabled",
			args: args{
				headers:        map[string]string{},
				ctxTraceparent: "INVALID",
				traceEnabled:   false,
			},
			want: "",
		},
		{
			name: "Traceparent Context+Header, trace enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
				traceEnabled:   true,
			},
			want: "80e1afed08e019fc1110464cfa66635c",
		},
		{
			name: "Traceparent Context+Header, trace not enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
				traceEnabled:   false,
			},
			want: "0af7651916cd43dd8448eb211c80319c",
		},
		{
			name: "Traceparent Context+Header, invalid #1, trace enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "INVALID",
				},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
				traceEnabled:   true,
			},
			want: "80e1afed08e019fc1110464cfa66635c",
		},
		{
			name: "Traceparent Context+Header, invalid #1, trace not enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "INVALID",
				},
				ctxTraceparent: "00-80e1afed08e019fc1110464cfa66635c-7a085853722dc6d2-01",
				traceEnabled:   false,
			},
			want: "",
		},
		{
			name: "Traceparent Context+Header, invalid #2, trace enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				ctxTraceparent: "INVALID",
				traceEnabled:   true,
			},
			want: "0af7651916cd43dd8448eb211c80319c",
		},
		{
			name: "Traceparent Context+Header, invalid #2, trace not enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				},
				ctxTraceparent: "INVALID",
				traceEnabled:   false,
			},
			want: "0af7651916cd43dd8448eb211c80319c",
		},
		{
			name: "Traceparent Context+Header, invalid #3, trace enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "INVALID",
				},
				ctxTraceparent: "INVALID",
				traceEnabled:   true,
			},
			want: "",
		},
		{
			name: "Traceparent Context+Header, invalid #3, trace not enabled",
			args: args{
				headers: map[string]string{
					"traceparent": "INVALID",
				},
				ctxTraceparent: "INVALID",
				traceEnabled:   false,
			},
			want: "",
		},
	}

	origEnabled := C.Enabled
	defer func() {
		C.Enabled = origEnabled
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			C.Enabled = tt.args.traceEnabled

			ctx := context.Background()
			if tt.args.ctxTraceparent != "" {
				var prop propagation.TraceContext
				ctx = prop.Extract(ctx, propagation.MapCarrier{"traceparent": tt.args.ctxTraceparent})
			}

			req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil).WithContext(ctx)
			for h, v := range tt.args.headers {
				req.Header.Set(h, v)
			}

			traceID := ExractTraceID(req)

			assert.Equal(t, tt.want, traceID, tt.name)
		})
	}
}
