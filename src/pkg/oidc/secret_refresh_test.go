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

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/testing/mock"
	daoMock "github.com/goharbor/harbor/src/testing/pkg/oidc/dao"
)

const testEncryptKey = "naa4JtarA1Zsc3uY"

// setupVerifySecret stubs the package level dependencies of VerifySecret and returns a manager whose DAO
// holds an OIDC user with the given secret and token.
func setupVerifySecret(t *testing.T, username, secret string, token *Token) (*defaultManager, *daoMock.MetaDAO) {
	t.Helper()
	oldKeyLoader, oldRefresh, oldUserInfo, oldStored := keyLoader, refreshTokenFn, userInfoFromTokenFn, userInfoFromStoredTokenFn
	keyLoader = &keyGetter{key: testEncryptKey}
	t.Cleanup(func() {
		keyLoader, refreshTokenFn, userInfoFromTokenFn, userInfoFromStoredTokenFn = oldKeyLoader, oldRefresh, oldUserInfo, oldStored
		refreshFailures.Delete(username)
	})

	encSecret, err := utils.ReversibleEncrypt(secret, testEncryptKey)
	require.NoError(t, err)
	tb, err := json.Marshal(token)
	require.NoError(t, err)
	encToken, err := utils.ReversibleEncrypt(string(tb), testEncryptKey)
	require.NoError(t, err)

	dao := &daoMock.MetaDAO{}
	dao.On("GetByUsername", mock.Anything, username).Return(&models.OIDCUser{UserID: 3, Secret: encSecret, Token: encToken}, nil)
	return &defaultManager{metaDao: dao}, dao
}

func expiredToken() *Token {
	return &Token{
		Token:      oauth2.Token{AccessToken: "old-access", RefreshToken: "old-refresh", Expiry: time.Now().Add(-time.Hour)},
		RawIDToken: "old-id-token",
	}
}

func TestVerifySecretKeepsTokenWhenRefreshFails(t *testing.T) {
	dm, dao := setupVerifySecret(t, "alice", "Secret123", expiredToken())
	refreshTokenFn = func(context.Context, *Token) (*Token, error) {
		return nil, fmt.Errorf("invalid_grant: session not active")
	}
	userInfoFromTokenFn = func(context.Context, *Token) (*UserInfo, error) {
		t.Fatal("the remote userinfo must not be called with an expired token")
		return nil, nil
	}
	var storedIDTokens []string
	userInfoFromStoredTokenFn = func(_ context.Context, tk *Token) (*UserInfo, error) {
		storedIDTokens = append(storedIDTokens, tk.RawIDToken)
		return &UserInfo{Username: "alice", Groups: []string{"devs"}}, nil
	}

	for range 3 {
		info, err := dm.VerifySecret(context.Background(), "alice", "Secret123")
		require.NoError(t, err)
		assert.Equal(t, "alice", info.Username)
		assert.Equal(t, []string{"devs"}, info.Groups)
	}
	assert.Equal(t, []string{"old-id-token", "old-id-token", "old-id-token"}, storedIDTokens)
	_, logged := refreshFailures.Load("alice")
	assert.True(t, logged, "the refresh failure must be recorded so it is logged once per user")
	// the stored token is kept, nothing is written back
	dao.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

// persistedTokens decrypts the tokens written by every Update call on the DAO.
func persistedTokens(t *testing.T, dao *daoMock.MetaDAO) []*Token {
	t.Helper()
	var res []*Token
	for _, call := range dao.Calls {
		if call.Method != "Update" {
			continue
		}
		tokenStr, err := utils.ReversibleDecrypt(call.Arguments.Get(1).(*models.OIDCUser).Token, testEncryptKey)
		require.NoError(t, err)
		tk := &Token{}
		require.NoError(t, json.Unmarshal([]byte(tokenStr), tk))
		res = append(res, tk)
	}
	return res
}

func TestVerifySecretUsesStoredUserInfoWhenRefreshFails(t *testing.T) {
	tk := expiredToken()
	// groups only known from the userinfo endpoint, not from the ID token
	tk.UserInfo = &UserInfo{Username: "erin", Groups: []string{"harbor-admins"}, AdminGroupMember: true}
	dm, dao := setupVerifySecret(t, "erin", "Secret123", tk)
	refreshTokenFn = func(context.Context, *Token) (*Token, error) {
		return nil, fmt.Errorf("invalid_grant: session not active")
	}
	userInfoFromStoredTokenFn = func(context.Context, *Token) (*UserInfo, error) {
		t.Fatal("the ID token must not be used when user info is stored with the token")
		return nil, nil
	}

	info, err := dm.VerifySecret(context.Background(), "erin", "Secret123")
	require.NoError(t, err)
	assert.Equal(t, []string{"harbor-admins"}, info.Groups)
	assert.True(t, info.AdminGroupMember)
	dao.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestVerifySecretPersistsRefreshedToken(t *testing.T) {
	dm, dao := setupVerifySecret(t, "bob", "Secret123", expiredToken())
	refreshFailures.Store("bob", struct{}{})
	refreshed := &Token{
		Token:      oauth2.Token{AccessToken: "new-access", RefreshToken: "new-refresh", Expiry: time.Now().Add(time.Hour)},
		RawIDToken: "new-id-token",
	}
	refreshTokenFn = func(context.Context, *Token) (*Token, error) { return refreshed, nil }
	userInfoFromTokenFn = func(_ context.Context, tk *Token) (*UserInfo, error) {
		assert.Equal(t, "new-id-token", tk.RawIDToken)
		return &UserInfo{Username: "bob", Groups: []string{"devs"}}, nil
	}
	dao.On("Update", mock.Anything, mock.Anything, "token").Return(nil)

	info, err := dm.VerifySecret(context.Background(), "bob", "Secret123")
	require.NoError(t, err)
	assert.Equal(t, "bob", info.Username)
	_, logged := refreshFailures.Load("bob")
	assert.False(t, logged, "a successful refresh must reset the once per user warning")
	tokens := persistedTokens(t, dao)
	require.NotEmpty(t, tokens)
	last := tokens[len(tokens)-1]
	assert.Equal(t, "new-refresh", last.RefreshToken)
	assert.Equal(t, "new-id-token", last.RawIDToken)
	require.NotNil(t, last.UserInfo, "the resolved user info must be stored with the token")
	assert.Equal(t, []string{"devs"}, last.UserInfo.Groups)
}

func TestVerifySecretKeepsIDTokenWhenRefreshOmitsIt(t *testing.T) {
	dm, dao := setupVerifySecret(t, "frank", "Secret123", expiredToken())
	refreshTokenFn = func(context.Context, *Token) (*Token, error) {
		// a valid refresh response without an id_token
		return &Token{Token: oauth2.Token{AccessToken: "new-access", RefreshToken: "new-refresh", Expiry: time.Now().Add(time.Hour)}}, nil
	}
	userInfoFromTokenFn = func(_ context.Context, tk *Token) (*UserInfo, error) {
		assert.Equal(t, "old-id-token", tk.RawIDToken)
		return &UserInfo{Username: "frank"}, nil
	}
	dao.On("Update", mock.Anything, mock.Anything, "token").Return(nil)

	_, err := dm.VerifySecret(context.Background(), "frank", "Secret123")
	require.NoError(t, err)
	for _, tk := range persistedTokens(t, dao) {
		assert.Equal(t, "old-id-token", tk.RawIDToken)
		assert.Equal(t, "new-access", tk.AccessToken)
	}
}

func TestVerifySecretStoresUserInfoOnlyWhenChanged(t *testing.T) {
	tk := expiredToken()
	tk.Expiry = time.Now().Add(time.Hour)
	tk.UserInfo = &UserInfo{Username: "gina", Groups: []string{"devs"}}
	dm, dao := setupVerifySecret(t, "gina", "Secret123", tk)
	refreshTokenFn = func(context.Context, *Token) (*Token, error) {
		t.Fatal("a valid token must not be refreshed")
		return nil, nil
	}
	groups := []string{"devs"}
	userInfoFromTokenFn = func(context.Context, *Token) (*UserInfo, error) {
		return &UserInfo{Username: "gina", Groups: groups}, nil
	}
	dao.On("Update", mock.Anything, mock.Anything, "token").Return(nil)

	_, err := dm.VerifySecret(context.Background(), "gina", "Secret123")
	require.NoError(t, err)
	dao.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)

	groups = []string{"devs", "ops"}
	_, err = dm.VerifySecret(context.Background(), "gina", "Secret123")
	require.NoError(t, err)
	tokens := persistedTokens(t, dao)
	require.Len(t, tokens, 1)
	assert.Equal(t, []string{"devs", "ops"}, tokens[0].UserInfo.Groups)
}

func TestVerifySecretRejectsWrongSecretWithoutRefresh(t *testing.T) {
	dm, _ := setupVerifySecret(t, "carol", "Secret123", expiredToken())
	refreshTokenFn = func(context.Context, *Token) (*Token, error) {
		t.Fatal("the token must not be refreshed for a wrong secret")
		return nil, nil
	}

	info, err := dm.VerifySecret(context.Background(), "carol", "WrongSecret1")
	assert.Nil(t, info)
	var sve *SecretVerifyError
	assert.ErrorAs(t, err, &sve)
}

func TestVerifySecretFailsWhenStoredTokenHasNoIDToken(t *testing.T) {
	tk := expiredToken()
	tk.RawIDToken = ""
	dm, _ := setupVerifySecret(t, "dave", "Secret123", tk)
	refreshTokenFn = func(context.Context, *Token) (*Token, error) {
		return nil, fmt.Errorf("invalid_grant")
	}

	info, err := dm.VerifySecret(context.Background(), "dave", "Secret123")
	assert.Nil(t, info)
	var sve *SecretVerifyError
	assert.ErrorAs(t, err, &sve)
}
