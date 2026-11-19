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

package huggingface

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestFactory(t *testing.T) {
	f, err := adp.GetFactory(model.RegistryTypeHuggingFace)
	require.NoError(t, err)
	a, err := f.Create(&model.Registry{URL: DefaultEndpoint})
	require.NoError(t, err)
	assert.NotNil(t, a)

	pattern := f.AdapterPattern()
	assert.Equal(t, model.EndpointPatternTypeList, pattern.EndpointPattern.EndpointType)
	assert.Equal(t, DefaultEndpoint, pattern.EndpointPattern.Endpoints[0].Value)
	assert.Equal(t, model.AccessKeyTypeFix, pattern.CredentialPattern.AccessKeyType)
}

func TestInfo(t *testing.T) {
	info, err := New(&model.Registry{}).Info()
	require.NoError(t, err)
	assert.Equal(t, model.RegistryTypeHuggingFace, info.Type)
	// must not be a replication candidate
	assert.Empty(t, info.SupportedResourceTypes)
}

func TestEndpointAndToken(t *testing.T) {
	assert.Equal(t, DefaultEndpoint, Endpoint(&model.Registry{}))
	assert.Equal(t, "https://mirror.example.com", Endpoint(&model.Registry{URL: "https://mirror.example.com/"}))
	assert.Equal(t, "", Token(&model.Registry{}))
	assert.Equal(t, "hf_x", Token(&model.Registry{Credential: &model.Credential{AccessSecret: " hf_x "}}))
}

func TestHealthCheck(t *testing.T) {
	var gotPath, gotAuth string
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if r.URL.Path == "/api/whoami-v2" {
			_, _ = w.Write([]byte(`{"type":"user","name":"test-user"}`))
		} else {
			_, _ = w.Write([]byte(`[{"id":"test/model"}]`))
		}
	}))
	defer server.Close()

	// anonymous
	a := New(&model.Registry{URL: server.URL})
	s, err := a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, s)
	assert.Equal(t, "/api/models", gotPath)
	assert.Equal(t, "", gotAuth)

	// with token
	a = New(&model.Registry{URL: server.URL, Credential: &model.Credential{AccessSecret: "hf_token"}})
	s, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, s)
	assert.Equal(t, "/api/whoami-v2", gotPath)
	assert.Equal(t, "Bearer hf_token", gotAuth)

	// invalid token
	status = http.StatusUnauthorized
	s, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Unhealthy, s)

	// unreachable
	server.Close()
	s, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Unhealthy, s)
}

func TestHealthCheckResponse(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		body     string
		redirect bool
		healthy  bool
	}{
		{name: "models", body: `[{"id":"test/model","private":false}]`, healthy: true},
		{name: "empty models", body: `[]`, healthy: true},
		{name: "models HTML fallback", body: `<html><body>Model hub</body></html>`},
		{name: "models empty body"},
		{name: "models null", body: `null`},
		{name: "models object", body: `{}`},
		{name: "models error", body: `{"error":"unavailable"}`},
		{name: "models missing ID", body: `[{}]`},
		{name: "models blank ID", body: `[{"id":" "}]`},
		{name: "models invalid ID type", body: `[{"id":123}]`},
		{name: "models null entry", body: `[null]`},
		{name: "models malformed JSON", body: `[{"id":"test/model"}`},
		{name: "models trailing HTML", body: `[{"id":"test/model"}]<html></html>`},
		{name: "models login redirect", body: `<html><body>Login</body></html>`, redirect: true},
		{name: "models API redirect", body: `[{"id":"test/model"}]`, redirect: true, healthy: true},
		{name: "identity", token: "hf_token", body: `{"type":"user","name":"test-user"}`, healthy: true},
		{name: "identity HTML fallback", token: "hf_token", body: `<html><body>Login</body></html>`},
		{name: "identity empty body", token: "hf_token"},
		{name: "identity null", token: "hf_token", body: `null`},
		{name: "identity empty object", token: "hf_token", body: `{}`},
		{name: "identity error", token: "hf_token", body: `{"error":"Invalid token"}`},
		{name: "identity blank name", token: "hf_token", body: `{"name":" "}`},
		{name: "identity invalid name type", token: "hf_token", body: `{"name":123}`},
		{name: "identity array", token: "hf_token", body: `[]`},
		{name: "identity malformed JSON", token: "hf_token", body: `{"name":"test-user"`},
		{name: "identity trailing JSON", token: "hf_token", body: `{"name":"test-user"}{}`},
		{name: "identity login redirect", token: "hf_invalid", body: `<html><body>Login</body></html>`, redirect: true},
		{name: "identity API redirect", token: "hf_token", body: `{"type":"user","name":"test-user"}`, redirect: true, healthy: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redirected := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.redirect && r.URL.Path != "/redirected" {
					http.Redirect(w, r, "/redirected", http.StatusFound)
					return
				}
				redirected = r.URL.Path == "/redirected"
				// Even an HTML fallback mislabeled as JSON must be rejected.
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			registry := &model.Registry{URL: server.URL}
			if tt.token != "" {
				registry.Credential = &model.Credential{AccessSecret: tt.token}
			}
			status, err := New(registry).HealthCheck()
			require.NoError(t, err)
			expected := model.Unhealthy
			if tt.healthy {
				expected = model.Healthy
			}
			assert.Equal(t, expected, status)
			assert.Equal(t, tt.redirect, redirected)
		})
	}
}

func TestPrepareForPush(t *testing.T) {
	assert.Error(t, New(&model.Registry{}).PrepareForPush(nil))
}
