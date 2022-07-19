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

package azurecr

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var (
	mockURL      = "https://test.azurecr.io"
	mockUsername = "user"
	mockPassword = "password"
	mockToken    = "test-token"
)

func TestAuth(t *testing.T) {
	// mock server
	defer gock.Off()
	// mock v2 API
	gock.New(mockURL).
		Get("/v2/").
		Reply(401).
		SetHeader("Www-Authenticate", `Bearer realm="https://test.azurecr.io/oauth2/token",service="test.azurecr.io"`)
	// mock token API
	gock.New(mockURL).
		Get("/oauth2/token").
		MatchParam("service", "test.azurecr.io").
		MatchParam("scope", `repository:library/busybox:metadata_read`).
		BasicAuth(mockUsername, mockPassword).
		Reply(200).
		JSON(fmt.Sprintf(`{"access_token": "%s"}`, mockToken))

	a := newAuthorizer(&model.Registry{URL: mockURL, Credential: &model.Credential{AccessKey: mockUsername, AccessSecret: mockPassword}})
	ct := &http.Client{}
	a.client = ct
	gock.InterceptClient(ct)

	// test authorize
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", mockURL, "/v2/library/busybox/tags/list"), nil)
	assert.NoError(t, err)
	err = a.Modify(req)
	assert.NoError(t, err)
	// check whether set bearer token
	tokenHeader := req.Header.Get("Authorization")
	assert.Equal(t, fmt.Sprintf("Bearer %s", mockToken), tokenHeader)
}
