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
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// StorageScope is the Google Cloud Storage scope for Docker registry access
	StorageScope = "https://www.googleapis.com/auth/devstorage.read_write"
	// CloudPlatformScope is the Cloud Platform scope for broader access
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

// OAuth2Authorizer implements the lib.Authorizer interface using Google OAuth2
type OAuth2Authorizer struct {
	tokenSource oauth2.TokenSource
	mu          sync.RWMutex
	cachedToken *oauth2.Token
}

// NewOAuth2Authorizer creates a new OAuth2 authorizer using Google's default credential chain
func NewOAuth2Authorizer(ctx context.Context, scopes []string) (*OAuth2Authorizer, error) {
	if len(scopes) == 0 {
		scopes = []string{StorageScope}
	}

	tokenSource, err := google.DefaultTokenSource(ctx, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to create default token source: %w", err)
	}

	return &OAuth2Authorizer{
		tokenSource: tokenSource,
	}, nil
}

// NewOAuth2AuthorizerWithCredentials creates a new OAuth2 authorizer with explicit credentials
func NewOAuth2AuthorizerWithCredentials(ctx context.Context, credentialsJSON []byte, scopes []string) (*OAuth2Authorizer, error) {
	if len(scopes) == 0 {
		scopes = []string{StorageScope}
	}

	credentials, err := google.CredentialsFromJSON(ctx, credentialsJSON, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &OAuth2Authorizer{
		tokenSource: credentials.TokenSource,
	}, nil
}

// Modify implements the lib.Authorizer interface by adding OAuth2 bearer token to the request
func (a *OAuth2Authorizer) Modify(req *http.Request) error {
	token, err := a.getToken()
	if err != nil {
		return fmt.Errorf("failed to get OAuth2 token: %w", err)
	}

	if token.AccessToken == "" {
		return fmt.Errorf("empty access token")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	return nil
}

// Authorize is an alias for Modify to maintain compatibility
func (a *OAuth2Authorizer) Authorize(req *http.Request, _ map[string]string) error {
	return a.Modify(req)
}

// getToken retrieves and caches the OAuth2 token
func (a *OAuth2Authorizer) getToken() (*oauth2.Token, error) {
	a.mu.RLock()
	cached := a.cachedToken
	a.mu.RUnlock()

	// Check if we have a valid cached token
	if cached != nil && cached.Valid() {
		return cached, nil
	}

	// Get a new token
	token, err := a.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	// Cache the token
	a.mu.Lock()
	a.cachedToken = token
	a.mu.Unlock()

	log.Debugf("OAuth2 token obtained, expires at: %v", token.Expiry)
	return token, nil
}

// Scheme returns the authentication scheme
func (a *OAuth2Authorizer) Scheme() string {
	return "bearer"
}

// AuthorizeRequest is an alias for Authorize to match the interface
func (a *OAuth2Authorizer) AuthorizeRequest(req *http.Request, params map[string]string) error {
	return a.Authorize(req, params)
}

// ServiceAccountKey represents the structure of a Google service account JSON key
type ServiceAccountKey struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientID     string `json:"client_id"`
}

// ParseGoogleCredentials attempts to determine project ID from credentials JSON
func ParseGoogleCredentials(credentialsJSON []byte) (projectID string, err error) {
	var key ServiceAccountKey
	if err := json.Unmarshal(credentialsJSON, &key); err != nil {
		return "", fmt.Errorf("failed to parse credentials JSON: %w", err)
	}

	if key.ProjectID == "" {
		return "", fmt.Errorf("project_id not found in credentials")
	}

	return key.ProjectID, nil
}
