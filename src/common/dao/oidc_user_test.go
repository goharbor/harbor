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

func TestOIDCUserMetaDaoMethods(t *testing.T) {

	user111 := models.User{
		Username: "user111",
		Email:    "user111@email.com",
	}
	user222 := models.User{
		Username: "user222",
		Email:    "user222@email.com",
	}
	userEmptyOuMeta := models.User{
		Username: "userEmptyOuMeta",
		Email:    "userEmptyOuMeta@email.com",
	}
	ou111 := models.OIDCUser{
		SubIss: "QWE123123RT1",
		Secret: "QWEQWE1",
	}
	ou222 := models.OIDCUser{
		SubIss: "QWE123123RT2",
		Secret: "QWEQWE2",
	}

	// onboard OIDC ...
	user111.OIDCUserMeta = &ou111
	err := OnBoardOIDCUser(&user111)
	require.Nil(t, err)
	defer CleanUser(int64(user111.UserID))
	user222.OIDCUserMeta = &ou222
	err = OnBoardOIDCUser(&user222)
	require.Nil(t, err)
	defer CleanUser(int64(user222.UserID))

	// empty OIDC user meta ...
	err = OnBoardOIDCUser(&userEmptyOuMeta)
	require.NotNil(t, err)
	assert.Equal(t, "unable to onboard as empty oidc user", err.Error())

	// test get by ID
	oidcUser1, err := GetOIDCUserByID(ou111.ID)
	require.Nil(t, err)
	assert.Equal(t, ou111.UserID, oidcUser1.UserID)

	// test get by userID
	oidcUser2, err := GetOIDCUserByUserID(user111.UserID)
	require.Nil(t, err)
	assert.Equal(t, "QWE123123RT1", oidcUser2.SubIss)

	// test get by sub and iss
	userGetBySubIss, err := GetUserBySubIss("QWE123", "123RT1")
	require.Nil(t, err)
	assert.Equal(t, "user111@email.com", userGetBySubIss.Email)

	// test update
	meta3 := &models.OIDCUser{
		ID:     ou111.ID,
		UserID: ou111.UserID,
		SubIss: "newSub",
		Secret: "newSecret",
	}
	require.Nil(t, UpdateOIDCUser(meta3))
	oidcUser1Update, err := GetOIDCUserByID(ou111.ID)
	require.Nil(t, err)
	assert.Equal(t, "QWE123123RT1", oidcUser1Update.SubIss)
	assert.Equal(t, "newSecret", oidcUser1Update.Secret)

	// clear data
	defer func() {
		_, err := GetOrmer().Raw(`delete from oidc_user`).Exec()
		require.Nil(t, err)
	}()
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
	user666 := models.User{
		Username: "user666",
		Email:    "user666@email.com",
	}
	userDup := models.User{
		Username: "user333",
		Email:    "userDup@email.com",
	}

	ou333 := &models.OIDCUser{
		SubIss: "QWE123123RT3",
		Secret: "QWEQWE333",
	}
	ou555 := &models.OIDCUser{
		SubIss: "QWE123123RT5",
		Secret: "QWEQWE555",
	}
	ouDup := &models.OIDCUser{
		SubIss: "QWE123123RT3",
		Secret: "QWEQWE333",
	}
	ouDupSub := &models.OIDCUser{
		SubIss: "QWE123123RT3",
		Secret: "ouDupSub",
	}

	// data prepare ...
	user333.OIDCUserMeta = ou333
	err := OnBoardOIDCUser(&user333)
	require.Nil(t, err)
	defer CleanUser(int64(user333.UserID))

	// duplicate user -- ErrDupRows
	// userDup is duplicate with user333
	userDup.OIDCUserMeta = ou555
	err = OnBoardOIDCUser(&userDup)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), ErrDupUser.Error())
	exist, err := UserExists(userDup, "email")
	require.Nil(t, err)
	require.False(t, exist)

	// duplicate OIDC user -- ErrDupRows
	// ouDup is duplicate with ou333
	user555.OIDCUserMeta = ouDup
	err = OnBoardOIDCUser(&user555)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), ErrDupOIDCUser.Error())
	exist, err = UserExists(user555, "username")
	require.Nil(t, err)
	require.False(t, exist)

	// success
	user555.OIDCUserMeta = ou555
	err = OnBoardOIDCUser(&user555)
	require.Nil(t, err)
	exist, err = UserExists(user555, "username")
	require.Nil(t, err)
	require.True(t, exist)
	defer CleanUser(int64(user555.UserID))

	// duplicate OIDC user's sub -- ErrDupRows
	// ouDup is duplicate with ou333
	user666.OIDCUserMeta = ouDupSub
	err = OnBoardOIDCUser(&user666)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), ErrDupOIDCUser.Error())
	exist, err = UserExists(user666, "username")
	require.Nil(t, err)
	require.False(t, exist)

	// clear data
	defer func() {
		_, err := GetOrmer().Raw(`delete from oidc_user`).Exec()
		require.Nil(t, err)
	}()

}
