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

package test

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
	"strings"
)

// MockServerConfig ...
type MockServerConfig struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

type tokenHandler struct {
	clientID     string
	clientSecret string
	username     string
	password     string
}

func currPath() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Dir(f)
}

func (t *tokenHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	u, p, ok := req.BasicAuth()
	if !ok || u != t.clientID || p != t.clientSecret {
		http.Error(rw, "invalid client id/secret in header", http.StatusUnauthorized)
		return
	}
	gt := req.FormValue("grant_type")
	if gt == "password" {
		reqUsername := req.FormValue("username")
		reqPasswd := req.FormValue("password")
		if reqUsername == t.username && reqPasswd == t.password {
			serveToken(rw)
		} else {
			http.Error(rw, fmt.Sprintf("invalid username/password %s/%s", html.EscapeString(reqUsername), html.EscapeString(reqPasswd)), http.StatusUnauthorized)
		}
	} else if gt == "client_credentials" {
		serveToken(rw)
	} else {
		http.Error(rw, fmt.Sprintf("invalid grant_type: %s", html.EscapeString(gt)), http.StatusBadRequest)
		return
	}
}

func serveToken(rw http.ResponseWriter) {
	serveJSONFile(rw, "uaa-token.json")
}

func serveJSONFile(rw http.ResponseWriter, filename string) {
	data, err := ioutil.ReadFile(path.Join(currPath(), filename))
	if err != nil {
		panic(err)
	}
	rw.Header().Add("Content-Type", "application/json")
	_, err2 := rw.Write(data)
	if err2 != nil {
		panic(err2)
	}
}

type userInfoHandler struct {
	token string
}

func (u *userInfoHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	v := req.Header.Get("authorization")
	prefix := v[0:7]
	reqToken := v[7:]
	if strings.ToLower(prefix) != "bearer " || reqToken != u.token {
		http.Error(rw, "invalid token", http.StatusUnauthorized)
		return
	}
	serveJSONFile(rw, "./user-info.json")
}

type searchUserHandler struct {
	token string
}

func (su *searchUserHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	v := req.Header.Get("authorization")
	if v == "" {
		v = req.Header.Get("Authorization")
	}
	prefix := v[0:7]
	reqToken := v[7:]
	if strings.ToLower(prefix) != "bearer " || reqToken != su.token {
		http.Error(rw, "invalid token", http.StatusUnauthorized)
		return
	}
	f := req.URL.Query().Get("filter")
	elements := strings.Split(f, " ")
	if len(elements) == 3 {
		if elements[0] == "Username" && elements[1] == "eq" {
			if elements[2] == "'one'" {
				serveJSONFile(rw, "one-user.json")
				return
			}
			serveJSONFile(rw, "no-user.json")
			return
		}
		http.Error(rw, "invalid request", http.StatusBadRequest)
		return
	}
	http.Error(rw, html.EscapeString(fmt.Sprintf("Invalid request, elements: %v", elements)), http.StatusBadRequest)
}

// NewMockServer ...
func NewMockServer(cfg *MockServerConfig) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("/oauth/token", &tokenHandler{
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.Username,
		cfg.Password,
	})
	token, err := ioutil.ReadFile(path.Join(currPath(), "./good-access-token.txt"))
	if err != nil {
		panic(err)
	}
	mux.Handle("/userinfo", &userInfoHandler{strings.TrimSpace(string(token))})
	mux.Handle("/Users", &searchUserHandler{strings.TrimSpace(string(token))})
	return httptest.NewTLSServer(mux)
}
