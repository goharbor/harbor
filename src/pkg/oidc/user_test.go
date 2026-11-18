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
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
)

func TestValidateUsername(t *testing.T) {
	cases := []struct {
		username string
		valid    bool
	}{
		{"user", true},
		{"repo:my-org/my-repo:ref:refs/heads/main", false},
		{"", false},
		{strings.Repeat("a", maxUsernameLen+1), false},
		{"robot$ci", false},
		{"user,name", false},
		{`user"name`, false},
	}
	for _, c := range cases {
		err := ValidateUsername(c.username)
		if c.valid {
			assert.NoError(t, err, c.username)
		} else {
			assert.True(t, errors.IsErr(err, errors.BadRequestCode), c.username)
		}
	}
}

func TestEncryptToken(t *testing.T) {
	token := &Token{
		Token:      oauth2.Token{AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: time.Now().Add(time.Hour).Round(0)},
		RawIDToken: "id-token",
	}
	tb, err := json.Marshal(token)
	require.NoError(t, err)

	encToken, err := EncryptToken(tb)
	require.NoError(t, err)

	decrypted, err := DecryptToken(encToken)
	require.NoError(t, err)
	assert.Equal(t, token.AccessToken, decrypted.AccessToken)
	assert.Equal(t, token.RefreshToken, decrypted.RefreshToken)
	assert.Equal(t, token.RawIDToken, decrypted.RawIDToken)
	assert.True(t, token.Expiry.Equal(decrypted.Expiry))

	_, err = DecryptToken("not-encrypted")
	assert.Error(t, err)
}

func TestNewOnboardUser(t *testing.T) {
	info := &UserInfo{
		Issuer:   "https://dex.example.com",
		Subject:  "repo:my-org/my-repo:ref:refs/heads/main",
		Username: "my-repo",
		Email:    "ci@example.com",
	}
	u, err := NewOnboardUser(info, "ci-user", []byte(`{"access_token":"token"}`))
	require.NoError(t, err)
	assert.Equal(t, "ci-user", u.Username)
	assert.Equal(t, "ci-user", u.Realname)
	assert.Equal(t, "ci@example.com", u.Email)
	assert.Equal(t, UserComment, u.Comment)
	require.NotNil(t, u.OIDCUserMeta)
	assert.Equal(t, info.Subject+info.Issuer, u.OIDCUserMeta.SubIss)

	key, err := keyLoader.encryptKey()
	require.NoError(t, err)
	secret, err := utils.ReversibleDecrypt(u.OIDCUserMeta.Secret, key)
	require.NoError(t, err)
	assert.NotEmpty(t, secret)

	token, err := DecryptToken(u.OIDCUserMeta.Token)
	require.NoError(t, err)
	assert.Equal(t, "token", token.AccessToken)
}

func TestNewOnboardUserSubIssLength(t *testing.T) {
	issuer := "https://dex.example.com"
	maxSubject := strings.Repeat("a", maxSubIssLen-len(issuer))

	u, err := NewOnboardUser(&UserInfo{Issuer: issuer, Subject: maxSubject}, "ci-user", []byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, maxSubject+issuer, u.OIDCUserMeta.SubIss)

	_, err = NewOnboardUser(&UserInfo{Issuer: issuer, Subject: maxSubject + "a"}, "ci-user", []byte(`{}`))
	assert.True(t, errors.IsErr(err, errors.BadRequestCode), "unexpected error: %v", err)

	// the length is checked in bytes
	_, err = NewOnboardUser(&UserInfo{Issuer: issuer, Subject: maxSubject[1:] + "ä"}, "ci-user", []byte(`{}`))
	assert.True(t, errors.IsErr(err, errors.BadRequestCode), "unexpected error: %v", err)
}
