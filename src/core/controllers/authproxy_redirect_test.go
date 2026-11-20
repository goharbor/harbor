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

package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/authentication/v1beta1"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	ormmiddleware "github.com/goharbor/harbor/src/server/middleware/orm"
)

func init() {
	web.Router(common.AuthProxyRedirectPath, &AuthProxyController{}, "get:HandleRedirect")
}

// authProxyReviewSrv returns a token review server that authenticates only
// the given token, mirroring the request the authproxy flow issues to the
// configured token review endpoint.
func authProxyReviewSrv(t *testing.T, validToken string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/test/tokenreview", func(rw http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		review := &v1beta1.TokenReview{}
		if err := json.Unmarshal(body, review); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		if review.Spec.Token != validToken {
			_, _ = rw.Write([]byte(`{"apiVersion":"authentication.k8s.io/v1beta1","kind":"TokenReview","status":{"authenticated":false}}`))
			return
		}
		_, _ = rw.Write([]byte(strings.ReplaceAll(
			`{"apiVersion":"authentication.k8s.io/v1beta1","kind":"TokenReview","status":{"authenticated":true,"user":{"username":"USERNAME","groups":[]}}}`,
			"USERNAME", "jt")))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// authProxyConf returns the in-memory config enabling the http_auth auth
// mode against the given token review server.
func authProxyConf(reviewEndpoint string) map[string]any {
	return map[string]any{
		common.AUTHMode:                         common.HTTPAuth,
		common.HTTPAuthProxyEndpoint:            reviewEndpoint + "/test/login",
		common.HTTPAuthProxyTokenReviewEndpoint: reviewEndpoint + "/test/tokenreview",
		common.HTTPAuthProxyVerifyCert:          false,
	}
}

// serveAuthProxyRedirect drives the authproxy redirect flow with the given
// config and query parameters and returns the response recorder.
func serveAuthProxyRedirect(t *testing.T, conf map[string]any, query map[string]string) *httptest.ResponseRecorder {
	config.InitWithSettings(conf)
	t.Cleanup(func() {
		config.InitWithSettings(conf)
	})

	r, err := http.NewRequest(http.MethodGet, common.AuthProxyRedirectPath, nil)
	assert.NoError(t, err)
	q := r.URL.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	r.URL.RawQuery = q.Encode()

	w := httptest.NewRecorder()
	// Wrap the beego handler with the ORM middleware, mirroring the production
	// filter chain, so PostAuthenticate can onboard the reviewed user.
	handler := ormmiddleware.Middleware()(http.Handler(web.BeeApp.Handlers))
	handler.ServeHTTP(w, r)
	return w
}

// TestAuthProxyRedirectPostURILocalPathOnly ensures the post_uri parameter of
// the authproxy redirect flow only allows local paths, aligning the flow with
// the redirect_url validation applied by the OIDC login flow. An absolute or
// protocol-relative post_uri must fall back to "/" instead of redirecting the
// authenticated user's browser to an attacker-controlled site.
func TestAuthProxyRedirectPostURILocalPathOnly(t *testing.T) {
	srv := authProxyReviewSrv(t, "jt-token")
	conf := authProxyConf(srv.URL)

	testCases := []struct {
		name           string
		postURI        string
		expectLocation string
	}{
		{
			name:           "empty post_uri defaults to root",
			postURI:        "",
			expectLocation: "/",
		},
		{
			name:           "local path is kept",
			postURI:        "/harbor/projects",
			expectLocation: "/harbor/projects",
		},
		{
			name:           "absolute url to other site falls back to root",
			postURI:        "https://evil.example.com/phish",
			expectLocation: "/",
		},
		{
			name:           "protocol relative url falls back to root",
			postURI:        "//evil.example.com/phish",
			expectLocation: "/",
		},
		{
			name:           "host without scheme falls back to root",
			postURI:        "evil.example.com/phish",
			expectLocation: "/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			w := serveAuthProxyRedirect(t, conf, map[string]string{
				"token":   "jt-token",
				"postURI": tc.postURI,
			})
			assert.Equal(http.StatusMovedPermanently, w.Code, "authproxy redirect should always respond 301")
			assert.Equal(tc.expectLocation, w.Header().Get("Location"),
				"post_uri %q should redirect to %q", tc.postURI, tc.expectLocation)
		})
	}
}

// TestAuthProxyRedirectInvalidToken ensures a rejected token falls back to the
// root path without honoring the post_uri parameter.
func TestAuthProxyRedirectInvalidToken(t *testing.T) {
	assert := assert.New(t)

	srv := authProxyReviewSrv(t, "jt-token")
	conf := authProxyConf(srv.URL)

	w := serveAuthProxyRedirect(t, conf, map[string]string{
		"token":   "invalid-token",
		"postURI": "https://evil.example.com/phish",
	})
	assert.Equal(http.StatusMovedPermanently, w.Code)
	assert.Equal("/", w.Header().Get("Location"))
}
