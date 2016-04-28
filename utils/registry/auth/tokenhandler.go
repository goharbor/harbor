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

type token struct {
	token     string
	expiresIn time.Time
}

type tokenGenerator func(realm, service string, scopes []string) (*token, error)

type tokenHandler struct {
	scope *scope
	cache map[string]*token
	tg    tokenGenerator
}

// Schema returns the schema that the handler can handle
func (t *tokenHandler) Schema() string {
	return "bearer"
}

// AuthorizeRequest will add authorization header which contains a token before the request is sent
func (t *tokenHandler) AuthorizeRequest(req *http.Request, params map[string]string) error {
	var token string
	var scopes []*scope

	// TODO handle additional scope: xxx.xxx.xxx?from=repo

	scopes = append(scopes, t.scope)
	key := cacheKey(scopes)

	value, ok := t.cache[key]
	var expired bool

	if ok {
		expired = value.expiresIn.Before(time.Now())
	}

	if ok && !expired {
		token = value.token
		log.Debugf("get token from cache: %s", key)
	} else {
		if ok && expired {
			delete(t.cache, key)
			log.Debugf("token is expired, remove from cache: %s", key)
		}

		scopeStrs := []string{}
		for _, scope := range scopes {
			scopeStrs = append(scopeStrs, scope.string())
		}
		tk, err := t.tg(params["realm"], params["service"], scopeStrs)
		if err != nil {
			return err
		}
		token = tk.token
		t.cache[key] = tk
		log.Debugf("add token to cache: %s", key)
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", token))
	log.Debugf("add token to request: %s %s", req.Method, req.URL.String())

	return nil
}

// cacheKey returns a string which can identify the scope array and can be used as the key in cache map
func cacheKey(scopes []*scope) string {
	key := ""
	for _, scope := range scopes {
		key = key + scope.string() + "|"
	}
	key = strings.TrimRight(key, "|")

	return key
}

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
	handler.cache = make(map[string]*token, 1)
	handler.tg = handler.generateToken

	return handler
}

func (s *standardTokenHandler) generateToken(realm, service string, scopes []string) (*token, error) {
	u, err := url.Parse(realm)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Add("service", service)
	for _, scope := range scopes {
		q.Add("scope", scope)
	}
	u.RawQuery = q.Encode()
	r, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	s.credential.AddAuthorization(r)

	resp, err := s.client.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, registry_errors.Error{
			StatusCode: resp.StatusCode,
			Message:    string(b),
		}
	}

	tk := struct {
		Token string `json:"token"`
	}{}
	if err = json.Unmarshal(b, &tk); err != nil {
		return nil, err
	}

	t := &token{
		token: tk.Token,
		// TODO handle the expires time
		expiresIn: time.Now().Add(5 * time.Minute),
	}

	log.Debug("get token from token server")

	return t, nil
}

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
	handler.cache = make(map[string]*token, 1)

	handler.tg = handler.generateToken

	return handler
}

func (u *usernameTokenHandler) generateToken(realm, service string, scopes []string) (*token, error) {
	tk, err := token_util.GenTokenForUI(u.username, service, scopes)
	if err != nil {
		return nil, err
	}

	t := &token{
		token: tk,
		// TODO handle the expires time
		expiresIn: time.Now().Add(5 * time.Minute),
	}

	log.Debug("get token by calling GenTokenForUI directly")

	return t, nil
}
