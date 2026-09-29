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

package googlegar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

type mockTokenSource struct {
	token  *oauth2.Token
	tokens []*oauth2.Token
	err    error
	calls  int
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	if len(m.tokens) > 0 {
		idx := m.calls - 1
		if idx >= len(m.tokens) {
			idx = len(m.tokens) - 1
		}
		return m.tokens[idx], nil
	}
	return m.token, nil
}

func getMockAdapter(t *testing.T, hasCred, health bool) (*adapter, *httptest.Server) {
	var server *httptest.Server
	server = test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/_catalog",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"repositories": ["test1"]}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/{repo}/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"name": "test1", "tags": ["latest", "v1.0"]}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/token",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				user, pass, ok := r.BasicAuth()
				if !ok || (user != jsonKeyUsername && user != oauth2AccessTokenUser) || pass == "" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"token": "issued-registry-jwt"}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if !health {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/v2/token",service="us-central1-docker.pkg.dev"`, server.URL))
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if authHeader != "Bearer issued-registry-jwt" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodDelete,
			Pattern: "/v2/{repo}/manifests/{reference}",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				if !health {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusAccepted)
			},
		},
	)

	registry := &model.Registry{
		Type: model.RegistryTypeGoogleGar,
		URL:  server.URL,
	}
	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    jsonKeyUsername,
			AccessSecret: `{"type":"service_account","project_id":"test-project"}`,
		}
		return &adapter{
			registry: registry,
			Adapter:  native.NewAdapter(registry),
		}, server
	}

	ts := &mockTokenSource{
		token: &oauth2.Token{
			AccessToken: "mock-gcp-access-token",
			Expiry:      time.Now().Add(time.Hour),
		},
	}
	return &adapter{
		registry: registry,
		Adapter:  native.NewAdapterWithAuthorizer(registry, newTokenAuthorizer(ts, registry.Insecure, registry.CACertificate)),
	}, server
}

func TestIsGoogleArtifactRegistry(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://us-central1-docker.pkg.dev", true},
		{"https://us-central1-docker.pkg.dev/", true},
		{"https://us-docker.pkg.dev", true},
		{"https://europe-docker.pkg.dev", true},
		{"https://asia-docker.pkg.dev", true},
		{"https://europe-west1-docker.pkg.dev", true},
		{"https://northamerica-northeast1-docker.pkg.dev", true},
		{"https://gcr.io", false},
		{"https://us.gcr.io", false},
		{"http://us-central1-docker.pkg.dev", false},
		{"https://us-central1-docker.pkg.dev/v2/", false},
		{"https://evil.example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			assert.Equal(t, tt.expected, isGoogleArtifactRegistry(tt.url))
		})
	}
}

func TestNewAdapter_URLValidation(t *testing.T) {
	f := &factory{}

	// Nil registry should be rejected
	ad, err := f.Create(nil)
	assert.Error(t, err)
	assert.Nil(t, ad)

	// Non-GAR URLs (including arbitrary hosts and gcr.io) must be rejected before touching credentials
	invalidURLs := []string{
		"",
		"https://gcr.io",
		"https://evil.example.com",
		"http://us-central1-docker.pkg.dev",
		"https://us-central1-docker.pkg.dev/subpath",
	}
	for _, u := range invalidURLs {
		ad, err = f.Create(&model.Registry{
			Type: model.RegistryTypeGoogleGar,
			URL:  u,
		})
		assert.Error(t, err, "expected error for URL %q", u)
		assert.Nil(t, ad)
	}
}

func TestNewAdapter_WithServiceAccountCredentials(t *testing.T) {
	tests := []struct {
		name              string
		inputAccessKey    string
		expectedAccessKey string
	}{
		{
			name:              "default empty access key normalizes to _json_key",
			inputAccessKey:    "",
			expectedAccessKey: jsonKeyUsername,
		},
		{
			name:              "oauth2 access key normalizes to _json_key",
			inputAccessKey:    "oauth2",
			expectedAccessKey: jsonKeyUsername,
		},
		{
			name:              "explicit _json_key preserved",
			inputAccessKey:    jsonKeyUsername,
			expectedAccessKey: jsonKeyUsername,
		},
		{
			name:              "explicit _json_key_base64 preserved",
			inputAccessKey:    "_json_key_base64",
			expectedAccessKey: "_json_key_base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &model.Registry{
				Type: model.RegistryTypeGoogleGar,
				URL:  "https://us-central1-docker.pkg.dev",
				Credential: &model.Credential{
					AccessKey:    tt.inputAccessKey,
					AccessSecret: `{"type":"service_account","project_id":"my-project"}`,
				},
			}
			a, err := newAdapter(reg)
			require.NoError(t, err)
			require.NotNil(t, a)
			// Normalized AccessKey must be written back to reg.Credential so
			// reg/manager.go persists the credential when saving to the database.
			assert.Equal(t, tt.expectedAccessKey, reg.Credential.AccessKey)
			assert.Equal(t, `{"type":"service_account","project_id":"my-project"}`, reg.Credential.AccessSecret)

			// Reloading the persisted registry must preserve the normalized service-account credential
			reloaded, err := newAdapter(reg)
			require.NoError(t, err)
			require.NotNil(t, reloaded)
			assert.Equal(t, tt.expectedAccessKey, reg.Credential.AccessKey)
		})
	}
}

func TestNewAdapter_WithADC(t *testing.T) {
	origTokenSource := defaultTokenSource
	t.Cleanup(func() {
		defaultTokenSource = origTokenSource
	})

	t.Run("ADC success uses oauth2accesstoken", func(t *testing.T) {
		var capturedScopes []string
		defaultTokenSource = func(_ context.Context, scopes ...string) (oauth2.TokenSource, error) {
			capturedScopes = scopes
			return &mockTokenSource{
				token: &oauth2.Token{
					AccessToken: "adc-token-123",
					Expiry:      time.Now().Add(time.Hour),
				},
			}, nil
		}

		reg := &model.Registry{
			Type: model.RegistryTypeGoogleGar,
			URL:  "https://us-docker.pkg.dev",
		}
		a, err := newAdapter(reg)
		require.NoError(t, err)
		require.NotNil(t, a)
		assert.Equal(t, []string{CloudPlatformScope}, capturedScopes)
		assert.Nil(t, reg.Credential)
	})

	t.Run("ADC with empty access secret falls back to defaultTokenSource", func(t *testing.T) {
		defaultTokenSource = func(_ context.Context, _ ...string) (oauth2.TokenSource, error) {
			return &mockTokenSource{
				token: &oauth2.Token{
					AccessToken: "adc-token-456",
					Expiry:      time.Now().Add(time.Hour),
				},
			}, nil
		}

		reg := &model.Registry{
			Type: model.RegistryTypeGoogleGar,
			URL:  "https://europe-docker.pkg.dev",
			Credential: &model.Credential{
				AccessKey:    jsonKeyUsername,
				AccessSecret: "   ",
			},
		}
		a, err := newAdapter(reg)
		require.NoError(t, err)
		require.NotNil(t, a)
		// Ambient ADC tokens and empty credentials must not be persisted
		assert.Empty(t, reg.Credential.AccessKey)
		assert.Empty(t, reg.Credential.AccessSecret)
	})

	t.Run("ADC token source error", func(t *testing.T) {
		defaultTokenSource = func(_ context.Context, _ ...string) (oauth2.TokenSource, error) {
			return nil, errors.New("no ambient credentials")
		}

		reg := &model.Registry{
			Type: model.RegistryTypeGoogleGar,
			URL:  "https://us-central1-docker.pkg.dev",
		}
		a, err := newAdapter(reg)
		assert.Error(t, err)
		assert.Nil(t, a)
		assert.Contains(t, err.Error(), "failed to create default token source")
	})

	t.Run("ADC token retrieval error", func(t *testing.T) {
		defaultTokenSource = func(_ context.Context, _ ...string) (oauth2.TokenSource, error) {
			return &mockTokenSource{
				err: errors.New("metadata server error"),
			}, nil
		}

		reg := &model.Registry{
			Type: model.RegistryTypeGoogleGar,
			URL:  "https://us-central1-docker.pkg.dev",
		}
		a, err := newAdapter(reg)
		assert.Error(t, err)
		assert.Nil(t, a)
		assert.Contains(t, err.Error(), "failed to get OAuth2 token")
	})

	t.Run("ADC empty access token error", func(t *testing.T) {
		defaultTokenSource = func(_ context.Context, _ ...string) (oauth2.TokenSource, error) {
			return &mockTokenSource{
				token: &oauth2.Token{AccessToken: ""},
			}, nil
		}

		reg := &model.Registry{
			Type: model.RegistryTypeGoogleGar,
			URL:  "https://us-central1-docker.pkg.dev",
		}
		a, err := newAdapter(reg)
		assert.Error(t, err)
		assert.Nil(t, a)
		assert.Contains(t, err.Error(), "empty OAuth2 access token")
	})
}

func TestFactory_AdapterPattern(t *testing.T) {
	f, err := adp.GetFactory(model.RegistryTypeGoogleGar)
	require.NoError(t, err)
	require.NotNil(t, f)

	pattern := f.AdapterPattern()
	require.NotNil(t, pattern)
	require.NotNil(t, pattern.EndpointPattern)
	assert.Equal(t, model.EndpointPatternTypeList, pattern.EndpointPattern.EndpointType)

	for _, ep := range pattern.EndpointPattern.Endpoints {
		assert.True(t, isGoogleArtifactRegistry(ep.Value), "endpoint %s (%s) must match garEndpointPattern", ep.Key, ep.Value)
	}

	require.NotNil(t, pattern.CredentialPattern)
	assert.Equal(t, model.AccessKeyTypeFix, pattern.CredentialPattern.AccessKeyType)
	assert.Equal(t, jsonKeyUsername, pattern.CredentialPattern.AccessKeyData)
	assert.Equal(t, model.AccessSecretTypeFile, pattern.CredentialPattern.AccessSecretType)
}

func TestAdapter_Info(t *testing.T) {
	a := &adapter{}
	info, err := a.Info()

	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, model.RegistryTypeGoogleGar, info.Type)
	assert.Equal(t, []string{model.ResourceTypeImage}, info.SupportedResourceTypes)
	assert.Len(t, info.SupportedResourceFilters, 2)
	assert.Equal(t, []string{model.TriggerTypeManual, model.TriggerTypeScheduled}, info.SupportedTriggers)
}

func TestAdapter_HealthCheck(t *testing.T) {
	// Healthy with explicit JSON key via token challenge flow
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	status, err := a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, status)

	// Healthy with ADC oauth2accesstoken via token challenge flow
	a, s = getMockAdapter(t, false, true)
	defer s.Close()
	status, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, status)

	// Unhealthy when registry returns error
	a, s = getMockAdapter(t, true, false)
	defer s.Close()
	status, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Unhealthy, status)
}

func TestAdapter_FetchArtifactsAndPrepareForPush(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()

	resources, err := a.FetchArtifacts([]*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "*",
		},
		{
			Type:  model.FilterTypeTag,
			Value: "*",
		},
	})
	require.NoError(t, err)
	require.Len(t, resources, 1)

	err = a.PrepareForPush(resources)
	require.NoError(t, err)
}

func TestAdapter_DeleteTag(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()

	err := a.DeleteTag("test1", "v1.0")
	assert.NoError(t, err)

	aUnhealthy, sUnhealthy := getMockAdapter(t, true, false)
	defer sUnhealthy.Close()

	err = aUnhealthy.DeleteTag("test1", "v1.0")
	assert.Error(t, err)

	// Invalid URL in registry triggers request creation error
	aBadURL := &adapter{
		registry: &model.Registry{URL: "://bad-url"},
		Adapter:  a.Adapter,
	}
	err = aBadURL.DeleteTag("test1", "v1.0")
	assert.Error(t, err)
}

func TestBuildManifestURL(t *testing.T) {
	assert.Equal(
		t,
		"https://us-central1-docker.pkg.dev/v2/my-project/my-repo/manifests/latest",
		buildManifestURL("https://us-central1-docker.pkg.dev", "my-project/my-repo", "latest"),
	)
	assert.Equal(
		t,
		"https://us-central1-docker.pkg.dev/v2/my-project/my-repo/manifests/latest",
		buildManifestURL("https://us-central1-docker.pkg.dev/", "my-project/my-repo", "latest"),
	)
}

func TestTokenAuthorizer_RefreshExpiredToken(t *testing.T) {
	expectedToken := "adc-token-initial"
	var tokenExchangePasswords []string

	var server *httptest.Server
	server = test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/token",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				user, pass, ok := r.BasicAuth()
				tokenExchangePasswords = append(tokenExchangePasswords, pass)
				if !ok || user != oauth2AccessTokenUser || pass != expectedToken {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{"token": "jwt-for-%s"}`, pass)))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/v2/token",service="us-central1-docker.pkg.dev"`, server.URL))
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if authHeader != fmt.Sprintf("Bearer jwt-for-%s", expectedToken) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
		},
	)
	defer server.Close()

	// First token is already within the expiry window so ReuseTokenSource
	// uses it only if passed directly or refreshes it on the next call.
	ts := &mockTokenSource{
		tokens: []*oauth2.Token{
			{
				AccessToken: "adc-token-initial",
				Expiry:      time.Now().Add(-time.Minute),
			},
			{
				AccessToken: "adc-token-refreshed",
				Expiry:      time.Now().Add(time.Hour),
			},
		},
	}

	registry := &model.Registry{
		Type: model.RegistryTypeGoogleGar,
		URL:  server.URL,
	}

	// Wire tokenAuthorizer directly to ts to verify rotation across exchanges
	authz := newTokenAuthorizer(ts, registry.Insecure, registry.CACertificate)
	a := &adapter{
		registry: registry,
		Adapter:  native.NewAdapterWithAuthorizer(registry, authz),
	}

	// First health check uses adc-token-initial
	status, err := a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, status)

	// Rotate expected token on server; second health check must refresh to adc-token-refreshed
	expectedToken = "adc-token-refreshed"
	status, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, status)
	assert.Equal(t, []string{"adc-token-initial", "adc-token-refreshed"}, tokenExchangePasswords)

	// Verify ReuseTokenSource integration when initial token expires
	expiredInitial := &oauth2.Token{
		AccessToken: "adc-token-expired",
		Expiry:      time.Now().Add(-time.Minute),
	}
	refreshTS := &mockTokenSource{
		token: &oauth2.Token{
			AccessToken: "adc-token-refreshed",
			Expiry:      time.Now().Add(time.Hour),
		},
	}
	reusableAuthz := newTokenAuthorizer(oauth2.ReuseTokenSource(expiredInitial, refreshTS), registry.Insecure, registry.CACertificate)
	aReusable := &adapter{
		registry: registry,
		Adapter:  native.NewAdapterWithAuthorizer(registry, reusableAuthz),
	}
	status, err = aReusable.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, status)
	assert.Equal(t, 1, refreshTS.calls)

	// Error and empty-token branches in tokenAuthorizer.Modify
	req, err := http.NewRequest(http.MethodGet, server.URL+"/v2/", nil)
	require.NoError(t, err)

	errAuthz := newTokenAuthorizer(&mockTokenSource{err: errors.New("refresh failed")}, false, "")
	err = errAuthz.Modify(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get OAuth2 token")

	emptyAuthz := newTokenAuthorizer(&mockTokenSource{token: &oauth2.Token{AccessToken: ""}}, false, "")
	err = emptyAuthz.Modify(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty OAuth2 access token")
}
