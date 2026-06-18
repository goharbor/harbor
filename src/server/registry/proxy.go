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
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
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

var tokenCache = struct {
	mu   sync.RWMutex
	data *cachedToken
}{data: &cachedToken{}}

func newProxy() http.Handler {
	regURL, _ := config.RegistryURL()
	url, err := url.Parse(regURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse the URL of registry: %v", err))
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	if commonhttp.InternalTLSEnabled() {
		proxy.Transport = commonhttp.GetHTTPTransport()
	}

	proxy.Director = authDirector(proxy.Director)
	return proxy
}

func authDirector(d func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		d(r)
		if r == nil {
			return
		}

		// Check if registry uses token-based auth (OIDC mode)
		if usesTokenAuth() {
			// Use Bearer token
			tk := getRegistryToken()
			if tk != "" {
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk))
			}
		} else {
			// Use basic auth (legacy mode)
			u, p := config.RegistryCredential()
			r.SetBasicAuth(u, p)
		}
	}
}

func usesTokenAuth() bool {
	ctx := context.Background()
	authMode, err := config.AuthMode(ctx)
	if err != nil {
		log.Warningf("failed to get auth mode: %v, defaulting to basic auth", err)
		return false
	}
	// OIDC auth mode requires token-based authentication to registry
	// Also check for UAA auth which also uses tokens
	return authMode == "oidc_auth" || authMode == "uaa_auth"
}

func getRegistryToken() string {
	tokenCache.mu.RLock()
	if tokenCache.data.token != "" && time.Now().Before(tokenCache.data.expires) {
		tk := tokenCache.data.token
		tokenCache.mu.RUnlock()
		return tk
	}
	tokenCache.mu.RUnlock()

	tokenCache.mu.Lock()
	defer tokenCache.mu.Unlock()

	// Double-check after acquiring write lock
	if tokenCache.data.token != "" && time.Now().Before(tokenCache.data.expires) {
		return tokenCache.data.token
	}

	// Generate token using the internal token service
	// We create a request to the token service and use the existing creator
	coreURL := config.InternalCoreURL()
	if coreURL == "" {
		log.Errorf("failed to get internal core URL")
		// Fall back to external URL
		coreURL, _ = config.ExtEndpoint()
	}

	tokenURL := fmt.Sprintf("%s/service/token?service=harbor-registry&scope=repository:*:pull,push", coreURL)
	tokenURL = strings.TrimSuffix(tokenURL, "/")

	req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
	if err != nil {
		log.Errorf("failed to create token request: %v", err)
		return ""
	}

	// Add basic auth from registry credentials to authenticate with token service
	// This is the same credentials used to access Harbor UI/API
	username, password := config.RegistryCredential()
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// Make the request to get the token
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("failed to get token from token service: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("token service returned status %d", resp.StatusCode)
		return ""
	}

	// Read the body
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	if n > 0 {
		// Try to parse as JSON
		if strings.Contains(string(buf[:n]), "token") {
			// Simple parsing - extract token from JSON
			tokenStart := strings.Index(string(buf[:n]), `"token":"`) + 8
			tokenEnd := strings.Index(string(buf[:n])[tokenStart:], `"`)
			if tokenStart > 7 && tokenEnd > 0 {
				tokenCache.data.token = string(buf[:n])[tokenStart : tokenStart+tokenEnd]
				tokenCache.data.expires = time.Now().Add(30 * time.Minute)
				return tokenCache.data.token
			}
		}
	}

	log.Errorf("failed to parse token response")
	return ""
}

func basicAuthDirector(d func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		d(r)
		if r != nil {
			u, p := config.RegistryCredential()
			r.SetBasicAuth(u, p)
		}
	}
}
