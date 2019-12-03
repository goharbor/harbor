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
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils"
	"io/ioutil"
	"k8s.io/api/authentication/v1beta1"
	"net/http"
	"net/http/httptest"
	"strings"
)

type userEntry struct {
	username     string
	password     string
	sessionId    string
	reviewStatus string
}

type authHandler struct {
	entries []userEntry
	m       map[string]string
}

var reviewStatusTpl = `{"apiVersion": "authentication.k8s.io/v1beta1", "kind": "TokenReview", "status": {"authenticated": true, "user": {"username": "%s", "groups": ["vsphere.local\\users", "vsphere.local\\administrators", "vsphere.local\\caadmins", "vsphere.local\\systemconfiguration.bashshelladministrators", "vsphere.local\\systemconfiguration.administrators", "vsphere.local\\licenseservice.administrators", "vsphere.local\\everyone"], "extra": {"method": ["basic"]}}}}`

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
		for _, e := range ah.entries {
			if e.username == strings.ToLower(u) {
				_, err := rw.Write([]byte(fmt.Sprintf(`{"session_id": "%s"}`, e.sessionId)))
				if err != nil {
					panic(err)
				} else {
					return
				}
			}
		}
		http.Error(rw, fmt.Sprintf("Do not find entry in entrylist, username: %s", u), http.StatusUnauthorized)
	}
}

type reviewTokenHandler struct {
	entries []userEntry
}

func (rth *reviewTokenHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "", http.StatusMethodNotAllowed)
	}
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to read request body, error: %v", err), http.StatusBadRequest)
	}
	reviewData := &v1beta1.TokenReview{}
	if err := json.Unmarshal(bodyBytes, reviewData); err != nil {
		http.Error(rw, fmt.Sprintf("failed to decode request body, error: %v", err), http.StatusBadRequest)
	}
	defer req.Body.Close()
	for _, e := range rth.entries {
		if reviewData.Spec.Token == e.sessionId {
			_, err := rw.Write([]byte(fmt.Sprintf(reviewStatusTpl, e.username)))
			if err != nil {
				panic(err)
			} else {
				return
			}
		}
	}
	http.Error(rw, fmt.Sprintf("failed to match token: %s, entrylist: %+v", reviewData.Spec.Token, rth.entries), http.StatusUnauthorized)
}

// NewMockServer creates the mock server for testing
func NewMockServer(creds map[string]string) *httptest.Server {
	mux := http.NewServeMux()
	entryList := []userEntry{}
	for user, pwd := range creds {
		e := userEntry{
			username:     strings.ToLower(user),
			password:     pwd,
			sessionId:    utils.GenerateRandomString(),
			reviewStatus: fmt.Sprintf(reviewStatusTpl, user),
		}
		entryList = append(entryList, e)
	}
	mux.Handle("/test/login", &authHandler{m: creds, entries: entryList})
	mux.Handle("/test/tokenreview", &reviewTokenHandler{entries: entryList})
	return httptest.NewTLSServer(mux)
}
