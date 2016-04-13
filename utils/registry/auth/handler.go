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
	"net/http"
	"net/url"
	"strings"

	"github.com/astaxie/beego"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"
)

// Handler authorizes the request when encounters a 401 error
type Handler interface {
	// Schema : basic, bearer
	Schema() string
	//AuthorizeRequest adds basic auth or token auth to the header of request
	AuthorizeRequest(req *http.Request, params map[string]string) error
}

// Credential ...
type Credential struct {
	// Username ...
	Username string
	// Password ...
	Password string
	// SecretKey ...
	SecretKey string
}

type token struct {
	Token string `json:"token"`
}

type standardTokenHandler struct {
	client     *http.Client
	credential *Credential
}

// NewTokenHandler ...
// TODO deal with https
func NewStandardTokenHandler(credential *Credential) Handler {
	return &standardTokenHandler{
		client: &http.Client{
			Transport: http.DefaultTransport,
		},
		credential: credential,
	}
}

// Scheme : see interface AuthHandler
func (t *standardTokenHandler) Schema() string {
	return "bearer"
}

// AuthorizeRequest : see interface AuthHandler
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

	// TODO support secretKey
	if len(t.credential.Username) != 0 {
		r.SetBasicAuth(t.credential.Username, t.credential.Password)
	}

	resp, err := t.client.Do(r)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error occured when get token from %s, status code: %d, status info: %s",
			realm, resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	tk := &token{}
	if err = decoder.Decode(tk); err != nil {
		return err
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", tk.Token))

	log.Debugf("standardTokenHandler generated token successfully | %s %s", req.Method, req.URL)

	return nil
}

type sessionTokenHandler struct {
	sessionID string
}

func NewSessionTokenHandler(sessionID string) Handler {
	return &sessionTokenHandler{
		sessionID: sessionID,
	}
}

func (s *sessionTokenHandler) Schema() string {
	return "bearer"
}

func (s *sessionTokenHandler) AuthorizeRequest(req *http.Request, params map[string]string) error {
	session, err := beego.GlobalSessions.GetSessionStore(s.sessionID)
	if err != nil {
		return err
	}

	username, ok := session.Get("username").(string)
	if !ok {
		return errors.New("username in session can not be converted to string")
	}

	if len(username) == 0 {
		userID, ok := session.Get("userId").(int)
		if !ok {
			return errors.New("userId in session can not be converted to int")
		}

		u := models.User{
			UserID: userID,
		}
		user, err := dao.GetUser(u)
		if err != nil {
			return err
		}

		if user == nil {
			return fmt.Errorf("user with id %d does not exist", userID)
		}

		username = user.Username
	}

	service := params["service"]

	scopes := []string{}
	scope := params["scope"]
	if len(scope) != 0 {
		scopes = strings.Split(scope, " ")
	}

	token, err := utils.GenTokenForUI(username, service, scopes)
	if err != nil {
		return err
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", token))

	log.Debugf("sessionTokenHandler generated token successfully | %s %s", req.Method, req.URL)

	return nil
}
