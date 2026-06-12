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
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	htesting "github.com/goharbor/harbor/src/testing"

	"github.com/goharbor/harbor/src/common/models"
	patctl "github.com/goharbor/harbor/src/controller/pat"
	userctl "github.com/goharbor/harbor/src/controller/user"
	patmodel "github.com/goharbor/harbor/src/pkg/pat/model"
)

type PATSecurityTestSuite struct {
	htesting.Suite
	patCtl  patctl.Controller
	userCtl userctl.Controller
}

func (suite *PATSecurityTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.ClearTables = []string{"personal_access_token", "harbor_user"}
	suite.patCtl = patctl.Ctl
	suite.userCtl = userctl.Ctl
}

func (suite *PATSecurityTestSuite) TestGenerateWithValidPAT() {
	ctx := suite.Context()

	// Create a test user
	u := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Realname: "Test User",
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)
	u.UserID = int(uid)

	// Create a PAT
	token := &patmodel.PersonalAccessToken{
		UserID:      int(uid),
		Name:        "test-token",
		ExpiresAt:   time.Now().AddDate(0, 0, 30).Unix(),
	}
	patID, plainSecret, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Create HTTP request with Basic Auth using PAT
	req, err := http.NewRequest("GET", "http://localhost", nil)
	suite.NoError(err)
	req.SetBasicAuth("testuser", plainSecret)

	// Generate should succeed and return valid context
	generator := &pat{}
	secCtx := generator.Generate(req)
	suite.NotNil(secCtx, "should generate security context for valid PAT")

	// Verify user context
	retrievedToken, err := suite.patCtl.Get(ctx, patID)
	suite.NoError(err)
	suite.NotNil(retrievedToken)
}

func (suite *PATSecurityTestSuite) TestGenerateWithExpiredPAT() {
	ctx := suite.Context()

	// Create a test user
	u := &models.User{
		Username: "expireduser",
		Email:    "expired@example.com",
		Realname: "Expired User",
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)

	// Create an expired PAT
	token := &patmodel.PersonalAccessToken{
		UserID:    int(uid),
		Name:      "expired-token",
		ExpiresAt: time.Now().AddDate(0, 0, -1).Unix(), // Expired yesterday
	}
	_, plainSecret, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Create HTTP request
	req, err := http.NewRequest("GET", "http://localhost", nil)
	suite.NoError(err)
	req.SetBasicAuth("expireduser", plainSecret)

	// Generate should return nil for expired token
	generator := &pat{}
	secCtx := generator.Generate(req)
	suite.Nil(secCtx, "should not generate context for expired PAT")
}

func (suite *PATSecurityTestSuite) TestGenerateWithDisabledPAT() {
	ctx := suite.Context()

	// Create a test user
	u := &models.User{
		Username: "disableduser",
		Email:    "disabled@example.com",
		Realname: "Disabled User",
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)

	// Create a disabled PAT
	token := &patmodel.PersonalAccessToken{
		UserID:    int(uid),
		Name:      "disabled-token",
		ExpiresAt: -1,
		Disabled:  true,
	}
	_, plainSecret, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Create HTTP request
	req, err := http.NewRequest("GET", "http://localhost", nil)
	suite.NoError(err)
	req.SetBasicAuth("disableduser", plainSecret)

	// Generate should return nil for disabled token
	generator := &pat{}
	secCtx := generator.Generate(req)
	suite.Nil(secCtx, "should not generate context for disabled PAT")
}

func (suite *PATSecurityTestSuite) TestGenerateWithInvalidSecret() {
	ctx := suite.Context()

	// Create a test user
	u := &models.User{
		Username: "invaliduser",
		Email:    "invalid@example.com",
		Realname: "Invalid User",
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)

	// Create a PAT
	token := &patmodel.PersonalAccessToken{
		UserID:    int(uid),
		Name:      "test-token",
		ExpiresAt: -1,
	}
	_, _, err = suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Create HTTP request with wrong secret
	req, err := http.NewRequest("GET", "http://localhost", nil)
	suite.NoError(err)
	wrongSecret := "hbr_pat_wrongsecretvalue1234567890"
	req.SetBasicAuth("invaliduser", wrongSecret)

	// Generate should return nil for invalid secret
	generator := &pat{}
	secCtx := generator.Generate(req)
	suite.Nil(secCtx, "should not generate context for invalid PAT secret")
}

func (suite *PATSecurityTestSuite) TestGenerateWithoutPATPrefix() {
	// Create HTTP request without hbr_pat_ prefix
	req, err := http.NewRequest("GET", "http://localhost", nil)
	suite.NoError(err)
	req.SetBasicAuth("someuser", "somepasswordwithoutprefix")

	// Generate should return nil (not a PAT)
	generator := &pat{}
	secCtx := generator.Generate(req)
	suite.Nil(secCtx, "should not generate context for request without hbr_pat_ prefix")
}

func (suite *PATSecurityTestSuite) TestGenerateWithoutBasicAuth() {
	// Create HTTP request without Basic Auth
	req, err := http.NewRequest("GET", "http://localhost", nil)
	suite.NoError(err)

	// Generate should return nil
	generator := &pat{}
	secCtx := generator.Generate(req)
	suite.Nil(secCtx, "should not generate context for request without Basic Auth")
}

func (suite *PATSecurityTestSuite) TestGenerateWithNeverExpiresPAT() {
	ctx := suite.Context()

	// Create a test user
	u := &models.User{
		Username: "neverexpireuser",
		Email:    "neverexpire@example.com",
		Realname: "Never Expire User",
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)

	// Create a PAT that never expires
	token := &patmodel.PersonalAccessToken{
		UserID:    int(uid),
		Name:      "never-expire-token",
		ExpiresAt: -1, // Never expires
	}
	_, plainSecret, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Create HTTP request
	req, err := http.NewRequest("GET", "http://localhost", nil)
	suite.NoError(err)
	req.SetBasicAuth("neverexpireuser", plainSecret)

	// Generate should succeed
	generator := &pat{}
	secCtx := generator.Generate(req)
	suite.NotNil(secCtx, "should generate context for PAT that never expires")
}

func (suite *PATSecurityTestSuite) TestGenerateWithBasicAuthHeaderFormat() {
	ctx := suite.Context()

	// Create a test user
	u := &models.User{
		Username: "headeruser",
		Email:    "header@example.com",
		Realname: "Header User",
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)

	// Create a PAT
	token := &patmodel.PersonalAccessToken{
		UserID:    int(uid),
		Name:      "test-token",
		ExpiresAt: -1,
	}
	_, plainSecret, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Create HTTP request with Authorization header
	req, err := http.NewRequest("GET", "http://localhost", nil)
	suite.NoError(err)
	credentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("headeruser:%s", plainSecret)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", credentials))

	// Generate should succeed
	generator := &pat{}
	secCtx := generator.Generate(req)
	suite.NotNil(secCtx, "should generate context from Authorization header")
}

func TestPATSecurityTestSuite(t *testing.T) {
	suite.Run(t, new(PATSecurityTestSuite))
}
