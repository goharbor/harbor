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
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
)

var proxy = newProxy()

type cachedToken struct {
	token   string
	expires time.Time
}

var tokenCache struct {
	mu   sync.RWMutex
	data map[string]*cachedToken
}

func init() {
	tokenCache.data = make(map[string]*cachedToken)
}

var detectedAuthType atomic.Value

// Override in tests to control the HTTP client used for registry probe.
var probeHTTPClient = defaultProbeClient()

// Override in tests to control the HTTP client used for token exchange.
var exchangeHTTPClient = defaultProbeClient()

func defaultProbeClient() *http.Client {
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: commonhttp.GetHTTPTransport(),
	}
}

// Override in tests to control the token service endpoint.
var getTokenServiceURL = func() string {
	return config.InternalTokenServiceEndpoint()
}

func newProxy() http.Handler {
	regURL, _ := config.RegistryURL()
	url, err := url.Parse(regURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse URL of registry: %v", err))
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	if commonhttp.InternalTLSEnabled() {
		proxy.Transport = commonhttp.GetHTTPTransport()
	}

	proxy.Director = authDirector(proxy.Director)
	return proxy
}

// authDirector returns a Director that authenticates to the upstream registry.
// If the request has Bearer auth (from the user's token), it passes through.
// If the request has Basic auth (username:password from docker login), it
// exchanges it for a Bearer token via the token service before forwarding,
// since the upstream registry uses token-based auth.
// Otherwise it falls back to the shared registry credential.
func authDirector(d func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		d(r)
		if r == nil {
			return
		}
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			return // pass through user's Bearer token
		}
		if strings.HasPrefix(auth, "Basic ") {
			// Exchange Basic auth for a Bearer token via the token service
			tk, err := exchangeBasicForToken(r, auth)
			if err != nil {
				log.Warningf("failed to exchange basic auth: %v, using shared registry credential", err)
			} else if tk != "" {
				r.Header.Set("Authorization", "Bearer "+tk)
				return
			}
		}
		switch detectRegistryAuthType() {
		case "token":
			if tk := getRegistryToken(r); tk != "" {
				r.Header.Set("Authorization", "Bearer "+tk)
			}
		default:
			u, p := config.RegistryCredential()
			r.SetBasicAuth(u, p)
		}
	}
}

// exchangeBasicForToken sends the Basic auth credentials to the token service
// and returns a Bearer token scoped to the request's repository.
func exchangeBasicForToken(r *http.Request, basicAuth string) (string, error) {
	scope := scopeFromRequest(r)
	tokenURL := fmt.Sprintf("%s?service=harbor-registry&scope=%s",
		getTokenServiceURL(), url.QueryEscape(scope))
	req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Authorization", basicAuth)
	resp, err := exchangeHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange basic auth for token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token service returned %d for basic auth exchange", resp.StatusCode)
	}
	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}
	if tokenResp.Token == "" {
		return "", fmt.Errorf("token service returned empty token")
	}
	return tokenResp.Token, nil
}

// detectRegistryAuthType probes the upstream registry to determine which
// authentication scheme it expects (bearer token or basic auth). It first
// tries basic auth with the shared registry credential; if the registry
// responds with a Bearer challenge it returns "token".  The result is cached
// on success; on probe failure "basic" is returned as a safe default and the
// probe is retried on the next request.
func detectRegistryAuthType() string {
	if v := detectedAuthType.Load(); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}

	authType, err := probeRegistry()
	if err != nil {
		log.Warningf("registry auth probe failed: %v, using basic auth as default", err)
		return "basic"
	}

	detectedAuthType.Store(authType)
	return authType
}

// probeRegistry makes a request to the registry's /v2/ endpoint with the
// shared registry credential as basic auth.  If the registry accepts the
// credential (any non-401 response) it returns "basic".  If the registry
// returns 401 with a Www-Authenticate: Bearer challenge it returns "token".
func probeRegistry() (string, error) {
	regURL, err := config.RegistryURL()
	if err != nil {
		return "", fmt.Errorf("failed to get registry URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, strings.TrimSuffix(regURL, "/")+"/v2/", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create probe request: %w", err)
	}
	u, p := config.RegistryCredential()
	if u != "" {
		req.SetBasicAuth(u, p)
	}

	resp, err := probeHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to probe registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		wwwAuth := resp.Header.Get("Www-Authenticate")
		if strings.HasPrefix(wwwAuth, "Bearer") {
			return "token", nil
		}
	}
	return "basic", nil
}

// getRegistryToken obtains a bearer token for the upstream registry by
// exchanging the shared registry credential with Harbor's /service/token
// endpoint. The token is cached for 30 minutes per scope.
func getRegistryToken(r *http.Request) string {
	scope := scopeFromRequest(r)

	tokenCache.mu.RLock()
	if cached, ok := tokenCache.data[scope]; ok && cached.token != "" && time.Now().Before(cached.expires) {
		tk := cached.token
		tokenCache.mu.RUnlock()
		return tk
	}
	tokenCache.mu.RUnlock()

	tokenCache.mu.Lock()
	defer tokenCache.mu.Unlock()

	// Double-check after acquiring write lock
	if cached, ok := tokenCache.data[scope]; ok && cached.token != "" && time.Now().Before(cached.expires) {
		return cached.token
	}

	tokenURL := fmt.Sprintf("%s?service=harbor-registry&scope=%s", getTokenServiceURL(), url.QueryEscape(scope))

	req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
	if err != nil {
		log.Warningf("failed to create token request: %v", err)
		return ""
	}

	username, password := config.RegistryCredential()
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Warningf("failed to get registry token: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warningf("token service returned status %d", resp.StatusCode)
		return ""
	}

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		log.Warningf("failed to decode token response: %v", err)
		return ""
	}
	if tokenResp.Token == "" {
		return ""
	}

	tokenCache.data[scope] = &cachedToken{
		token:   tokenResp.Token,
		expires: time.Now().Add(30 * time.Minute),
	}
	return tokenResp.Token
}

// scopeFromRequest extracts the Docker registry scope from the request path.
// For a path like /v2/library/nginx/manifests/latest it returns
// repository:library/nginx:pull,push.
func scopeFromRequest(r *http.Request) string {
	if r == nil || r.URL == nil {
		return "repository:*:pull,push"
	}
	path := r.URL.Path
	if !strings.HasPrefix(path, "/v2/") {
		return "repository:*:pull,push"
	}
	parts := strings.SplitN(strings.TrimPrefix(path, "/v2/"), "/", 3)
	if len(parts) < 2 {
		return "repository:*:pull,push"
	}
	return fmt.Sprintf("repository:%s/%s:pull,push", parts[0], parts[1])
}
