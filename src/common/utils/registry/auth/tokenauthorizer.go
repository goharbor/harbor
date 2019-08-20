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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	token_util "github.com/goharbor/harbor/src/core/service/token"
)

const (
	latency int = 10 // second, the network latency when token is received
	scheme      = "bearer"
)

type tokenGenerator interface {
	generate(scopes []*token.ResourceActions, endpoint string) (*models.Token, error)
}

// UserAgentModifier adds the "User-Agent" header to the request
type UserAgentModifier struct {
	UserAgent string
}

// Modify adds user-agent header to the request
func (u *UserAgentModifier) Modify(req *http.Request) error {
	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), u.UserAgent)
	return nil
}

// tokenAuthorizer implements registry.Modifier interface. It parses scopses
// from the request, generates authentication token and modifies the requset
// by adding the token
type tokenAuthorizer struct {
	registryURL  *url.URL // used to filter request
	generator    tokenGenerator
	client       *http.Client
	cachedTokens map[string]*models.Token
	sync.Mutex
}

// add token to the request
func (t *tokenAuthorizer) Modify(req *http.Request) error {
	// only handle requests sent to registry
	goon, err := t.filterReq(req)
	if err != nil {
		return err
	}

	if !goon {
		log.Debugf("the request %s is not sent to registry, skip", req.URL.String())
		return nil
	}

	// parse scopes from request
	scopes, err := parseScopes(req)
	if err != nil {
		return err
	}

	var token *models.Token
	// try to get token from cache if the request is for empty scope(login)
	// or single scope
	if len(scopes) <= 1 {
		key := ""
		if len(scopes) == 1 {
			key = scopeString(scopes[0])
		}
		token = t.getCachedToken(key)
	}

	// request a new token if the token is null
	if token == nil {
		token, err = t.generator.generate(scopes, t.registryURL.String())
		if err != nil {
			return err
		}
		// if the token is null(this happens if the registry needs no authentication), return
		// directly. Or the token will be cached
		if token == nil {
			return nil
		}
		// only cache the token for empty scope(login) or single scope request
		if len(scopes) <= 1 {
			key := ""
			if len(scopes) == 1 {
				key = scopeString(scopes[0])
			}
			t.updateCachedToken(key, token)
		}
	}

	tk := token.GetToken()
	if len(tk) == 0 {
		return errors.New("empty token content")
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", tk))

	return nil
}

func scopeString(scope *token.ResourceActions) string {
	if scope == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s:%s", scope.Type, scope.Name, strings.Join(scope.Actions, ","))
}

// some requests are sent to backend storage, such as s3, this method filters
// the requests only sent to registry
func (t *tokenAuthorizer) filterReq(req *http.Request) (bool, error) {
	// the registryURL is nil when the first request comes, init it with
	// the scheme and host of the request which must be sent to the registry
	if t.registryURL == nil {
		u, err := url.Parse(buildPingURL(req.URL.Scheme + "://" + req.URL.Host))
		if err != nil {
			return false, err
		}
		t.registryURL = u
	}

	v2Index := strings.Index(req.URL.Path, "/v2/")
	if v2Index == -1 {
		return false, nil
	}

	if req.URL.Host != t.registryURL.Host || req.URL.Scheme != t.registryURL.Scheme ||
		req.URL.Path[:v2Index+4] != t.registryURL.Path {
		return false, nil
	}

	return true, nil
}

// parse scopes from the request according to its method, path and query string
func parseScopes(req *http.Request) ([]*token.ResourceActions, error) {
	scopes := []*token.ResourceActions{}

	from := req.URL.Query().Get("from")
	if len(from) != 0 {
		scopes = append(scopes, &token.ResourceActions{
			Type:    "repository",
			Name:    from,
			Actions: []string{"pull"},
		})
	}

	var scope *token.ResourceActions
	path := strings.TrimRight(req.URL.Path, "/")
	repository := parseRepository(path)
	if len(repository) > 0 {
		// pull, push, delete blob/manifest
		scope = &token.ResourceActions{
			Type: "repository",
			Name: repository,
		}
		switch req.Method {
		case http.MethodGet, http.MethodHead:
			scope.Actions = []string{"pull"}
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			scope.Actions = []string{"pull", "push"}
		case http.MethodDelete:
			scope.Actions = []string{"*"}
		default:
			scope = nil
			log.Warningf("unsupported method: %s", req.Method)
		}
	} else if catalog.MatchString(path) {
		// catalog
		scope = &token.ResourceActions{
			Type:    "registry",
			Name:    "catalog",
			Actions: []string{"*"},
		}
	} else if base.MatchString(path) {
		// base
		scope = nil
	} else {
		// unknown
		return scopes, fmt.Errorf("can not parse scope from the request: %s %s", req.Method, req.URL.Path)
	}

	if scope != nil {
		scopes = append(scopes, scope)
	}

	strs := []string{}
	for _, s := range scopes {
		strs = append(strs, scopeString(s))
	}
	log.Debugf("scopes parsed from request: %s", strings.Join(strs, " "))

	return scopes, nil
}

func (t *tokenAuthorizer) getCachedToken(scope string) *models.Token {
	t.Lock()
	defer t.Unlock()
	token := t.cachedTokens[scope]
	if token == nil {
		return nil
	}

	issueAt, err := time.Parse(time.RFC3339, token.IssuedAt)
	if err != nil {
		log.Errorf("failed parse %s: %v", token.IssuedAt, err)
		delete(t.cachedTokens, scope)
		return nil
	}

	if issueAt.Add(time.Duration(token.ExpiresIn-latency) * time.Second).Before(time.Now().UTC()) {
		delete(t.cachedTokens, scope)
		return nil
	}

	log.Debugf("get token for scope %s from cache", scope)
	return token
}

func (t *tokenAuthorizer) updateCachedToken(scope string, token *models.Token) {
	t.Lock()
	defer t.Unlock()
	t.cachedTokens[scope] = token
}

// ping returns the realm, service and error
func ping(client *http.Client, endpoint string) (string, string, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	challenges := ParseChallengeFromResponse(resp)
	for _, challenge := range challenges {
		if scheme == challenge.Scheme {
			realm := challenge.Parameters["realm"]
			service := challenge.Parameters["service"]
			return realm, service, nil
		}
	}

	log.Warningf("Schemas %v are unsupported", challenges)
	return "", "", nil
}

// NewStandardTokenAuthorizer returns a standard token authorizer. The authorizer will request a token
// from token server and add it to the origin request
// If customizedTokenService is set, the token request will be sent to it instead of the server get from authorizer
func NewStandardTokenAuthorizer(client *http.Client, credential Credential,
	customizedTokenService ...string) modifier.Modifier {
	generator := &standardTokenGenerator{
		credential: credential,
		client:     client,
	}

	// when the registry client is used inside Harbor, the token request
	// can be posted to token service directly rather than going through nginx.
	// If realm is set as the internal url of token service, this can resolve
	// two problems:
	// 1. performance issue
	// 2. the realm field returned by registry is an IP which can not reachable
	// inside Harbor
	if len(customizedTokenService) > 0 && len(customizedTokenService[0]) > 0 {
		generator.realm = customizedTokenService[0]
	}

	return &tokenAuthorizer{
		cachedTokens: make(map[string]*models.Token),
		generator:    generator,
		client:       client,
	}
}

// standardTokenGenerator implements interface tokenGenerator
type standardTokenGenerator struct {
	realm      string
	service    string
	credential Credential
	client     *http.Client
}

// get token from token service
func (s *standardTokenGenerator) generate(scopes []*token.ResourceActions, endpoint string) (*models.Token, error) {
	// ping first if the realm or service is null
	if len(s.realm) == 0 || len(s.service) == 0 {
		realm, service, err := ping(s.client, endpoint)
		if err != nil {
			return nil, err
		}
		if len(realm) == 0 {
			log.Warning("empty realm, skip")
			return nil, nil
		}
		if len(s.realm) == 0 {
			s.realm = realm
		}
		s.service = service
	}

	return getToken(s.client, s.credential, s.realm, s.service, scopes)
}

// NewRawTokenAuthorizer returns a token authorizer which calls method to create
// token directly
func NewRawTokenAuthorizer(username, service string) modifier.Modifier {
	generator := &rawTokenGenerator{
		service:  service,
		username: username,
	}

	return &tokenAuthorizer{
		cachedTokens: make(map[string]*models.Token),
		generator:    generator,
	}
}

// rawTokenGenerator implements interface tokenGenerator
type rawTokenGenerator struct {
	service  string
	username string
}

// generate token directly
func (r *rawTokenGenerator) generate(scopes []*token.ResourceActions, endpoint string) (*models.Token, error) {
	return token_util.MakeToken(r.username, r.service, scopes)
}

func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}
