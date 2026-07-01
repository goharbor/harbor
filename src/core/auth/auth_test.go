// Copyright 2018 Project Harbor Authors
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
package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	testinguserpkg "github.com/goharbor/harbor/src/testing/pkg/user"
)

var l = NewUserLock(2 * time.Second)

func TestLock(t *testing.T) {
	t.Log("Locking john")
	l.Lock("john")
	if !l.IsLocked("john") {
		t.Errorf("John should be locked")
	}
	t.Log("Locking jack")
	l.Lock("jack")
	t.Log("Sleep for 2 seconds and check...")
	time.Sleep(2 * time.Second)
	if l.IsLocked("jack") {
		t.Errorf("After 2 seconds, jack shouldn't be locked")
	}
	if l.IsLocked("daniel") {
		t.Errorf("daniel has never been locked, he should not be locked")
	}
}

func TestDefaultAuthenticate(t *testing.T) {
	authHelper := DefaultAuthenticateHelper{}
	m := models.AuthModel{}
	user, err := authHelper.Authenticate(context.TODO(), m)
	if user != nil || err == nil {
		t.Fatal("Default implementation should return nil")
	}
}

func TestDefaultOnBoardUser(t *testing.T) {
	user := &models.User{}
	authHelper := DefaultAuthenticateHelper{}
	err := authHelper.OnBoardUser(context.TODO(), user)
	if err == nil {
		t.Fatal("Default implementation should return error")
	}
}

func TestDefaultMethods(t *testing.T) {
	authHelper := DefaultAuthenticateHelper{}
	_, err := authHelper.SearchUser(context.TODO(), "sample")
	if err == nil {
		t.Fatal("Default implementation should return error")
	}

	_, err = authHelper.SearchGroup(context.TODO(), "sample")
	if err == nil {
		t.Fatal("Default implementation should return error")
	}

	err = authHelper.OnBoardGroup(context.TODO(), &model.UserGroup{}, "sample")
	if err == nil {
		t.Fatal("Default implementation should return error")
	}
}

func TestErrAuth(t *testing.T) {
	assert := assert.New(t)
	e := NewErrAuth("test")
	expectedStr := "Failed to authenticate user, due to error 'test'"
	assert.Equal(expectedStr, e.Error())
}

func TestCanUseOIDCAuth(t *testing.T) {
	// Test all code paths of canUseOIDCAuth with mocked user manager.
	// Saves and restores the original user.Mgr to avoid side effects.
	tests := []struct {
		name        string
		username    string
		mockUser    *models.User
		mockErr     error
		expected    bool
		description string
	}{
		{
			name:        "DBLookupError",
			username:    "alice",
			mockUser:    nil,
			mockErr:     fmt.Errorf("database connection failed"),
			expected:    false,
			description: "User lookup error should return false (fallback to DB auth)",
		},
		{
			name:        "UserNotFound",
			username:    "bob",
			mockUser:    nil,
			mockErr:     nil,
			expected:    false,
			description: "User not found (nil return, no error) should return false",
		},
		{
			name:     "UserWithoutOIDCMetadata",
			username: "charlie",
			mockUser: &models.User{
				UserID:       2,
				Username:    "charlie",
				OIDCUserMeta: nil,
			},
			mockErr:     nil,
			expected:    false,
			description: "User without OIDC metadata should return false (fallback to DB auth)",
		},
		{
			name:     "UserWithOIDCMetadata",
			username: "dave",
			mockUser: &models.User{
				UserID:       3,
				Username:    "dave",
				OIDCUserMeta: &models.OIDCUser{ID: 1, UserID: 3},
			},
			mockErr:     nil,
			expected:    true,
			description: "User with OIDC metadata should return true (use OIDC auth)",
		},
	}

	ctx := context.Background()
	orig := user.Mgr
	defer func() { user.Mgr = orig }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := &testinguserpkg.Manager{}
			user.Mgr = mockMgr

			mockMgr.On("GetByName", ctx, tt.username).Return(tt.mockUser, tt.mockErr)

			result := canUseOIDCAuth(ctx, tt.username)
			assert.Equal(t, tt.expected, result, tt.description)
			mockMgr.AssertExpectations(t)
		})
	}
}

func TestIsSuperUser_WithMockedUserManager(t *testing.T) {
	// Test IsSuperUser with mocked user.Mgr to exercise the user lookup path.
	// This also exercises the same GetByName call used by canUseOIDCAuth during Login flow.

	orig := user.Mgr
	defer func() { user.Mgr = orig }()

	ctx := context.Background()

	tests := []struct {
		name        string
		username    string
		mockUser    *models.User
		mockErr     error
		expected    bool
		description string
	}{
		{
			name:        "SuperUserIsAdmin",
			username:    "admin",
			mockUser:    &models.User{UserID: 1, Username: "admin"},
			mockErr:     nil,
			expected:    true,
			description: "User with UserID=1 should be super user",
		},
		{
			name:        "NonSuperUser",
			username:    "regular",
			mockUser:    &models.User{UserID: 2, Username: "regular"},
			mockErr:     nil,
			expected:    false,
			description: "User with UserID != 1 should not be super user",
		},
		{
			name:        "UserLookupError",
			username:    "unknown",
			mockUser:    nil,
			mockErr:     fmt.Errorf("db error"),
			expected:    false,
			description: "User lookup error should return false (not super user)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := &testinguserpkg.Manager{}
			user.Mgr = mockMgr

			mockMgr.On("GetByName", ctx, tt.username).Return(tt.mockUser, tt.mockErr)

			result := IsSuperUser(ctx, tt.username)
			assert.Equal(t, tt.expected, result, tt.description)
			mockMgr.AssertExpectations(t)
		})
	}
}
