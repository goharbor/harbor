// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	//"github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	registry_error "github.com/vmware/harbor/src/common/utils/registry/error"
	token_util "github.com/vmware/harbor/src/ui/service/token"
)

const (
	latency int = 10 //second, the network latency when token is received
)

type scope struct {
	Type    string
	Name    string
	Actions []string
}

func (s *scope) string() string {
	return fmt.Sprintf("%s:%s:%s", s.Type, s.Name, strings.Join(s.Actions, ","))
}

type tokenGenerator func(realm, service string, scopes []string) (token string, expiresIn int, issuedAt *time.Time, err error)

// Implements interface Authorizer
type tokenAuthorizer struct {
	scope     *scope
	tg        tokenGenerator
	cache     string     // cached token
	expiresAt *time.Time // The UTC standard time at when the token will expire
	sync.Mutex
}

// Scheme returns the scheme that the handler can handle
func (t *tokenAuthorizer) Scheme() string {
	return "bearer"
}

// AuthorizeRequest will add authorization header which contains a token before the request is sent
func (t *tokenAuthorizer) Authorize(req *http.Request, params map[string]string) error {
	var scopes []*scope
	var token string

	hasFrom := false
	from := req.URL.Query().Get("from")
	if len(from) != 0 {
		s := &scope{
			Type:    "repository",
			Name:    from,
			Actions: []string{"pull"},
		}
		scopes = append(scopes, s)
		// do not cache the token if "from" appears
		hasFrom = true
	}

	if t.scope != nil {
		scopes = append(scopes, t.scope)
	}

	expired := true

	cachedToken, cachedExpiredAt := t.getCachedToken()

	if len(cachedToken) != 0 && cachedExpiredAt != nil {
		expired = cachedExpiredAt.Before(time.Now().UTC())
	}

	if expired || hasFrom {
		scopeStrs := []string{}
		for _, scope := range scopes {
			scopeStrs = append(scopeStrs, scope.string())
		}
		to, expiresIn, _, err := t.tg(params["realm"], params["service"], scopeStrs)
		if err != nil {
			return err
		}
		token = to

		if !hasFrom {
			t.updateCachedToken(to, expiresIn)
		}
	} else {
		token = cachedToken
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", token))

	return nil
}

func (t *tokenAuthorizer) getCachedToken() (string, *time.Time) {
	t.Lock()
	defer t.Unlock()
	return t.cache, t.expiresAt
}

func (t *tokenAuthorizer) updateCachedToken(token string, expiresIn int) {
	t.Lock()
	defer t.Unlock()
	t.cache = token
	n := (time.Duration)(expiresIn - latency)
	e := time.Now().Add(n * time.Second).UTC()
	t.expiresAt = &e
}

// Implements interface Authorizer
type standardTokenAuthorizer struct {
	tokenAuthorizer
	client               *http.Client
	credential           Credential
	tokenServiceEndpoint string
}

// NewStandardTokenAuthorizer returns a standard token authorizer. The authorizer will request a token
// from token server and add it to the origin request
// If tokenServiceEndpoint is set, the token request will be sent to it instead of the server get from authorizer
// The usage please refer to the function tokenURL
func NewStandardTokenAuthorizer(credential Credential, insecure bool,
	tokenServiceEndpoint string, scopeType, scopeName string, scopeActions ...string) Authorizer {
	authorizer := &standardTokenAuthorizer{
		client: &http.Client{
			Transport: registry.GetHTTPTransport(insecure),
			Timeout:   30 * time.Second,
		},
		credential:           credential,
		tokenServiceEndpoint: tokenServiceEndpoint,
	}

	if len(scopeType) != 0 || len(scopeName) != 0 {
		authorizer.scope = &scope{
			Type:    scopeType,
			Name:    scopeName,
			Actions: scopeActions,
		}
	}

	authorizer.tg = authorizer.generateToken

	return authorizer
}

func (s *standardTokenAuthorizer) generateToken(realm, service string, scopes []string) (token string, expiresIn int, issuedAt *time.Time, err error) {
	realm = s.tokenURL(realm)

	u, err := url.Parse(realm)
	if err != nil {
		return
	}
	q := u.Query()
	q.Add("service", service)
	for _, scope := range scopes {
		q.Add("scope", scope)
	}
	u.RawQuery = q.Encode()
	r, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return
	}

	if s.credential != nil {
		s.credential.AddAuthorization(r)
	}

	resp, err := s.client.Do(r)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = &registry_error.Error{
			StatusCode: resp.StatusCode,
			Detail:     string(b),
		}
		return
	}

	tk := struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
		IssuedAt  string `json:"issued_at"`
	}{}
	if err = json.Unmarshal(b, &tk); err != nil {
		return
	}

	token = tk.Token

	expiresIn = tk.ExpiresIn

	if len(tk.IssuedAt) != 0 {
		t, err := time.Parse(time.RFC3339, tk.IssuedAt)
		if err != nil {
			log.Errorf("error occurred while parsing issued_at: %v", err)
			err = nil
		} else {
			issuedAt = &t
		}
	}

	return
}

// when the registry client is used inside Harbor, the token request
// can be posted to token service directly rather than going through nginx.
// If realm is set as the internal url of token service, this can resolve
// two problems:
// 1. performance issue
// 2. the realm field returned by registry is an IP which can not reachable
// inside Harbor
func (s *standardTokenAuthorizer) tokenURL(realm string) string {
	if len(s.tokenServiceEndpoint) != 0 {
		return s.tokenServiceEndpoint
	}
	return realm
}

// Implements interface Handler
type usernameTokenAuthorizer struct {
	tokenAuthorizer
	username string
}

// NewRegistryUsernameTokenAuthorizer returns an authorizer to generate token for registry according to
// the user's privileges
func NewRegistryUsernameTokenAuthorizer(username, scopeType, scopeName string, scopeActions ...string) Authorizer {
	return newUsernameTokenAuthorizer(false, username, scopeType, scopeName, scopeActions...)
}

// NewNotaryUsernameTokenAuthorizer returns an authorizer to generate token for notary according to
// the user's privileges
func NewNotaryUsernameTokenAuthorizer(username, scopeType, scopeName string, scopeActions ...string) Authorizer {
	return newUsernameTokenAuthorizer(true, username, scopeType, scopeName, scopeActions...)
}

// newUsernameTokenAuthorizer returns a authorizer which will generate a token according to
// the user's privileges
func newUsernameTokenAuthorizer(notary bool, username, scopeType, scopeName string, scopeActions ...string) Authorizer {
	authorizer := &usernameTokenAuthorizer{
		username: username,
	}

	authorizer.scope = &scope{
		Type:    scopeType,
		Name:    scopeName,
		Actions: scopeActions,
	}
	if notary {
		authorizer.tg = authorizer.genNotaryToken
	} else {
		authorizer.tg = authorizer.genRegistryToken
	}
	return authorizer
}

func (u *usernameTokenAuthorizer) generateToken(realm, service string, scopes []string) (token string, expiresIn int, issuedAt *time.Time, err error) {
	token, expiresIn, issuedAt, err = token_util.RegistryTokenForUI(u.username, service, scopes)
	return
}

func (u *usernameTokenAuthorizer) genRegistryToken(realm, service string, scopes []string) (token string, expiresIn int, issuedAt *time.Time, err error) {
	token, expiresIn, issuedAt, err = token_util.RegistryTokenForUI(u.username, service, scopes)
	return
}

func (u *usernameTokenAuthorizer) genNotaryToken(realm, service string, scopes []string) (token string, expiresIn int, issuedAt *time.Time, err error) {
	token, expiresIn, issuedAt, err = token_util.NotaryTokenForUI(u.username, service, scopes)
	return
}
