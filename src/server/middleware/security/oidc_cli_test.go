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
	"fmt"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/oidc"
	testingUser "github.com/goharbor/harbor/src/testing/controller/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOIDCCli(t *testing.T) {
	oidcCli := &oidcCli{}
	// not the candidate request
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/v2.0/users/", nil)
	require.Nil(t, err)
	ctx := oidcCli.Generate(req)
	assert.Nil(t, ctx)

	// the auth mode isn't OIDC
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/service/token", nil)
	require.Nil(t, err)
	ctx = oidcCli.Generate(req)
	assert.Nil(t, ctx)

	// pass
	username := "oidcModiferTester"
	password := "oidcSecret"
	testCtl := &testingUser.Controller{}
	testCtl.On("GetByName", mock.Anything, username).Return(
		&models.User{
			Username: username,
			Email:    fmt.Sprintf("%s@test.domain", username),
		}, nil)
	uctl = testCtl
	oidc.SetHardcodeVerifierForTest(password)
	req = req.WithContext(lib.WithAuthMode(req.Context(), common.OIDCAuth))
	req.SetBasicAuth(username, password)
	ctx = oidcCli.Generate(req)
	assert.NotNil(t, ctx)
}

func TestOIDCCliValid(t *testing.T) {
	oc := &oidcCli{}
	req1, _ := http.NewRequest(http.MethodPost, "https://test.goharbor.io/api/v2.0/projects", nil)
	req2, _ := http.NewRequest(http.MethodGet, "https://test.goharbor.io/api/v2.0/projects?name=test", nil)
	req3, _ := http.NewRequest(http.MethodGet, "https://test.goharbor.io/api/v2.0/projects/library/repositories/", nil)
	req4, _ := http.NewRequest(http.MethodGet, "https://test.goharbor.io/api/v2.0/projects/library/repositories/ubuntu/artifacts", nil)
	req5, _ := http.NewRequest(http.MethodGet, "https://test.goharbor.io/api/v2.0/systeminfo", nil)
	req6, _ := http.NewRequest(http.MethodGet, "https://test.goharbor.io/api/version", nil)
	req7, _ := http.NewRequest(http.MethodGet, "https://test.goharbor.io/api/v2.0/labels?scope=g", nil)
	req8, _ := http.NewRequest(http.MethodDelete, "https://test.goharbor.io/api/v2.0/projects/library/repositories/ubuntu/artifacts/sha256:xxxxx/tags/v14.04", nil)
	req9, _ := http.NewRequest(http.MethodPut, "https://test.goharbor.io/api/v2.0/projects/library/repositories/ubuntu", nil)
	req10, _ := http.NewRequest(http.MethodGet, "https://test.goharbor.io/api/v2.0/projects/library/repositores/ubuntu/artifacts/sha256:xxxx/tags", nil)
	req11, _ := http.NewRequest(http.MethodGet, "https://test.goharbor.io/api/v2.0/projects/library/repositories/ubuntu", nil)

	cases := []struct {
		r     *http.Request
		valid bool
	}{
		{req1, true},
		{req2, true},
		{req3, true},
		{req4, true},
		{req5, true},
		{req6, true},
		{req7, true},
		{req8, true},
		{req9, false},
		{req10, false},
		{req11, false},
	}

	for _, c := range cases {
		assert.Equal(t, c.valid, oc.valid(c.r), fmt.Sprintf("Failed. path: %s, method: %s, expected: %v", c.r.URL.Path, c.r.Method, c.valid))
	}

}
