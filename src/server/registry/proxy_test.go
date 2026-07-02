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

package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func direct(req *http.Request) {
	req.Header.Add("test-key", "test-value")
}

func TestAuthDirectorBasicAuth(t *testing.T) {
	// Mock registry that responds with Basic Www-Authenticate
	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Www-Authenticate", `Basic realm="Registry Realm"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer registryServer.Close()

	// Reset probe state for each test
	detectedAuthType = atomic.Value{}

	// Override probe client to use the test server
	originalClient := probeHTTPClient
	probeHTTPClient = registryServer.Client()
	defer func() { probeHTTPClient = originalClient }()

	// Override registry URL via env (probe uses config.RegistryURL which reads env)
	t.Setenv("REGISTRY_URL", registryServer.URL)
	t.Setenv("REGISTRY_CREDENTIAL_USERNAME", "testuser")
	t.Setenv("REGISTRY_CREDENTIAL_PASSWORD", "testpassword")

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/v2/test/repo/manifests/latest", nil)
	d := authDirector(direct)
	d(req)
	assert.Equal(t, "test-value", req.Header.Get("test-key"))
	user, pass, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "testuser", user)
	assert.Equal(t, "testpassword", pass)
}

func TestAuthDirectorBearerToken(t *testing.T) {
	// Mock token service
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("service") != "harbor-registry" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "testuser" || pass != "testpassword" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": "test.jwt.token"})
	}))
	defer tokenServer.Close()

	// Mock registry that responds with Bearer Www-Authenticate
	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Www-Authenticate", `Bearer realm="http://example.com/service/token",service="harbor-registry"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer registryServer.Close()

	detectedAuthType = atomic.Value{} // reset to empty (not yet probed)
	originalClient := probeHTTPClient
	probeHTTPClient = registryServer.Client()
	defer func() { probeHTTPClient = originalClient }()

	originalTokenURL := getTokenServiceURL
	getTokenServiceURL = func() string {
		return tokenServer.URL
	}
	defer func() { getTokenServiceURL = originalTokenURL }()

	t.Setenv("REGISTRY_URL", registryServer.URL)
	t.Setenv("REGISTRY_CREDENTIAL_USERNAME", "testuser")
	t.Setenv("REGISTRY_CREDENTIAL_PASSWORD", "testpassword")

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/v2/test/repo/manifests/latest", nil)
	d := authDirector(direct)
	d(req)

	// Should have Bearer token, no basic auth
	_, _, hasBasicAuth := req.BasicAuth()
	assert.False(t, hasBasicAuth, "should not have basic auth when using bearer token")
	assert.Equal(t, "Bearer test.jwt.token", req.Header.Get("Authorization"))
}

func TestAuthDirectorProbeFailureDefaultsToBasic(t *testing.T) {
	detectedAuthType = atomic.Value{} // reset to empty (not yet probed)

	// Registry URL that will fail to connect
	t.Setenv("REGISTRY_URL", "http://127.0.0.1:1")
	t.Setenv("REGISTRY_CREDENTIAL_USERNAME", "fallbackuser")
	t.Setenv("REGISTRY_CREDENTIAL_PASSWORD", "fallbackpass")

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/v2/test/repo/manifests/latest", nil)
	d := authDirector(direct)
	d(req)

	// Should fall back to basic auth when probe fails
	user, pass, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "fallbackuser", user)
	assert.Equal(t, "fallbackpass", pass)
}

func TestAuthDirectorCachesProbeResult(t *testing.T) {
	probeCount := 0

	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probeCount++
		w.Header().Set("Www-Authenticate", `Basic realm="Registry Realm"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer registryServer.Close()

	detectedAuthType = atomic.Value{} // reset to empty (not yet probed)
	originalClient := probeHTTPClient
	probeHTTPClient = registryServer.Client()
	defer func() { probeHTTPClient = originalClient }()

	t.Setenv("REGISTRY_URL", registryServer.URL)
	t.Setenv("REGISTRY_CREDENTIAL_USERNAME", "u")
	t.Setenv("REGISTRY_CREDENTIAL_PASSWORD", "p")

	// First request triggers probe
	req1, _ := http.NewRequest(http.MethodGet, "http://example.com/v2/repo", nil)
	authDirector(direct)(req1)
	assert.Equal(t, 1, probeCount, "probe should run on first request")

	// Second request should use cached result, no additional probe
	req2, _ := http.NewRequest(http.MethodGet, "http://example.com/v2/repo", nil)
	authDirector(direct)(req2)
	assert.Equal(t, 1, probeCount, "probe should NOT run again (cached)")
}

func TestScopeFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "nil request",
			path:     "",
			expected: "repository:*:pull,push",
		},
		{
			name:     "full manifest path",
			path:     "/v2/library/nginx/manifests/latest",
			expected: "repository:library/nginx:pull,push",
		},
		{
			name:     "tags list",
			path:     "/v2/library/nginx/tags/list",
			expected: "repository:library/nginx:pull,push",
		},
		{
			name:     "blobs",
			path:     "/v2/library/nginx/blobs/sha256:abc123",
			expected: "repository:library/nginx:pull,push",
		},
		{
			name:     "single component path",
			path:     "/v2/repo",
			expected: "repository:*:pull,push",
		},
		{
			name:     "non-v2 path",
			path:     "/api/v2.0/projects",
			expected: "repository:*:pull,push",
		},
		{
			name:     "multi-level repository",
			path:     "/v2/registry.example.com/port/agent/blobs/sha256:def456",
			expected: "repository:registry.example.com/port:pull,push",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.path != "" {
				req = &http.Request{URL: &url.URL{Path: tt.path}}
			}
			assert.Equal(t, tt.expected, scopeFromRequest(req))
		})
	}
}

func TestGetRegistryToken(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "harbor-registry", r.URL.Query().Get("service"))
		assert.Contains(t, r.URL.Query().Get("scope"), "repository:test/repo:pull,push")

		user, pass, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		assert.Equal(t, "testuser", user)
		assert.Equal(t, "testpass", pass)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": "test.jwt.token"})
	}))
	defer tokenServer.Close()

	originalTokenURL := getTokenServiceURL
	getTokenServiceURL = func() string {
		return tokenServer.URL
	}
	defer func() { getTokenServiceURL = originalTokenURL }()

	// Reset token cache
	tokenCache.mu.Lock()
	tokenCache.data = make(map[string]*cachedToken)
	tokenCache.mu.Unlock()

	t.Setenv("REGISTRY_CREDENTIAL_USERNAME", "testuser")
	t.Setenv("REGISTRY_CREDENTIAL_PASSWORD", "testpass")

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/v2/test/repo/manifests/latest", nil)
	token := getRegistryToken(req)
	assert.Equal(t, "test.jwt.token", token)

	// Verify caching: second call should return cached token without hitting server
	token2 := getRegistryToken(req)
	assert.Equal(t, "test.jwt.token", token2)
}

func TestGetRegistryTokenCaching(t *testing.T) {
	callCount := 0
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": "cached.jwt.token"})
	}))
	defer tokenServer.Close()

	originalTokenURL := getTokenServiceURL
	getTokenServiceURL = func() string {
		return tokenServer.URL
	}
	defer func() { getTokenServiceURL = originalTokenURL }()

	tokenCache.mu.Lock()
	tokenCache.data = make(map[string]*cachedToken)
	tokenCache.mu.Unlock()

	t.Setenv("REGISTRY_CREDENTIAL_USERNAME", "u")
	t.Setenv("REGISTRY_CREDENTIAL_PASSWORD", "p")

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/v2/repo", nil)

	// First call should hit the server
	token1 := getRegistryToken(req)
	require.Equal(t, "cached.jwt.token", token1)
	assert.Equal(t, 1, callCount)

	// Second call should use cached token
	token2 := getRegistryToken(req)
	assert.Equal(t, "cached.jwt.token", token2)
	assert.Equal(t, 1, callCount, "should not call token server again (cached)")

	// Force expiry and verify re-fetch
	tokenCache.mu.Lock()
	scope := scopeFromRequest(req)
	if cached, ok := tokenCache.data[scope]; ok {
		cached.expires = time.Now().Add(-1 * time.Minute)
	}
	tokenCache.mu.Unlock()

	token3 := getRegistryToken(req)
	assert.Equal(t, "cached.jwt.token", token3)
	assert.Equal(t, 2, callCount, "should call token server again after expiry")
}

func TestProbeRegistryBasic(t *testing.T) {
	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Www-Authenticate", `Basic realm="Registry Realm"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer registryServer.Close()

	t.Setenv("REGISTRY_URL", registryServer.URL)
	originalClient := probeHTTPClient
	probeHTTPClient = registryServer.Client()
	defer func() { probeHTTPClient = originalClient }()

	authType, err := probeRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "basic", authType)
}

func TestProbeRegistryBearer(t *testing.T) {
	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Www-Authenticate", `Bearer realm="http://example.com/service/token",service="harbor-registry"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer registryServer.Close()

	t.Setenv("REGISTRY_URL", registryServer.URL)
	originalClient := probeHTTPClient
	probeHTTPClient = registryServer.Client()
	defer func() { probeHTTPClient = originalClient }()

	authType, err := probeRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "token", authType)
}

func TestProbeRegistryUnknownAuthDefaultsToBasic(t *testing.T) {
	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer registryServer.Close()

	t.Setenv("REGISTRY_URL", registryServer.URL)
	originalClient := probeHTTPClient
	probeHTTPClient = registryServer.Client()
	defer func() { probeHTTPClient = originalClient }()

	authType, err := probeRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "basic", authType, "unknown auth type should default to basic")
}

func TestGetRegistryTokenEmptyCredentials(t *testing.T) {
	tokenCache.mu.Lock()
	tokenCache.data = make(map[string]*cachedToken)
	tokenCache.mu.Unlock()

	originalTokenURL := getTokenServiceURL
	getTokenServiceURL = func() string {
		return "http://127.0.0.1:1/token"
	}
	defer func() { getTokenServiceURL = originalTokenURL }()

	t.Setenv("REGISTRY_CREDENTIAL_USERNAME", "")
	t.Setenv("REGISTRY_CREDENTIAL_PASSWORD", "")

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/v2/repo", nil)
	token := getRegistryToken(req)
	assert.Empty(t, token, "token should be empty when credentials are empty")
}
