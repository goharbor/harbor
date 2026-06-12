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

package migration

import (
	"testing"

	"github.com/stretchr/testify/suite"
	htesting "github.com/goharbor/harbor/src/testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/q"
	pat "github.com/goharbor/harbor/src/pkg/pat"
	oidcdao "github.com/goharbor/harbor/src/pkg/oidc/dao"
)

type MigrateCliSecretsTestSuite struct {
	htesting.Suite
	userCtl user.Controller
	patMgr  pat.Manager
}

func (suite *MigrateCliSecretsTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.ClearTables = []string{"personal_access_token", "oidc_user", "harbor_user"}
	suite.userCtl = user.Ctl
	suite.patMgr = pat.NewManager()
}

func (suite *MigrateCliSecretsTestSuite) TestMigrateCliSecretsBasic() {
	ctx := suite.Context()

	// Create a test user
	u := &models.User{
		Username: "oidcuser",
		Email:    "oidc@example.com",
		Realname: "OIDC User",
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)

	// Encrypt a CLI secret
	plainSecret := "test-cli-secret-1234567890"
	secretKey, err := config.SecretKey()
	suite.NoError(err)
	encryptedSecret, err := utils.ReversibleEncrypt(plainSecret, secretKey)
	suite.NoError(err)

	// Create OIDC user record with encrypted secret directly in DB
	oidcDAO := oidcdao.NewMetaDao()
	oidcU := &models.OIDCUser{
		UserID:  int(uid),
		Secret:  encryptedSecret,
		SubIss:  "test-sub|test-issuer",
		Token:   "test-token",
	}
	_, err = oidcDAO.Create(ctx, oidcU)
	suite.NoError(err)

	// Run migration
	err = MigrateCliSecretsToLegacyPATs(ctx)
	suite.NoError(err)

	// Verify that a legacy PAT was created
	query := q.New(q.KeyWords{"user_id": int(uid), "name": "cli-secret"})
	pats, err := suite.patMgr.List(ctx, query)
	suite.NoError(err)
	suite.Equal(1, len(pats))

	pat := pats[0]
	suite.Equal(int(uid), pat.UserID)
	suite.Equal("cli-secret", pat.Name)
	suite.True(pat.IsLegacy)
	suite.Equal(int64(-1), pat.ExpiresAt)
	suite.NotEmpty(pat.Secret)
	suite.NotEmpty(pat.Salt)
}

func (suite *MigrateCliSecretsTestSuite) TestMigrateCliSecretsIdempotent() {
	ctx := suite.Context()

	// Create a test user
	u := &models.User{
		Username: "idempotentuser",
		Email:    "idempotent@example.com",
		Realname: "Idempotent User",
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)

	// Encrypt and create OIDC user
	plainSecret := "idempotent-secret-1234567890"
	secretKey, err := config.SecretKey()
	suite.NoError(err)
	encryptedSecret, err := utils.ReversibleEncrypt(plainSecret, secretKey)
	suite.NoError(err)

	oidcDAO := oidcdao.NewMetaDao()
	oidcU := &models.OIDCUser{
		UserID:  int(uid),
		Secret:  encryptedSecret,
		SubIss:  "idempotent-sub|idempotent-issuer",
		Token:   "idempotent-token",
	}
	_, err = oidcDAO.Create(ctx, oidcU)
	suite.NoError(err)

	// Run migration first time
	err = MigrateCliSecretsToLegacyPATs(ctx)
	suite.NoError(err)

	// Get the first migration result
	query := q.New(q.KeyWords{"user_id": int(uid), "name": "cli-secret"})
	pats1, err := suite.patMgr.List(ctx, query)
	suite.NoError(err)
	suite.Equal(1, len(pats1))

	// Run migration again
	err = MigrateCliSecretsToLegacyPATs(ctx)
	suite.NoError(err)

	// Verify only one legacy PAT exists (idempotent)
	pats2, err := suite.patMgr.List(ctx, query)
	suite.NoError(err)
	suite.Equal(1, len(pats2))
	suite.Equal(pats1[0].ID, pats2[0].ID)
}

func (suite *MigrateCliSecretsTestSuite) TestMigrateCliSecretsMultipleUsers() {
	ctx := suite.Context()

	// Create multiple test users
	for i := 1; i <= 3; i++ {
		u := &models.User{
			Username: "user" + string(rune(48+i)),
			Email:    "user" + string(rune(48+i)) + "@example.com",
			Realname: "User " + string(rune(48+i)),
		}
		uid, err := suite.userCtl.Create(ctx, u)
		suite.NoError(err)

		// Create OIDC secret for each user
		plainSecret := "secret-" + string(rune(48+i))
		secretKey, err := config.SecretKey()
		suite.NoError(err)
		encryptedSecret, err := utils.ReversibleEncrypt(plainSecret, secretKey)
		suite.NoError(err)

		oidcDAO := oidcdao.NewMetaDao()
		oidcU := &models.OIDCUser{
			UserID:  int(uid),
			Secret:  encryptedSecret,
			SubIss:  "sub" + string(rune(48+i)) + "|issuer",
			Token:   "token" + string(rune(48+i)),
		}
		_, err = oidcDAO.Create(ctx, oidcU)
		suite.NoError(err)
	}

	// Run migration
	err := MigrateCliSecretsToLegacyPATs(ctx)
	suite.NoError(err)

	// Verify all users have legacy PATs
	allPATs, err := suite.patMgr.List(ctx, q.New(q.KeyWords{"name": "cli-secret"}))
	suite.NoError(err)
	suite.Equal(3, len(allPATs))

	for _, token := range allPATs {
		suite.Equal("cli-secret", token.Name)
		suite.True(token.IsLegacy)
	}
}

func (suite *MigrateCliSecretsTestSuite) TestMigrateCliSecretsNoExistingSecrets() {
	ctx := suite.Context()

	// Create a user without OIDC secret
	u := &models.User{
		Username: "nosecretuser",
		Email:    "nosecret@example.com",
		Realname: "No Secret User",
	}
	_, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)

	// Run migration (should not error)
	err = MigrateCliSecretsToLegacyPATs(ctx)
	suite.NoError(err)

	// Verify no PATs were created
	allPATs, err := suite.patMgr.List(ctx, q.New(q.KeyWords{}))
	suite.NoError(err)
	suite.Equal(0, len(allPATs))
}

func TestMigrateCliSecretsTestSuite(t *testing.T) {
	suite.Run(t, new(MigrateCliSecretsTestSuite))
}
