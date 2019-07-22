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
	"net/http"
	"net/http/httptest"
)

type authHandler struct {
	m map[string]string
}

// ServeHTTP handles HTTP requests
func (ah *authHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "", http.StatusMethodNotAllowed)
	}
	if u, p, ok := req.BasicAuth(); !ok {
		// Simulate a service error
		http.Error(rw, "", http.StatusInternalServerError)
	} else if pass, ok := ah.m[u]; !ok || pass != p {
		http.Error(rw, "", http.StatusUnauthorized)
	} else {
		_, e := rw.Write([]byte(`{"session_id": "hgx59wuWI3b0jcbtidv5mU1YCp-DOQ9NKR1iYKACdKCvbVn7"}`))
		if e != nil {
			panic(e)
		}
	}
}

type reviewTokenHandler struct {
}

func (rth *reviewTokenHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "", http.StatusMethodNotAllowed)
	}
	rw.Write([]byte(`{"apiVersion": "authentication.k8s.io/v1beta1", "kind": "TokenReview", "status": {"authenticated": true, "user": {"username": "administrator@vsphere.local", "groups": ["vsphere.local\\users", "vsphere.local\\administrators", "vsphere.local\\caadmins", "vsphere.local\\systemconfiguration.bashshelladministrators", "vsphere.local\\systemconfiguration.administrators", "vsphere.local\\licenseservice.administrators", "vsphere.local\\everyone"], "extra": {"method": ["basic"]}}}}`))
}

// NewMockServer creates the mock server for testing
func NewMockServer(creds map[string]string) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("/test/login", &authHandler{m: creds})
	mux.Handle("/test/tokenreview", &reviewTokenHandler{})
	return httptest.NewTLSServer(mux)
}
