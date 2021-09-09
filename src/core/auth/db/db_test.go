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
package db

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/testing/mock"
	testinguserpkg "github.com/goharbor/harbor/src/testing/pkg/user"
	testifymock "github.com/stretchr/testify/mock"
)

func TestSearchUser(t *testing.T) {
	user := &models.User{
		UserID:   123,
		Username: "existuser",
		Email:    "existuser@placeholder.com",
		Realname: "Existing user",
	}

	mockUserMgr := &testinguserpkg.Manager{}
	auth := &Auth{
		userMgr: mockUserMgr,
	}

	mockUserMgr.On("GetByName", mock.Anything, testifymock.MatchedBy(
		func(name string) bool {
			return name == "existuser"
		})).Return(user, nil)

	newUser, err := auth.SearchUser(context.TODO(), "existuser")
	if err != nil {
		t.Fatalf("Failed to search user, error %v", err)
	}
	if newUser == nil {
		t.Fatalf("Failed to search user %v", newUser)
	}
}
