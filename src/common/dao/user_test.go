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

package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
)

func TestOnBoardUser(t *testing.T) {
	assert := assert.New(t)
	u := &models.User{
		Username: "user1",
		Password: "password1",
		Email:    "dummy@placehodler.com",
		Realname: "daniel",
	}
	err := OnBoardUser(u)
	assert.Nil(err)
	id := u.UserID
	assert.True(id > 0)
	err = OnBoardUser(u)
	assert.Nil(err)
	assert.True(u.UserID == id)
	CleanUser(int64(id))
}
func TestOnBoardUser_EmptyEmail(t *testing.T) {
	assert := assert.New(t)
	u := &models.User{
		Username: "empty_email",
		Password: "password1",
		Realname: "empty_email",
	}
	err := OnBoardUser(u)
	assert.Nil(err)
	id := u.UserID
	assert.True(id > 0)
	err = OnBoardUser(u)
	assert.Nil(err)
	assert.True(u.UserID == id)
	assert.Equal("", u.Email)

	user, err := GetUser(models.User{Username: "empty_email"})
	assert.Equal("", user.Email)
	CleanUser(int64(id))
}
