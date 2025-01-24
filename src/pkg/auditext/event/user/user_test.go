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

package user

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/pkg/member/models"
	pkgUser "github.com/goharbor/harbor/src/pkg/user"
	mockUser "github.com/goharbor/harbor/src/testing/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
}

func (u *UserTestSuite) SetupTest() {
}

func (c *UserTestSuite) TestUserIDToName() {
	// Replace the real user manager with the mock
	realUserMgr := pkgUser.Mgr
	defer func() {
		pkgUser.Mgr = realUserMgr
	}()
	mockManager := &mockUser.Manager{}
	pkgUser.Mgr = mockManager

	mockManager.On("Get", context.Background(), 1).Return(&models.User{Username: "testuser"}, nil)

	tests := []struct {
		userID   string
		expected string
	}{
		{"1", "testuser"},
		{"invalid", ""},
		{"2", ""},
	}
	for _, test := range tests {
		assert.Equal(c.T(), test.expected, UserIDToName(test.userID))
	}
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, &UserTestSuite{})
}
