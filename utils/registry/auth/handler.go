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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	token_util "github.com/vmware/harbor/service/token"
	"github.com/vmware/harbor/utils/log"
	registry_errors "github.com/vmware/harbor/utils/registry/errors"
)

const (
	// credential type
	basicAuth string = "basic_auth"
	secretKey string = "secret_key"
)

// Handler authorizes the request when encounters a 401 error
type Handler interface {
	// Schema : basic, bearer
	Schema() string
	//AuthorizeRequest adds basic auth or token auth to the header of request
	AuthorizeRequest(req *http.Request, params map[string]string) error
}

// Credential ...
type Credential interface {
	// AddAuthorization adds authorization information to request
	AddAuthorization(req *http.Request)
}

type basicAuthCredential struct {
	username string
	password string
}

// NewBasicAuthCredential ...
func NewBasicAuthCredential(username, password string) Credential {
	return &basicAuthCredential{
		username: username,
		password: password,
	}
}

func (b *basicAuthCredential) AddAuthorization(req *http.Request) {
	req.SetBasicAuth(b.username, b.password)
}

type token struct {
	Token string `json:"token"`
}

type standardTokenHandler struct {
	client     *http.Client
	credential Credential
}

// NewStandardTokenHandler returns a standard token handler. The handler will request a token
// from token server whose URL is specified in the "WWW-authentication" header and add it to
// the origin request
// TODO deal with https
func NewStandardTokenHandler(credential Credential) Handler {
	return &standardTokenHandler{
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
		credential: credential,
	}
}

// Schema implements the corresponding method in interface AuthHandler
func (t *standardTokenHandler) Schema() string {
	return "bearer"
}

// AuthorizeRequest implements the corresponding method in interface AuthHandler
func (t *standardTokenHandler) AuthorizeRequest(req *http.Request, params map[string]string) error {
	realm, ok := params["realm"]
	if !ok {
		return errors.New("no realm")
	}

	service := params["service"]
	scope := params["scope"]

	u, err := url.Parse(realm)
	if err != nil {
		return err
	}

	q := u.Query()
	q.Add("service", service)

	for _, s := range strings.Split(scope, " ") {
		q.Add("scope", s)
	}

	u.RawQuery = q.Encode()

	r, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	t.credential.AddAuthorization(r)

	resp, err := t.client.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return registry_errors.Error{
			StatusCode: resp.StatusCode,
			Message:    string(b),
		}
	}

	decoder := json.NewDecoder(resp.Body)

	tk := &token{}
	if err = decoder.Decode(tk); err != nil {
		return err
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", tk.Token))

	log.Debugf("standardTokenHandler generated token successfully | %s %s", req.Method, req.URL)

	return nil
}

type usernameTokenHandler struct {
	username string
}

// NewUsernameTokenHandler returns a handler which will generate
// a token according the user's privileges
func NewUsernameTokenHandler(username string) Handler {
	return &usernameTokenHandler{
		username: username,
	}
}

// Schema implements the corresponding method in interface AuthHandler
func (u *usernameTokenHandler) Schema() string {
	return "bearer"
}

// AuthorizeRequest implements the corresponding method in interface AuthHandler
func (u *usernameTokenHandler) AuthorizeRequest(req *http.Request, params map[string]string) error {
	service := params["service"]

	scopes := []string{}
	scope := params["scope"]
	if len(scope) != 0 {
		scopes = strings.Split(scope, " ")
	}

	token, err := token_util.GenTokenForUI(u.username, service, scopes)
	if err != nil {
		return err
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", token))

	log.Debugf("usernameTokenHandler generated token successfully | %s %s", req.Method, req.URL)

	return nil
}
