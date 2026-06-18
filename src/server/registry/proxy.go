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
	"encoding/json"
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
	u, err := url.Parse(regURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse the URL of registry: %v", err))
	}
	p := httputil.NewSingleHostReverseProxy(u)
	if commonhttp.InternalTLSEnabled() {
		p.Transport = commonhttp.GetHTTPTransport()
	}

	p.Director = authDirector(p.Director)
	return p
}

func authDirector(d func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		d(r)
		if r == nil {
			return
		}

		if usesTokenAuth() {
			tk := getRegistryToken()
			if tk != "" {
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk))
			}
		} else {
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

	if tokenCache.data.token != "" && time.Now().Before(tokenCache.data.expires) {
		return tokenCache.data.token
	}

	urls := []string{
		config.InternalCoreURL(),
		config.LocalCoreURL(),
	}
	if extURL, err := config.ExtEndpoint(); err == nil && extURL != "" {
		urls = append(urls, extURL)
	}

	for _, baseURL := range urls {
		if baseURL == "" {
			continue
		}
		tokenURL := fmt.Sprintf("%s/service/token?service=harbor-registry&scope=repository:*:pull,push", strings.TrimSuffix(baseURL, "/"))

		req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
		if err != nil {
			log.Warningf("failed to create token request for %s: %v", baseURL, err)
			continue
		}

		username, password := config.RegistryCredential()
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Warningf("failed to get token from %s: %v", baseURL, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Warningf("token service at %s returned status %d", baseURL, resp.StatusCode)
			continue
		}

		var tokenResp struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			log.Warningf("failed to decode token response from %s: %v", baseURL, err)
			continue
		}

		if tokenResp.Token != "" {
			tokenCache.data.token = tokenResp.Token
			tokenCache.data.expires = time.Now().Add(30 * time.Minute)
			return tokenResp.Token
		}
	}

	log.Error("failed to get token from any token service endpoint")
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