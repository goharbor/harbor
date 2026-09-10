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

package bearer

import (
	"context"
	"fmt"
	libcache "github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/memory"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

func TestModify(t *testing.T) {
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiIsImtpZCI6IlBZWU86VEVXVTpWN0pIOjI2SlY6QVFUWjpMSkMzOlNYVko6WEdIQTozNEYyOjJMQVE6WlJNSzpaN1E2In0.eyJpc3MiOiJhdXRoLmRvY2tlci5jb20iLCJzdWIiOiJqbGhhd24iLCJhdWQiOiJyZWdpc3RyeS5kb2NrZXIuY29tIiwiZXhwIjoxNDE1Mzg3MzE1LCJuYmYiOjE0MTUzODcwMTUsImlhdCI6MTQxNTM4NzAxNSwianRpIjoidFlKQ08xYzZjbnl5N2tBbjBjN3JLUGdiVjFIMWJGd3MiLCJhY2Nlc3MiOlt7InR5cGUiOiJyZXBvc2l0b3J5IiwibmFtZSI6InNhbWFsYmEvbXktYXBwIiwiYWN0aW9ucyI6WyJwdXNoIl19XX0.QhflHPfbd6eVF4lM9bwYpFZIV0PfikbyXuLx959ykRTBpe3CYnzs6YBK8FToVb5R47920PVLrh8zuLzdCr9t3w"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "username" || p != "password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte(fmt.Sprintf(`{"token": "%s", "expires_in": 3600,"issued_at": "2009-11-10T23:00:00Z"}`, token)))
	}))
	defer server.Close()

	// invalid credential
	a := basic.NewAuthorizer("username", "invalid_password")
	authorizer := NewAuthorizer(server.URL, "service", a, commonhttp.NewTransport())
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	err := authorizer.Modify(req)
	require.NotNil(t, err)
	assert.True(t, errors.IsErr(err, errors.UnAuthorizedCode))

	// valid credential
	a = basic.NewAuthorizer("username", "password")
	authorizer = NewAuthorizer(server.URL, "service", a, commonhttp.NewTransport())
	req, _ = http.NewRequest(http.MethodGet, server.URL, nil)
	err = authorizer.Modify(req)
	require.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("Bearer %s", token), req.Header.Get("Authorization"))
}

// New clients must share tokens, but credentials and repository scopes must not.
func TestSharedTokenCache(t *testing.T) {
	require.NoError(t, libcache.Initialize(libcache.Memory, ""))
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		fmt.Fprintf(w, `{"access_token":"token-%d","expires_in":3600}`, requests)
	}))
	defer server.Close()
	newAuth := func(password string) *authorizer {
		return NewAuthorizer(server.URL, "service", basic.NewAuthorizer("user", password), http.DefaultTransport).(*authorizer)
	}
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/v2/repo/manifests/latest", nil)
	require.NoError(t, newAuth("secret").Modify(req))
	require.NoError(t, newAuth("secret").Modify(req))
	assert.Equal(t, 1, requests)
	assert.Equal(t, "Bearer token-1", req.Header.Get("Authorization"))

	require.NoError(t, newAuth("changed").Modify(req))
	assert.Equal(t, 2, requests)
	other, _ := http.NewRequest(http.MethodGet, server.URL+"/v2/other/manifests/latest", nil)
	require.NoError(t, newAuth("secret").Modify(other))
	assert.Equal(t, 3, requests)

	a := newAuth("secret")
	a.Invalidate(req)
	require.NoError(t, a.Modify(req))
	assert.Equal(t, 4, requests)
	tokenReq, err := a.tokenRequest(parseScopes(req))
	require.NoError(t, err)
	expired := &token{Token: "expired", ExpiresIn: 60, IssuedAt: "2000-01-01T00:00:00Z"}
	require.NoError(t, libcache.Default().Save(context.Background(), tokenCacheKey(tokenReq), expired, time.Hour))
	require.NoError(t, newAuth("secret").Modify(req))
	assert.Equal(t, 5, requests)
}

func TestRejectedTokenRefresh(t *testing.T) {
	require.NoError(t, libcache.Initialize(libcache.Memory, ""))
	tokens, pulls := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			tokens++
			fmt.Fprintf(w, `{"token":"token-%d","expires_in":3600}`, tokens)
			return
		}
		pulls++
		if r.Header.Get("Authorization") != "Bearer token-2" {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()
	a := NewAuthorizer(server.URL+"/token", "service", basic.NewAuthorizer("user", "secret"), http.DefaultTransport)
	client := commonhttp.NewClient(server.Client(), a)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/v2/repo/manifests/latest", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, tokens)
	assert.Equal(t, 2, pulls)
	// A permanently rejected token must only be retried once.
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/v2/other/manifests/latest", nil)
	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, 4, tokens)
	assert.Equal(t, 4, pulls)
}
