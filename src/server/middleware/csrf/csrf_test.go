package csrf

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

func TestMain(m *testing.M) {
	conf := map[string]any{}
	config.InitWithSettings(conf)
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

type handler struct {
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestMiddleware(t *testing.T) {
	cases := []struct {
		req        *http.Request
		statusCode int
	}{
		{
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			statusCode: http.StatusOK,
		},
		{
			// neither Sec-Fetch-Site nor Origin is present, so the request is either
			// same-origin or not a browser request at all, and is allowed by design
			req:        httptest.NewRequest(http.MethodDelete, "/", nil),
			statusCode: http.StatusOK,
		},
		{
			// httptest.NewRequest sets the Host to "example.com", so the Origin matches it
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/", nil)
				req.Header.Set("Origin", "http://example.com")
				return req
			}(),
			statusCode: http.StatusOK,
		},
		{
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/", nil)
				req.Header.Set("Origin", "http://attacker.example")
				return req
			}(),
			statusCode: http.StatusForbidden,
		},
		{
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/", nil)
				req.Header.Set("Sec-Fetch-Site", "same-origin")
				return req
			}(),
			statusCode: http.StatusOK,
		},
		{
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/", nil)
				req.Header.Set("Sec-Fetch-Site", "none")
				return req
			}(),
			statusCode: http.StatusOK,
		},
		{
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/", nil)
				req.Header.Set("Sec-Fetch-Site", "cross-site")
				return req
			}(),
			statusCode: http.StatusForbidden,
		},
		{
			// Sec-Fetch-Site takes precedence over the Origin fallback, so a matching
			// Origin does not rescue a request the browser reported as cross-site
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/", nil)
				req.Header.Set("Sec-Fetch-Site", "cross-site")
				req.Header.Set("Origin", "http://example.com")
				return req
			}(),
			statusCode: http.StatusForbidden,
		},
		{
			req:        httptest.NewRequest(http.MethodGet, "/api/2.0/projects", nil), // should be skipped
			statusCode: http.StatusOK,
		},
		{
			req:        httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/latest", nil), // should be skipped
			statusCode: http.StatusOK,
		},
	}
	for _, c := range cases {
		srv := Middleware()(&handler{})
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, c.req)
		assert.Equal(t, c.statusCode, rec.Result().StatusCode)
	}
}

func TestMiddlewareCrossOriginError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "http://attacker.example")
	rec := httptest.NewRecorder()

	Middleware()(&handler{}).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.NotEmpty(t, rec.Body.String())
}

func TestCSRFSkipper(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		carrySession bool
		want         bool
	}{
		{name: "registry path", path: "/v2/library/alpine/manifests/latest", want: true},
		{name: "api path", path: "/api/2.0/projects", want: true},
		{name: "service path", path: "/service/token", want: true},
		{name: "registry root is protected", path: "/v2", want: false},
		{name: "path not matching /api/ prefix is protected", path: "/apis/projects", want: false},
		{name: "service root is protected", path: "/service", want: false},
		{name: "session registry request is protected", path: "/v2/library/alpine/manifests/latest", carrySession: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req = req.WithContext(lib.WithCarrySession(context.Background(), tt.carrySession))
			assert.Equal(t, tt.want, csrfSkipper(req))
		})
	}
}
