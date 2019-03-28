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
	"github.com/stretchr/testify/require"
)

var (
	user111 = models.User{
		Username: "user111",
		Email:    "user111@email.com",
	}
	user222 = models.User{
		Username: "user222",
		Email:    "user222@email.com",
	}
	ou111 = &models.OIDCUser{
		SubIss: "QWE123123RT1",
		Secret: "QWEQWE1",
	}
	ou222 = &models.OIDCUser{
		SubIss: "QWE123123RT2",
		Secret: "QWEQWE2",
	}
)

func TestOIDCUserMetaDaoMethods(t *testing.T) {

	err := OnBoardUser(&user111)
	require.Nil(t, err)
	ou111.UserID = user111.UserID
	err = OnBoardUser(&user222)
	require.Nil(t, err)
	ou222.UserID = user222.UserID

	// test add
	_, err = AddOIDCUser(ou111)
	require.Nil(t, err)
	_, err = AddOIDCUser(ou222)
	require.Nil(t, err)

	// test get
	oidcUser1, err := GetOIDCUserByID(ou111.ID)
	require.Nil(t, err)
	assert.Equal(t, ou111.UserID, oidcUser1.UserID)

	// test update
	meta3 := &models.OIDCUser{
		ID:     ou111.ID,
		UserID: ou111.UserID,
		SubIss: "newSub",
	}
	require.Nil(t, UpdateOIDCUser(meta3))
	oidcUser1Update, err := GetOIDCUserByID(ou111.ID)
	require.Nil(t, err)
	assert.Equal(t, "newSub", oidcUser1Update.SubIss)

	user, err := GetUserBySub("newSub")
	require.Nil(t, err)
	assert.Equal(t, "user111", user.Username)
}

func TestOIDCOnboard(t *testing.T) {
	user333 := models.User{
		Username: "user333",
		Email:    "user333@email.com",
	}
	user555 := models.User{
		Username: "user555",
		Email:    "user555@email.com",
	}
	ou333 := &models.OIDCUser{
		UserID: 333,
		SubIss: "QWE123123RT1",
		Secret: "QWEQWE333",
	}
	ouDupSub := &models.OIDCUser{
		UserID: 444,
		SubIss: "QWE123123RT1",
		Secret: "QWEQWE444",
	}

	// duplicate user -- ErrDupRows
	user111.OIDCUserMeta = ou333
	err := OnBoardOIDCUser(user111)
	require.NotNil(t, err)
	require.Equal(t, err, ErrDupRows)

	// duplicate OIDC user -- ErrDupRows
	user333.OIDCUserMeta = ou111
	err = OnBoardOIDCUser(user333)
	require.NotNil(t, err)
	require.Equal(t, err, ErrDupRows)

	// success
	user333.OIDCUserMeta = ou333
	err = OnBoardOIDCUser(user333)
	require.Nil(t, err)

	// duplicate OIDC user's sub -- ErrDupRows
	user555.OIDCUserMeta = ouDupSub
	err = OnBoardOIDCUser(user555)
	require.NotNil(t, err)
	require.Equal(t, err, ErrDupRows)

}
