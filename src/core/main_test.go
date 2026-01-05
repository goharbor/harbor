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
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/testing/mock"
	usermock "github.com/goharbor/harbor/src/testing/pkg/user"
)

func TestUpdateInitPassword_NewUser_SetsPassword(t *testing.T) {
	ctx := context.Background()
	mgr := &usermock.Manager{}

	// user with empty salt means password not yet set
	mgr.On("Get", mock.Anything, 1).Return(&models.User{
		UserID: 1,
		Salt:   "",
	}, nil)
	mgr.On("UpdatePassword", mock.Anything, 1, "testpassword").Return(nil)

	err := updateInitPasswordWithMgr(ctx, mgr, 1, "testpassword")

	assert.NoError(t, err)
	mgr.AssertExpectations(t)
}

func TestUpdateInitPassword_ExistingUser_SkipsUpdate(t *testing.T) {
	ctx := context.Background()
	mgr := &usermock.Manager{}

	// user with non-empty salt means password already set
	mgr.On("Get", mock.Anything, 1).Return(&models.User{
		UserID: 1,
		Salt:   "existingsalt",
	}, nil)

	err := updateInitPasswordWithMgr(ctx, mgr, 1, "newpassword")

	assert.NoError(t, err)
	mgr.AssertExpectations(t)
	mgr.AssertNotCalled(t, "UpdatePassword")
}

func TestUpdateInitPassword_ExistingUser_WarnsWhenConfigPasswordSet(t *testing.T) {
	ctx := context.Background()
	mgr := &usermock.Manager{}

	mgr.On("Get", mock.Anything, 1).Return(&models.User{
		UserID: 1,
		Salt:   "existingsalt",
	}, nil)

	// capture log output
	buf := &bytes.Buffer{}
	log.DefaultLogger().SetOutput(buf)
	defer log.DefaultLogger().SetOutput(os.Stdout)

	err := updateInitPasswordWithMgr(ctx, mgr, 1, "configpassword")

	assert.NoError(t, err)
	assert.True(t, strings.Contains(buf.String(), "will not be applied"),
		"expected warning about config password not being applied")
}

func TestUpdateInitPassword_ExistingUser_NoWarningWhenConfigPasswordEmpty(t *testing.T) {
	ctx := context.Background()
	mgr := &usermock.Manager{}

	mgr.On("Get", mock.Anything, 1).Return(&models.User{
		UserID: 1,
		Salt:   "existingsalt",
	}, nil)

	buf := &bytes.Buffer{}
	log.DefaultLogger().SetOutput(buf)
	defer log.DefaultLogger().SetOutput(os.Stdout)

	err := updateInitPasswordWithMgr(ctx, mgr, 1, "")

	assert.NoError(t, err)
	assert.False(t, strings.Contains(buf.String(), "will not be applied"),
		"should not warn when config password is empty")
}
