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
	"net/url"
	"strings"
	"sync"

	"github.com/docker/distribution/registry/client/auth/challenge"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
	"github.com/goharbor/harbor/src/pkg/registry/auth/bearer"
	"github.com/goharbor/harbor/src/pkg/registry/auth/null"
)

// NewAuthorizer creates an authorizer that can handle different auth schemes
func NewAuthorizer(username, password string, insecure bool, caCert ...string) lib.Authorizer {
	var transport http.RoundTripper
	if len(caCert) > 0 && caCert[0] != "" {
		transport = commonhttp.GetHTTPTransport(
			commonhttp.WithInsecure(insecure),
			commonhttp.WithCACert(caCert[0]),
		)
	} else {
		transport = commonhttp.GetHTTPTransport(commonhttp.WithInsecure(insecure))
	}

	return &authorizer{
		username: username,
		password: password,
		client: &http.Client{
			Transport: transport,
			Timeout:   config.RegistryHTTPClientTimeout(),
		},
	}
}

// authorizer authorizes the request with the provided credential.
// It determines the auth scheme of registry automatically and calls
// different underlying authorizers to do the auth work
type authorizer struct {
	sync.RWMutex
	username   string
	password   string
	client     *http.Client
	url        *url.URL          // registry URL
	authorizer modifier.Modifier // the underlying authorizer
}

func (a *authorizer) Modify(req *http.Request) error {
	a.RLock()
	auth := a.authorizer
	targetURL := a.url
	a.RUnlock()

	// Nil underlying authorizer means this is the first time the authorizer is called
	// Try to connect to the registry and determine the auth scheme
	if auth == nil {
		// to avoid concurrent issue
		a.Lock()
		if err := a.initialize(req.URL); err != nil {
			a.Unlock()
			return err
		}
		auth = a.authorizer
		targetURL = a.url
		a.Unlock()
	}

	// check whether the request targets the registry
	if !isTarget(req, targetURL) {
		return nil
	}

	return auth.Modify(req)
}

func (a *authorizer) initialize(u *url.URL) error {
	if a.authorizer != nil {
		return nil
	}
	// Extract the path prefix before "/v2/" so that registries served under
	// a sub-path (e.g. https://hostname/prefix/v2/...) are probed correctly.
	prefix := ""
	if idx := strings.Index(u.Path, "/v2/"); idx > 0 {
		prefix = u.Path[:idx]
	}
	url, err := url.Parse(u.Scheme + "://" + u.Host + prefix + "/v2/")
	if err != nil {
		return err
	}
	a.url = url
	req, err := http.NewRequest(http.MethodGet, a.url.String(), nil)
	if err != nil {
		return err
	}
	// set user agent to avoid some registry (e.g. docker hub) return 403 when user agent is not set to harbor-registry-client
	utils.SetUserAgentHeader(req)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	challenges := challenge.ResponseChallenges(resp)
	// no challenge, mean no auth
	if len(challenges) == 0 {
		a.authorizer = null.NewAuthorizer()
		return nil
	}
	cm := map[string]challenge.Challenge{}
	for _, challenge := range challenges {
		cm[challenge.Scheme] = challenge
	}
	if challenge, exist := cm["bearer"]; exist {
		a.authorizer = bearer.NewAuthorizer(challenge.Parameters["realm"],
			challenge.Parameters["service"], basic.NewAuthorizer(a.username, a.password),
			a.client.Transport)
		return nil
	}
	if _, exist := cm["basic"]; exist {
		a.authorizer = basic.NewAuthorizer(a.username, a.password)
		return nil
	}
	return fmt.Errorf("unsupported auth scheme: %v", challenges)
}

// Check whether the request targets to the registry.
// If doesn't, the request shouldn't be handled by the authorizer.
// e.g. the requests sent to backend storage(s3, etc.)
func (a *authorizer) isTarget(req *http.Request) bool {
	return isTarget(req, a.url)
}

func isTarget(req *http.Request, target *url.URL) bool {
	index := strings.Index(req.URL.Path, "/v2/")
	if index == -1 {
		return false
	}
	if !strings.EqualFold(req.URL.Host, target.Host) || req.URL.Scheme != target.Scheme ||
		req.URL.Path[:index+4] != target.Path {
		return false
	}
	return true
}
