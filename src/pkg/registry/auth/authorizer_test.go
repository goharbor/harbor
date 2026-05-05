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

package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize_NoPathPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	a := &authorizer{client: server.Client()}
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/v2/library/nginx/manifests/latest", nil)
	err := a.initialize(req.URL)
	require.NoError(t, err)
	assert.Equal(t, "/v2/", a.url.Path)
}

func TestInitialize_WithPathPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prefix/v2/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	a := &authorizer{client: server.Client()}
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/prefix/v2/library/nginx/manifests/latest", nil)
	err := a.initialize(req.URL)
	require.NoError(t, err)
	assert.Equal(t, "/prefix/v2/", a.url.Path)
}

func TestInitialize_WithDeepPathPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deep/nested/prefix/v2/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	a := &authorizer{client: server.Client()}
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/deep/nested/prefix/v2/library/nginx/manifests/latest", nil)
	err := a.initialize(req.URL)
	require.NoError(t, err)
	assert.Equal(t, "/deep/nested/prefix/v2/", a.url.Path)
}

func TestInitialize_BearerWithPathPrefix(t *testing.T) {
	token := "test-token-value"
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`{"token": "%s", "expires_in": 3600, "issued_at": "2009-11-10T23:00:00Z"}`, token)))
	}))
	defer tokenServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prefix/v2/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Www-Authenticate", fmt.Sprintf(`Bearer realm="%s",service="test-service"`, tokenServer.URL))
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	a := &authorizer{
		username: "user",
		password: "pass",
		client:   server.Client(),
	}
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/prefix/v2/library/nginx/manifests/latest", nil)
	err := a.initialize(req.URL)
	require.NoError(t, err)
	assert.Equal(t, "/prefix/v2/", a.url.Path)
	assert.NotNil(t, a.authorizer)
}

func TestIsTarget_NoPathPrefix(t *testing.T) {
	a := &authorizer{}
	a.url, _ = parseURL("http://registry.example.com/v2/")

	req, _ := http.NewRequest(http.MethodGet, "http://registry.example.com/v2/library/nginx/manifests/latest", nil)
	assert.True(t, a.isTarget(req))

	req, _ = http.NewRequest(http.MethodGet, "http://s3.amazonaws.com/bucket/layer.tar.gz", nil)
	assert.False(t, a.isTarget(req))
}

func TestIsTarget_WithPathPrefix(t *testing.T) {
	a := &authorizer{}
	a.url, _ = parseURL("http://registry.example.com/prefix/v2/")

	req, _ := http.NewRequest(http.MethodGet, "http://registry.example.com/prefix/v2/library/nginx/manifests/latest", nil)
	assert.True(t, a.isTarget(req))

	req, _ = http.NewRequest(http.MethodGet, "http://registry.example.com/v2/library/nginx/manifests/latest", nil)
	assert.False(t, a.isTarget(req))

	req, _ = http.NewRequest(http.MethodGet, "http://s3.amazonaws.com/bucket/layer.tar.gz", nil)
	assert.False(t, a.isTarget(req))
}

func parseURL(raw string) (*url.URL, error) {
	return url.Parse(raw)
}
