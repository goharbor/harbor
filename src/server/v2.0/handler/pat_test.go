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

package handler

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	htesting "github.com/goharbor/harbor/src/testing"

	"github.com/goharbor/harbor/src/common/models"
	patctl "github.com/goharbor/harbor/src/controller/pat"
	userctl "github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/q"
	patmodel "github.com/goharbor/harbor/src/pkg/pat/model"
)

type PATHandlerTestSuite struct {
	htesting.Suite
	patCtl  patctl.Controller
	userCtl userctl.Controller
}

func (suite *PATHandlerTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.ClearTables = []string{"personal_access_token", "harbor_user"}
	suite.patCtl = patctl.Ctl
	suite.userCtl = userctl.Ctl
}

func (suite *PATHandlerTestSuite) createTestUser(username string) int {
	ctx := suite.Context()
	u := &models.User{
		Username: username,
		Email:    fmt.Sprintf("%s@example.com", username),
		Realname: strings.ToTitle(username),
	}
	uid, err := suite.userCtl.Create(ctx, u)
	suite.NoError(err)
	return int(uid)
}

func (suite *PATHandlerTestSuite) TestCreatePersonalAccessToken() {
	ctx := suite.Context()
	userID := suite.createTestUser("testuser")

	token := &patmodel.PersonalAccessToken{
		UserID:      userID,
		Name:        "test-token",
		Description: "Test token",
		ExpiresAt:   time.Now().AddDate(0, 0, 30).Unix(),
	}

	id, plainSecret, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)
	suite.True(id > 0)
	suite.NotEmpty(plainSecret)
	suite.Contains(plainSecret, "hbr_pat_")
}

func (suite *PATHandlerTestSuite) TestListPersonalAccessTokens() {
	ctx := suite.Context()
	userID := suite.createTestUser("listuser")

	// Create 3 PATs
	for i := 1; i <= 3; i++ {
		token := &patmodel.PersonalAccessToken{
			UserID: userID,
			Name:   fmt.Sprintf("token-%d", i),
		}
		_, _, err := suite.patCtl.Create(ctx, token)
		suite.NoError(err)
	}

	// List tokens
	query := q.New(q.KeyWords{"user_id": userID})
	tokens, err := suite.patCtl.List(ctx, query)
	suite.NoError(err)
	suite.Equal(3, len(tokens))

	for _, token := range tokens {
		suite.Equal(userID, token.UserID)
		suite.Empty(token.Secret) // Secret should not be exposed in list
		suite.Empty(token.Salt)
	}
}

func (suite *PATHandlerTestSuite) TestGetPersonalAccessToken() {
	ctx := suite.Context()
	userID := suite.createTestUser("getuser")

	// Create a PAT
	token := &patmodel.PersonalAccessToken{
		UserID:      userID,
		Name:        "get-token",
		Description: "Get test token",
	}
	id, _, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Get token
	retrieved, err := suite.patCtl.Get(ctx, id)
	suite.NoError(err)
	suite.Equal(id, retrieved.ID)
	suite.Equal(userID, retrieved.UserID)
	suite.Equal("get-token", retrieved.Name)
	suite.Equal("Get test token", retrieved.Description)
}

func (suite *PATHandlerTestSuite) TestUpdatePersonalAccessToken() {
	ctx := suite.Context()
	userID := suite.createTestUser("updateuser")

	// Create a PAT
	token := &patmodel.PersonalAccessToken{
		UserID:      userID,
		Name:        "update-token",
		Description: "Original",
		Disabled:    false,
	}
	id, _, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Update token
	token.ID = id
	token.Description = "Updated"
	token.Disabled = true

	err = suite.patCtl.Update(ctx, token)
	suite.NoError(err)

	// Verify update
	updated, err := suite.patCtl.Get(ctx, id)
	suite.NoError(err)
	suite.Equal("Updated", updated.Description)
	suite.True(updated.Disabled)
}

func (suite *PATHandlerTestSuite) TestDeletePersonalAccessToken() {
	ctx := suite.Context()
	userID := suite.createTestUser("deleteuser")

	// Create a PAT
	token := &patmodel.PersonalAccessToken{
		UserID: userID,
		Name:   "delete-token",
	}
	id, _, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Delete token
	err = suite.patCtl.Delete(ctx, id)
	suite.NoError(err)

	// Verify deletion
	_, err = suite.patCtl.Get(ctx, id)
	suite.Error(err)
}

func (suite *PATHandlerTestSuite) TestRefreshPersonalAccessTokenSecret() {
	ctx := suite.Context()
	userID := suite.createTestUser("refreshuser")

	// Create a PAT
	token := &patmodel.PersonalAccessToken{
		UserID: userID,
		Name:   "refresh-token",
	}
	id, originalSecret, err := suite.patCtl.Create(ctx, token)
	suite.NoError(err)

	// Refresh secret
	newSecret, err := suite.patCtl.RefreshSecret(ctx, id, "")
	suite.NoError(err)

	// Verify new secret is different
	suite.NotEqual(originalSecret, newSecret)
	suite.Contains(newSecret, "hbr_pat_")
}

func (suite *PATHandlerTestSuite) TestCountPersonalAccessTokens() {
	ctx := suite.Context()
	userID := suite.createTestUser("countuser")

	// Create 2 PATs
	for i := 1; i <= 2; i++ {
		token := &patmodel.PersonalAccessToken{
			UserID: userID,
			Name:   fmt.Sprintf("token-%d", i),
		}
		_, _, err := suite.patCtl.Create(ctx, token)
		suite.NoError(err)
	}

	// Count tokens
	query := q.New(q.KeyWords{"user_id": userID})
	count, err := suite.patCtl.Count(ctx, query)
	suite.NoError(err)
	suite.Equal(int64(2), count)
}

func TestPATHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PATHandlerTestSuite))
}
