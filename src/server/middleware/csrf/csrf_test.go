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

package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib"
)

// reached answers 200 and records that the request got past the middleware.
type reached struct {
	called bool
}

func (r *reached) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	r.called = true
	w.WriteHeader(http.StatusOK)
}

// serve runs one request through the middleware on the given host.
func serve(t *testing.T, method, host string, headers map[string]string, carrySession bool) (int, bool) {
	t.Helper()

	handler := &reached{}
	req := httptest.NewRequest(method, "https://"+host+"/c/login", nil)
	req.Host = host
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if carrySession {
		req = req.WithContext(lib.WithCarrySession(req.Context(), true))
	}

	rec := httptest.NewRecorder()
	Middleware()(handler).ServeHTTP(rec, req)

	return rec.Code, handler.called
}

const host = "harbor.example.com"

// TestMiddlewareAllows covers the traffic that must keep working. Note that no
// case names a configured endpoint: each request is judged on the origin the
// client actually used, which is what lets one Harbor answer on several
// ingresses, over either scheme, behind a proxy that rewrites Host.
func TestMiddlewareAllows(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		host    string
		headers map[string]string
	}{
		{
			name:    "same-origin write from a browser",
			method:  http.MethodPost,
			host:    host,
			headers: map[string]string{"Sec-Fetch-Site": "same-origin", "Origin": "https://" + host},
		},
		{
			name:    "top level navigation",
			method:  http.MethodPost,
			host:    host,
			headers: map[string]string{"Sec-Fetch-Site": "none"},
		},
		{
			name:    "a second ingress, unknown to any configuration",
			method:  http.MethodPost,
			host:    "eu." + host,
			headers: map[string]string{"Sec-Fetch-Site": "same-origin", "Origin": "https://eu." + host},
		},
		{
			name:    "browser predating Sec-Fetch-Site, Origin matches Host",
			method:  http.MethodPost,
			host:    host,
			headers: map[string]string{"Origin": "https://" + host},
		},
		{
			name:   "safe method needs no origin at all",
			method: http.MethodGet,
			host:   host,
		},
		{
			name:   "HEAD is safe too",
			method: http.MethodHead,
			host:   host,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, called := serve(t, c.method, c.host, c.headers, false)

			assert.Equal(t, http.StatusOK, code)
			assert.True(t, called, "request should have reached the handler")
		})
	}
}

// TestMiddlewareRejects covers the attacks. The same-site case is the one the
// previous token scheme let through: SameSite=Strict does not help, because a
// sibling subdomain is same-site and can write cookies on the parent domain.
func TestMiddlewareRejects(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
	}{
		{
			name:    "cross-site write",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "https://evil.example.com"},
		},
		{
			name:    "same-site subdomain, the CVE-2025-24358 case",
			headers: map[string]string{"Sec-Fetch-Site": "same-site", "Origin": "https://evil." + host},
		},
		{
			name:    "browser predating Sec-Fetch-Site, foreign Origin",
			headers: map[string]string{"Origin": "https://evil.example.com"},
		},
		{
			name:    "a forged token buys nothing, no token is consulted",
			headers: map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "https://evil.example.com", "X-Harbor-CSRF-Token": "anything"},
		},
		{
			// CrossOriginProtection admits this as non-browser traffic; the
			// middleware does not, since csrfSkipper already excused the routes
			// non-browser clients use.
			name:    "unsafe request that declares no origin at all",
			headers: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, called := serve(t, http.MethodPost, host, c.headers, false)

			assert.Equal(t, http.StatusForbidden, code)
			assert.False(t, called, "request should not have reached the handler")
		})
	}
}

// TestCsrfSkipper pins which routes the check applies to. Anything not carrying
// a session on these prefixes is a registry or API client rather than a browser.
func TestCsrfSkipper(t *testing.T) {
	cases := []struct {
		path         string
		carrySession bool
		skip         bool
	}{
		{"/v2/library/hello-world/manifests/latest", false, true},
		{"/api/v2.0/projects", false, true},
		{"/service/token", false, true},
		{"/v2/library/hello-world/manifests/latest", true, false},
		{"/api/v2.0/projects", true, false},
		{"/service/token", true, false},
		{"/c/login", false, false},
		{"/", false, false},
	}

	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "https://"+host+c.path, nil)
		if c.carrySession {
			req = req.WithContext(lib.WithCarrySession(req.Context(), true))
		}

		assert.Equal(t, c.skip, csrfSkipper(req), "path %q carrySession=%v", c.path, c.carrySession)
	}
}

// TestMiddlewareSkippedRoutesStayOpen checks the skipper and the check together:
// a registry client with no session must not be turned away.
func TestMiddlewareSkippedRoutesStayOpen(t *testing.T) {
	handler := &reached{}
	req := httptest.NewRequest(http.MethodPost, "https://"+host+"/api/v2.0/projects", nil)
	req.Host = host

	rec := httptest.NewRecorder()
	Middleware()(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handler.called)
}
