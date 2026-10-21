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
	"context"
	"net/http"
	"testing"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	pkgoidc "github.com/goharbor/harbor/src/pkg/oidc"
	testingUser "github.com/goharbor/harbor/src/testing/controller/user"
)

func TestIDToken(t *testing.T) {
	idToken := &idToken{}

	// not the OIDC mode
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	ctx := idToken.Generate(req)
	assert.Nil(t, ctx)

	// contains no authorization header
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req = req.WithContext(lib.WithAuthMode(req.Context(), common.OIDCAuth))
	ctx = idToken.Generate(req)
	assert.Nil(t, ctx)

	// contains no authorization header
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/service/token?service=harbor-registry&scope=repository:foo/bar:pull", nil)
	require.Nil(t, err)
	req = req.WithContext(lib.WithAuthMode(req.Context(), common.OIDCAuth))
	ctx = idToken.Generate(req)
	assert.Nil(t, ctx)
}

func TestIDTokenLoginGroupsBlocked(t *testing.T) {
	config.InitWithSettings(map[string]any{
		common.OIDCLoginGroups: "allowed-group",
	})
	t.Cleanup(func() {
		config.InitWithSettings(map[string]any{common.OIDCLoginGroups: ""})
	})

	origVerifyFn := idTokenVerifyFn
	idTokenVerifyFn = func(_ context.Context, _ string) (*gooidc.IDToken, error) {
		return &gooidc.IDToken{Issuer: "test-issuer", Subject: "test-subject"}, nil
	}
	t.Cleanup(func() { idTokenVerifyFn = origVerifyFn })

	testCtl := &testingUser.Controller{}
	testCtl.On("GetBySubIss", mock.Anything, "test-subject", "test-issuer").Return(
		&models.User{Username: "blockedUser"}, nil)
	origUserCtl := idTokenUserCtl
	idTokenUserCtl = testCtl
	t.Cleanup(func() { idTokenUserCtl = origUserCtl })

	origUserInfoFn := idTokenUserInfoFn
	idTokenUserInfoFn = func(_ context.Context, _ *pkgoidc.Token, _ cfgModels.OIDCSetting) (*pkgoidc.UserInfo, error) {
		return &pkgoidc.UserInfo{}, nil
	}
	t.Cleanup(func() { idTokenUserInfoFn = origUserInfoFn })

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req = req.WithContext(lib.WithAuthMode(req.Context(), common.OIDCAuth))
	req.Header.Set("Authorization", "Bearer fake-token")

	ctx := (&idToken{}).Generate(req)
	assert.Nil(t, ctx)
}
