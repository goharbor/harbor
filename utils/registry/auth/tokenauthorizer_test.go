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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	token = "token"
)

func TestAuthorizeOfStandardTokenAuthorizer(t *testing.T) {
	tokenServer := newTokenServer()
	defer tokenServer.Close()

	authorizer := NewStandardTokenAuthorizer(nil, false, "repository", "library/ubuntu", "pull")
	req, err := http.NewRequest("GET", "http://registry", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	params := map[string]string{
		"realm": tokenServer.URL + "/token",
	}

	if err := authorizer.Authorize(req, params); err != nil {
		t.Fatalf("failed to authorize request: %v", err)
	}

	tk := req.Header.Get("Authorization")
	if tk != "Bearer "+token {
		t.Errorf("unexpected token: %s != %s", tk, "Bearer "+token)
	}
}

func TestSchemeOfStandardTokenAuthorizer(t *testing.T) {
	authorizer := &standardTokenAuthorizer{}
	if authorizer.Scheme() != "bearer" {
		t.Errorf("unexpected scheme: %s != %s", authorizer.Scheme(), "bearer")
	}

}

func newTokenServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", handleToken)

	return httptest.NewServer(mux)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	result := map[string]interface{}{}
	result["token"] = token
	result["expires_in"] = 300
	result["issued_at"] = time.Now().Format(time.RFC3339)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
