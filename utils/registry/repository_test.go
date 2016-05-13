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

package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	//"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry/auth"
	"github.com/vmware/harbor/utils/registry/errors"
)

var (
	username         string  = "user"
	password         string  = "P@ssw0rd"
	repo             string  = "samalba/my-app"
	tags             tagResp = tagResp{Tags: []string{"1.0", "2.0", "3.0"}}
	validToken       string  = "valid_token"
	invalidToken     string  = "invalid_token"
	credential       auth.Credential
	registryServer   *httptest.Server
	tokenServer      *httptest.Server
	repositoryClient *Repository
)

type tagResp struct {
	Tags []string `json:"tags"`
}

func TestMain(m *testing.M) {
	//log.SetLevel(log.DebugLevel)
	credential = auth.NewBasicAuthCredential(username, password)

	tokenServer = initTokenServer()
	defer tokenServer.Close()

	registryServer = initRegistryServer()
	defer registryServer.Close()

	os.Exit(m.Run())
}

func initRegistryServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", servePing)
	mux.HandleFunc(fmt.Sprintf("/v2/%s/tags/list", repo), serveTaglisting)

	return httptest.NewServer(mux)
}

//response ping request: http://registry/v2
func servePing(w http.ResponseWriter, r *http.Request) {
	if !isTokenValid(r) {
		challenge(w)
		return
	}
}

func serveTaglisting(w http.ResponseWriter, r *http.Request) {
	if !isTokenValid(r) {
		challenge(w)
		return
	}

	if err := json.NewEncoder(w).Encode(tags); err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func isTokenValid(r *http.Request) bool {
	valid := false
	auth := r.Header.Get(http.CanonicalHeaderKey("Authorization"))
	if len(auth) != 0 {
		auth = strings.TrimSpace(auth)
		index := strings.Index(auth, "Bearer")
		token := auth[index+6:]
		token = strings.TrimSpace(token)
		if token == validToken {
			valid = true
		}
	}
	return valid
}

func challenge(w http.ResponseWriter) {
	challenge := "Bearer realm=\"" + tokenServer.URL + "/service/token\",service=\"token-service\""
	w.Header().Set("Www-Authenticate", challenge)
	w.WriteHeader(http.StatusUnauthorized)
	return
}

func initTokenServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/service/token", serveToken)

	return httptest.NewServer(mux)
}

func serveToken(w http.ResponseWriter, r *http.Request) {
	u, p, ok := r.BasicAuth()
	if !ok || u != username || p != password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	result := make(map[string]interface{})
	result["token"] = validToken
	result["expires_in"] = 300
	result["issued_at"] = time.Now().Format(time.RFC3339)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func TestListTag(t *testing.T) {
	client, err := NewRepositoryWithCredential(repo, registryServer.URL, credential)
	if err != nil {
		t.Error(err)
	}

	list, err := client.ListTag()
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) != len(tags.Tags) {
		t.Errorf("expected length: %d, actual length: %d", len(tags.Tags), len(list))
		return
	}

}

func TestListTagWithInvalidCredential(t *testing.T) {
	credential := auth.NewBasicAuthCredential(username, "wrong_password")
	client, err := NewRepositoryWithCredential(repo, registryServer.URL, credential)
	if err != nil {
		t.Error(err)
	}

	_, err = client.ListTag()
	if err != nil {
		e, ok := errors.ParseError(err)
		if ok && e.StatusCode == http.StatusUnauthorized {
			return
		}
		t.Error(err)
		return
	}
}
