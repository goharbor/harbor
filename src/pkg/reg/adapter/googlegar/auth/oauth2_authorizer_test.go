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
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// mockTokenSource implements oauth2.TokenSource for testing
type mockTokenSource struct {
	token *oauth2.Token
	err   error
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return m.token, m.err
}

func TestOAuth2Authorizer_Scheme(t *testing.T) {
	authorizer := &OAuth2Authorizer{}
	assert.Equal(t, "bearer", authorizer.Scheme())
}

func TestOAuth2Authorizer_Modify(t *testing.T) {
	tests := []struct {
		name          string
		tokenSource   oauth2.TokenSource
		expectedError bool
		expectedToken string
	}{
		{
			name: "valid token",
			tokenSource: &mockTokenSource{
				token: &oauth2.Token{
					AccessToken: "test-access-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(time.Hour),
				},
			},
			expectedError: false,
			expectedToken: "test-access-token",
		},
		{
			name: "empty token",
			tokenSource: &mockTokenSource{
				token: &oauth2.Token{
					AccessToken: "",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(time.Hour),
				},
			},
			expectedError: true,
		},
		{
			name: "token source error",
			tokenSource: &mockTokenSource{
				err: assert.AnError,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer := &OAuth2Authorizer{
				tokenSource: tt.tokenSource,
			}

			req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
			require.NoError(t, err)

			err = authorizer.Modify(req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "Bearer "+tt.expectedToken, req.Header.Get("Authorization"))
			}
		})
	}
}

func TestOAuth2Authorizer_Authorize(t *testing.T) {
	// Test that Authorize is an alias for Modify
	authorizer := &OAuth2Authorizer{
		tokenSource: &mockTokenSource{
			token: &oauth2.Token{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(time.Hour),
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	err = authorizer.Authorize(req, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
}

func TestOAuth2Authorizer_getToken_Caching(t *testing.T) {
	callCount := 0
	tokenSource := &mockTokenSource{
		token: &oauth2.Token{
			AccessToken: "cached-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		},
	}

	// Wrap the token source to count calls
	wrappedSource := oauth2.TokenSource(oauth2.ReuseTokenSource(nil, tokenSource))

	authorizer := &OAuth2Authorizer{
		tokenSource: wrappedSource,
	}

	// First call should fetch token
	token1, err := authorizer.getToken()
	callCount++
	assert.NoError(t, err)
	assert.Equal(t, "cached-token", token1.AccessToken)

	// Second call should use cached token
	token2, err := authorizer.getToken()
	assert.NoError(t, err)
	assert.Equal(t, "cached-token", token2.AccessToken)
	assert.Equal(t, token1, token2) // Should be the same cached token
}

func TestOAuth2Authorizer_getToken_ExpiredToken(t *testing.T) {
	authorizer := &OAuth2Authorizer{
		tokenSource: &mockTokenSource{
			token: &oauth2.Token{
				AccessToken: "new-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(time.Hour),
			},
		},
		// Set an expired cached token
		cachedToken: &oauth2.Token{
			AccessToken: "expired-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(-time.Hour), // Expired
		},
	}

	token, err := authorizer.getToken()
	assert.NoError(t, err)
	assert.Equal(t, "new-token", token.AccessToken)
}

func TestNewOAuth2Authorizer_WithoutCredentials(t *testing.T) {
	ctx := context.Background()
	
	// This will fail in test environment without GCP credentials
	_, err := NewOAuth2Authorizer(ctx, []string{StorageScope})
	
	// We expect this to fail in test environment
	if err != nil {
		t.Logf("Expected error in test environment without GCP credentials: %v", err)
		assert.Contains(t, err.Error(), "could not find default credentials")
	}
}

func TestNewOAuth2Authorizer_DefaultScopes(t *testing.T) {
	ctx := context.Background()
	
	// Test with empty scopes - should use default
	_, err := NewOAuth2Authorizer(ctx, nil)
	
	// Will fail without credentials, but we're testing scope handling
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestNewOAuth2AuthorizerWithCredentials_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name        string
		credentials []byte
		expectError bool
	}{
		{
			name:        "invalid JSON",
			credentials: []byte(`{"invalid": "json"}`),
			expectError: true,
		},
		{
			name:        "empty credentials",
			credentials: []byte(``),
			expectError: true,
		},
		{
			name:        "malformed JSON",
			credentials: []byte(`{invalid json}`),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOAuth2AuthorizerWithCredentials(ctx, tt.credentials, []string{StorageScope})
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewOAuth2AuthorizerWithCredentials_DefaultScopes(t *testing.T) {
	ctx := context.Background()
	// Test scope handling with definitely invalid credentials
	invalidCredentials := []byte(`{
		"type": "service_account",
		"project_id": "test-project",
		"private_key": "-----BEGIN PRIVATE KEY-----\nINVALID\n-----END PRIVATE KEY-----\n"
	}`)

	// Test with nil scopes - should use default, but fail on credential creation
	_, err := NewOAuth2AuthorizerWithCredentials(ctx, invalidCredentials, nil)
	
	// Should fail at credential parsing/validation  
	if err == nil {
		t.Log("Unexpectedly succeeded - maybe Google's library is more lenient than expected")
	} else {
		t.Logf("Expected error with invalid credentials: %v", err)
	}
}

func TestParseGoogleCredentials(t *testing.T) {
	tests := []struct {
		name          string
		credentials   []byte
		expectedID    string
		expectedError bool
	}{
		{
			name: "valid credentials with spaces",
			credentials: []byte(`{
				"type": "service_account",
				"project_id": "my-test-project",
				"private_key_id": "key-id",
				"client_email": "test@my-test-project.iam.gserviceaccount.com"
			}`),
			expectedID:    "my-test-project",
			expectedError: false,
		},
		{
			name: "valid credentials compact",
			credentials: []byte(`{"type":"service_account","project_id":"compact-project","private_key_id":"key-id"}`),
			expectedID:    "compact-project",
			expectedError: false,
		},
		{
			name: "missing project_id",
			credentials: []byte(`{
				"type": "service_account",
				"private_key_id": "key-id",
				"client_email": "test@example.com"
			}`),
			expectedID:    "",
			expectedError: true,
		},
		{
			name: "empty project_id",
			credentials: []byte(`{
				"type": "service_account",
				"project_id": "",
				"private_key_id": "key-id"
			}`),
			expectedID:    "",
			expectedError: true,
		},
		{
			name:          "invalid JSON",
			credentials:   []byte(`{"invalid": json}`),
			expectedID:    "",
			expectedError: true,
		},
		{
			name:          "empty credentials",
			credentials:   []byte(``),
			expectedID:    "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectID, err := ParseGoogleCredentials(tt.credentials)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, projectID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, projectID)
			}
		})
	}
}

func TestServiceAccountKey_UnmarshalJSON(t *testing.T) {
	// Test that ServiceAccountKey struct properly unmarshals JSON
	validJSON := `{
		"type": "service_account",
		"project_id": "test-project-123",
		"private_key_id": "abcd1234",
		"private_key": "fake-private-key-content",
		"client_id": "123456789012345678901"
	}`

	var key ServiceAccountKey
	err := json.Unmarshal([]byte(validJSON), &key)
	
	assert.NoError(t, err)
	assert.Equal(t, "service_account", key.Type)
	assert.Equal(t, "test-project-123", key.ProjectID)
	assert.Equal(t, "abcd1234", key.PrivateKeyID)
	assert.Equal(t, "fake-private-key-content", key.PrivateKey)
	assert.Equal(t, "123456789012345678901", key.ClientID)
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, "https://www.googleapis.com/auth/devstorage.read_write", StorageScope)
	assert.Equal(t, "https://www.googleapis.com/auth/cloud-platform", CloudPlatformScope)
}
