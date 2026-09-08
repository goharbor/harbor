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

package main

import (
	"context"
	"errors"
	"testing"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	usertesting "github.com/goharbor/harbor/src/testing/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestUpdateInitPasswordWithMgr_NewUser(t *testing.T) {
	ctx := context.Background()
	mockMgr := usertesting.NewManager(t)

	// User without salt (new user needing password initialization)
	mockMgr.On("Get", ctx, 1).Return(&commonmodels.User{
		UserID: 1,
		Salt:   "",
	}, nil)
	mockMgr.On("UpdatePassword", ctx, 1, "newpassword").Return(nil)

	err := updateInitPasswordWithMgr(ctx, mockMgr, 1, "newpassword")
	assert.NoError(t, err)
}

func TestUpdateInitPasswordWithMgr_ExistingUser(t *testing.T) {
	ctx := context.Background()
	mockMgr := usertesting.NewManager(t)

	// User with existing salt (password already set)
	mockMgr.On("Get", ctx, 1).Return(&commonmodels.User{
		UserID: 1,
		Salt:   "existingsalt",
	}, nil)

	err := updateInitPasswordWithMgr(ctx, mockMgr, 1, "")
	assert.NoError(t, err)
}

func TestUpdateInitPasswordWithMgr_AdminPasswordIgnored(t *testing.T) {
	ctx := context.Background()
	mockMgr := usertesting.NewManager(t)

	// Admin user (ID=1) with existing password, config password provided
	// This should trigger the warning log
	mockMgr.On("Get", ctx, adminUserID).Return(&commonmodels.User{
		UserID: adminUserID,
		Salt:   "existingsalt",
	}, nil)

	err := updateInitPasswordWithMgr(ctx, mockMgr, adminUserID, "configpassword")
	assert.NoError(t, err)
}

func TestUpdateInitPasswordWithMgr_NonAdminExistingPassword(t *testing.T) {
	ctx := context.Background()
	mockMgr := usertesting.NewManager(t)

	// Non-admin user with existing password - no warning should be logged
	mockMgr.On("Get", ctx, 2).Return(&commonmodels.User{
		UserID: 2,
		Salt:   "existingsalt",
	}, nil)

	err := updateInitPasswordWithMgr(ctx, mockMgr, 2, "somepassword")
	assert.NoError(t, err)
}

func TestUpdateInitPasswordWithMgr_GetUserError(t *testing.T) {
	ctx := context.Background()
	mockMgr := usertesting.NewManager(t)

	mockMgr.On("Get", ctx, 1).Return(nil, errors.New("database error"))

	err := updateInitPasswordWithMgr(ctx, mockMgr, 1, "password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
}

func TestUpdateInitPasswordWithMgr_UpdatePasswordError(t *testing.T) {
	ctx := context.Background()
	mockMgr := usertesting.NewManager(t)

	mockMgr.On("Get", ctx, 1).Return(&commonmodels.User{
		UserID: 1,
		Salt:   "",
	}, nil)
	mockMgr.On("UpdatePassword", ctx, 1, "newpassword").Return(errors.New("update failed"))

	err := updateInitPasswordWithMgr(ctx, mockMgr, 1, "newpassword")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update user encrypted password")
}
