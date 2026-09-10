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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib"
	libcache "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	cacheCapacity = 100
)

// NewAuthorizer return a bearer token authorizer
// The parameter "a" is an authorizer used to fetch the token
func NewAuthorizer(realm, service string, a lib.Authorizer, transport http.RoundTripper) lib.Authorizer {
	authorizer := &authorizer{
		realm:      realm,
		service:    service,
		authorizer: a,
		cache:      newCache(cacheCapacity),
	}

	authorizer.client = &http.Client{
		Transport: transport,
		Timeout:   config.RegistryHTTPClientTimeout(),
	}
	return authorizer
}

type authorizer struct {
	realm      string
	service    string
	authorizer lib.Authorizer
	cache      *cache
	client     *http.Client
}

func (a *authorizer) Modify(req *http.Request) error {
	// parse scopes from request
	scopes := parseScopes(req)

	// get token
	token, err := a.getToken(scopes)
	if err != nil {
		return err
	}

	// set authorization header
	if token != nil && len(token.Token) > 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	}
	return nil
}

func (a *authorizer) getToken(scopes []*scope) (*token, error) {
	// get token from cache first
	token := a.cache.get(scopes)
	if token != nil {
		return token, nil
	}

	// get no token from cache, fetch it from the token service
	token, err := a.fetchToken(scopes)
	if err != nil {
		return nil, err
	}

	// set the token into the cache
	a.cache.set(scopes, token)
	return token, nil
}

type token struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"` // the token returned by azure container registry is called "access_token"
	ExpiresIn   int    `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}

func (a *authorizer) fetchToken(scopes []*scope) (*token, error) {
	req, err := a.tokenRequest(scopes)
	if err != nil {
		return nil, err
	}
	c := libcache.Default()
	key := tokenCacheKey(req)
	if c != nil {
		cached := &token{}
		if c.Fetch(req.Context(), key, cached) == nil && cached.Token != "" {
			if expired, _ := a.cache.expired(cached); !expired {
				return cached, nil
			}
		}
	}
	requestedAt := time.Now()

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		code := errors.GeneralCode
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			code = errors.UnAuthorizedCode
		case http.StatusForbidden:
			code = errors.ForbiddenCode
		}
		return nil, errors.New(nil).WithCode(code).
			WithMessagef("http status code: %d, body: %s", resp.StatusCode, string(body))
	}
	token := &token{}
	if err = json.Unmarshal(body, token); err != nil {
		return nil, err
	}
	if len(token.Token) == 0 && len(token.AccessToken) > 0 {
		token.Token = token.AccessToken
	}
	if token.IssuedAt == "" {
		token.IssuedAt = requestedAt.UTC().Format(time.RFC3339)
	}
	if token.ExpiresIn == 0 {
		// The registry token specification defaults an omitted expires_in to 60 seconds.
		// JSON decoding leaves this int at zero when omitted; an explicit zero is
		// indistinguishable here and receives the same default.
		// https://distribution.github.io/distribution/spec/auth/token/#token-response-fields
		token.ExpiresIn = 60
	}
	if c != nil && token.Token != "" {
		if expired, expiresAt := a.cache.expired(token); !expired {
			// Cache failures must not prevent authenticated pulls.
			_ = c.Save(req.Context(), key, token, time.Until(expiresAt))
		}
	}
	return token, nil
}

// tokenRequest includes the credentials before deriving the cache identity, so
// rotating credentials cannot reuse tokens obtained with the previous secret.
func (a *authorizer) tokenRequest(scopes []*scope) (*http.Request, error) {
	url, err := url.Parse(a.realm)
	if err != nil {
		return nil, err
	}
	query := url.Query()
	query.Add("service", a.service)
	for _, scope := range scopes {
		query.Add("scope", scope.String())
	}
	url.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	// set user agent to avoid some registry (e.g. docker hub) return 403 when user agent is not set to harbor-registry-client
	utils.SetUserAgentHeader(req)
	if a.authorizer != nil {
		if err = a.authorizer.Modify(req); err != nil {
			return nil, err
		}
	}

	return req, nil
}

func tokenCacheKey(req *http.Request) string {
	identity, _ := json.Marshal([]any{req.URL.String(), req.Header})
	return fmt.Sprintf("{registry_bearer_token}:%x", sha256.Sum256(identity))
}

// Invalidate discards a token rejected by the registry.
func (a *authorizer) Invalidate(req *http.Request) {
	scopes := parseScopes(req)
	a.cache.Lock()
	delete(a.cache.cache, a.cache.key(scopes))
	a.cache.Unlock()
	if c := libcache.Default(); c != nil {
		if tokenReq, err := a.tokenRequest(scopes); err == nil {
			_ = c.Delete(context.Background(), tokenCacheKey(tokenReq))
		}
	}
}
