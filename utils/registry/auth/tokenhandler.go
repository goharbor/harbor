/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	token_util "github.com/vmware/harbor/service/token"
	"github.com/vmware/harbor/utils/log"
	registry_errors "github.com/vmware/harbor/utils/registry/errors"
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

// Implements interface Handler
type tokenHandler struct {
	scope     *scope
	tg        tokenGenerator
	cache     string     // cached token
	expiresIn int        // The duration in seconds since the token was issued that it will remain valid
	issuedAt  *time.Time // The RFC3339-serialized UTC standard time at which a given token was issued
}

// Scheme returns the scheme that the handler can handle
func (t *tokenHandler) Scheme() string {
	return "bearer"
}

// AuthorizeRequest will add authorization header which contains a token before the request is sent
func (t *tokenHandler) AuthorizeRequest(req *http.Request, params map[string]string) error {
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

	scopes = append(scopes, t.scope)

	expired := true

	if t.expiresIn != 0 && t.issuedAt != nil {
		expired = t.issuedAt.Add(time.Duration(t.expiresIn) * time.Second).Before(time.Now().UTC())
	}

	if expired || hasFrom {
		scopeStrs := []string{}
		for _, scope := range scopes {
			scopeStrs = append(scopeStrs, scope.string())
		}
		to, expiresIn, issuedAt, err := t.tg(params["realm"], params["service"], scopeStrs)
		if err != nil {
			return err
		}
		token = to

		if !hasFrom {
			t.cache = token
			t.expiresIn = expiresIn
			t.issuedAt = issuedAt
			log.Debug("add token to cache")
		}
	} else {
		token = t.cache
		log.Debug("get token from cache")
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", token))
	log.Debugf("add token to request: %s %s", req.Method, req.URL.String())

	return nil
}

// Implements interface Handler
type standardTokenHandler struct {
	tokenHandler
	client     *http.Client
	credential Credential
}

// NewStandardTokenHandler returns a standard token handler. The handler will request a token
// from token server and add it to the origin request
// TODO deal with https
func NewStandardTokenHandler(credential Credential, scopeType, scopeName string, scopeActions ...string) Handler {
	handler := &standardTokenHandler{
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
		credential: credential,
	}

	handler.scope = &scope{
		Type:    scopeType,
		Name:    scopeName,
		Actions: scopeActions,
	}
	handler.tg = handler.generateToken

	return handler
}

func (s *standardTokenHandler) generateToken(realm, service string, scopes []string) (token string, expiresIn int, issuedAt *time.Time, err error) {
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

	s.credential.AddAuthorization(r)

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
		err = registry_errors.Error{
			StatusCode: resp.StatusCode,
			Message:    string(b),
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

	t, err := time.Parse(time.RFC3339, tk.IssuedAt)
	if err != nil {
		log.Errorf("error occurred while parsing issued_at: %v", err)
		err = nil
	} else {
		issuedAt = &t
	}

	log.Debug("get token from token server")

	return
}

// Implements interface Handler
type usernameTokenHandler struct {
	tokenHandler
	username string
}

// NewUsernameTokenHandler returns a handler which will generate a token according to
// the user's privileges
func NewUsernameTokenHandler(username string, scopeType, scopeName string, scopeActions ...string) Handler {
	handler := &usernameTokenHandler{
		username: username,
	}

	handler.scope = &scope{
		Type:    scopeType,
		Name:    scopeName,
		Actions: scopeActions,
	}

	handler.tg = handler.generateToken

	return handler
}

func (u *usernameTokenHandler) generateToken(realm, service string, scopes []string) (token string, expiresIn int, issuedAt *time.Time, err error) {
	token, expiresIn, issuedAt, err = token_util.GenTokenForUI(u.username, service, scopes)
	log.Debug("get token by calling GenTokenForUI directly")
	return
}
