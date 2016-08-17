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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/docker/distribution/registry/client/auth"
)

func TestNewAuthorizerStore(t *testing.T) {
	server := newRegistryServer()
	defer server.Close()

	_, err := NewAuthorizerStore(server.URL, false, nil)
	if err != nil {
		t.Fatalf("failed to create authorizer store: %v", err)
	}
}

type simpleAuthorizer struct {
}

func (s *simpleAuthorizer) Scheme() string {
	return "bearer"
}

func (s *simpleAuthorizer) Authorize(req *http.Request,
	params map[string]string) error {
	req.Header.Set("Authorization", "Bearer token")
	return nil
}

func TestModify(t *testing.T) {
	authorizer := &simpleAuthorizer{}
	challenge := auth.Challenge{
		Scheme: "bearer",
	}

	as := &AuthorizerStore{
		authorizers: []Authorizer{authorizer},
		challenges:  []auth.Challenge{challenge},
	}

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if err = as.Modify(req); err != nil {
		t.Fatalf("failed to modify request: %v", err)
	}

	header := req.Header.Get("Authorization")
	if len(header) == 0 {
		t.Fatal("\"Authorization\" header not found")
	}

	if !strings.HasPrefix(header, "Bearer") {
		t.Fatal("\"Authorization\" header does not start with \"Bearer\"")
	}
}

func newRegistryServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", handlePing)

	return httptest.NewServer(mux)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	challenge := "Bearer realm=\"https://auth.docker.io/token\",service=\"registry.docker.io\""
	w.Header().Set("Www-Authenticate", challenge)
	w.WriteHeader(http.StatusUnauthorized)
}
