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

package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
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
	if gt := req.FormValue("grant_type"); gt != "password" {
		http.Error(rw, fmt.Sprintf("invalid grant_type: %s", gt), http.StatusBadRequest)
		return
	}
	reqUsername := req.FormValue("username")
	reqPasswd := req.FormValue("password")
	if reqUsername == t.username && reqPasswd == t.password {
		token, err := ioutil.ReadFile(path.Join(currPath(), "./uaa-token.json"))
		if err != nil {
			panic(err)
		}
		_, err2 := rw.Write(token)
		if err2 != nil {
			panic(err2)
		}
	} else {
		http.Error(rw, fmt.Sprintf("invalid username/password %s/%s", reqUsername, reqPasswd), http.StatusUnauthorized)
	}
}

// NewMockServer ...
func NewMockServer(cfg *MockServerConfig) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("/uaa/oauth/token", &tokenHandler{
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.Username,
		cfg.Password,
	})
	return httptest.NewTLSServer(mux)
}
