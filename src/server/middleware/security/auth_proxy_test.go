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

package security

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/goharbor/harbor/src/common"
	_ "github.com/goharbor/harbor/src/core/auth/authproxy"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/orm"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/authentication/v1beta1"
)

func TestAuthProxy(t *testing.T) {
	config.Init()
	authProxy := &authProxy{}

	server, err := newAuthProxyTestServer()
	require.Nil(t, err)
	defer server.Close()

	c := map[string]interface{}{
		common.HTTPAuthProxySkipSearch:          "true",
		common.HTTPAuthProxyVerifyCert:          "false",
		common.HTTPAuthProxyEndpoint:            "https://auth.proxy/suffix",
		common.HTTPAuthProxyTokenReviewEndpoint: server.URL,
		common.AUTHMode:                         common.HTTPAuth,
	}
	config.Upload(c)
	v, e := config.HTTPAuthProxySetting(orm.Context())
	require.Nil(t, e)
	assert.Equal(t, *v, models.HTTPAuthProxy{
		Endpoint:            "https://auth.proxy/suffix",
		SkipSearch:          true,
		VerifyCert:          false,
		TokenReviewEndpoint: server.URL,
		AdminGroups:         []string{},
		AdminUsernames:      []string{},
	})

	// No onboard
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/v2", nil)
	require.Nil(t, err)
	ormCtx := orm.Context()
	req = req.WithContext(lib.WithAuthMode(ormCtx, common.HTTPAuth))
	req.SetBasicAuth("tokenreview$administrator@vsphere.local", "reviEwt0k3n")
	ctx := authProxy.Generate(req)
	assert.NotNil(t, ctx)
}

// NewAuthProxyTestServer mocks a https server for auth proxy.
func newAuthProxyTestServer() (*httptest.Server, error) {
	const webhookPath = "/authproxy/tokenreview"

	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unexpected method: %v", r.Method), http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != webhookPath {
			http.Error(w, fmt.Sprintf("unexpected path: %v", r.URL.Path), http.StatusNotFound)
			return
		}

		var review v1beta1.TokenReview
		bodyData, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(bodyData, &review); err != nil {
			http.Error(w, fmt.Sprintf("failed to decode body: %v", err), http.StatusBadRequest)
			return
		}
		// ensure we received the serialized tokenreview as expected
		if review.APIVersion != "authentication.k8s.io/v1beta1" {
			http.Error(w, fmt.Sprintf("wrong api version: %s", string(bodyData)), http.StatusBadRequest)
			return
		}

		type userInfo struct {
			Username string              `json:"username"`
			UID      string              `json:"uid"`
			Groups   []string            `json:"groups"`
			Extra    map[string][]string `json:"extra"`
		}
		type status struct {
			Authenticated bool     `json:"authenticated"`
			User          userInfo `json:"user"`
			Audiences     []string `json:"audiences"`
		}

		var extra map[string][]string
		if review.Status.User.Extra != nil {
			extra = map[string][]string{}
			for k, v := range review.Status.User.Extra {
				extra[k] = v
			}
		}

		resp := struct {
			Kind       string `json:"kind"`
			APIVersion string `json:"apiVersion"`
			Status     status `json:"status"`
		}{
			Kind:       "TokenReview",
			APIVersion: v1beta1.SchemeGroupVersion.String(),
			Status: status{
				true,
				userInfo{
					Username: "administrator@vsphere.local",
					UID:      review.Status.User.UID,
					Groups:   review.Status.User.Groups,
					Extra:    extra,
				},
				review.Status.Audiences,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(serveHTTP))
	server.StartTLS()

	serverURL, _ := url.Parse(server.URL)
	serverURL.Path = webhookPath
	server.URL = serverURL.String()

	return server, nil
}
