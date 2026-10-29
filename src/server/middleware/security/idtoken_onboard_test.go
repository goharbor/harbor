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

package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/oidc"
)

const testSecretKey = "1234567890123456"

func TestHasCode(t *testing.T) {
	nf := errors.NotFoundError(nil).WithMessage("no such user")
	assert.True(t, hasCode(nf, errors.NotFoundCode))
	assert.False(t, hasCode(nf, errors.ConflictCode))

	wrapped := errors.Wrap(errors.ConflictError(nil).WithMessage("dup"), "failed to create user record")
	assert.True(t, hasCode(wrapped, errors.ConflictCode), "code must be visible through errors.Wrap")

	assert.False(t, hasCode(nil, errors.ConflictCode))
	assert.False(t, hasCode(fmt.Errorf("plain"), errors.ConflictCode))
}

func TestEncryptIDTokenAsToken(t *testing.T) {
	exp := time.Now().Add(time.Hour)

	enc, err := encryptIDTokenAsToken("raw.id.token", exp, testSecretKey)
	require.NoError(t, err)
	plain, err := utils.ReversibleDecrypt(enc, testSecretKey)
	require.NoError(t, err)
	tok := &oidc.Token{}
	require.NoError(t, json.Unmarshal([]byte(plain), tok))
	assert.Equal(t, "raw.id.token", tok.RawIDToken)
	assert.Equal(t, "raw.id.token", tok.AccessToken)
	assert.Empty(t, tok.RefreshToken)
	assert.True(t, tok.Valid(), "stored token must satisfy oauth2.Token.Valid so VerifySecret does not attempt a refresh")

	// An expired ID token yields an invalid stored token, which is what makes
	// renewStoredToken replace it on the next bearer request.
	enc, err = encryptIDTokenAsToken("raw.id.token", time.Now().Add(-time.Minute), testSecretKey)
	require.NoError(t, err)
	plain, err = utils.ReversibleDecrypt(enc, testSecretKey)
	require.NoError(t, err)
	tok = &oidc.Token{}
	require.NoError(t, json.Unmarshal([]byte(plain), tok))
	assert.False(t, tok.Valid())
}

// TestOnboardFromIDToken exercises the DB path: first contact creates the
// user, a repeated first contact resolves to the same user, and a different
// subject reusing the username is refused without leaving a half-created row.
func TestOnboardFromIDToken(t *testing.T) {
	keyFile := filepath.Join(t.TempDir(), "key")
	require.NoError(t, os.WriteFile(keyFile, []byte(testSecretKey), 0o600))
	config.InitWithSettings(map[string]any{
		common.AUTHMode:        common.OIDCAuth,
		common.OIDCAutoOnboard: true,
		common.OIDCGroupsClaim: "groups",
		common.OIDCGroupFilter: "",
	}, encrypt.NewFileKeyProvider(keyFile))

	ctx := orm.Context()
	iss := "https://issuer.test/dex"
	sub := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	info := &oidc.UserInfo{
		Subject:  sub,
		Issuer:   iss,
		Username: "onboard test",
		Email:    fmt.Sprintf("%s@example.test", sub),
		Groups:   []string{"onboard-test-group"},
	}
	exp := time.Now().Add(time.Hour)

	u, err := onboardFromIDToken(ctx, "raw.id.token", sub, iss, exp, info)
	require.NoError(t, err)
	require.NotNil(t, u)
	defer func() { _ = user.Ctl.Delete(ctx, u.UserID) }()
	assert.Equal(t, "onboard_test", u.Username, "spaces in the username claim become underscores, as in the callback")
	assert.Equal(t, info.Email, u.Email)
	assert.Equal(t, idTokenOnboardComment, u.Comment)

	got, err := user.Ctl.GetBySubIss(ctx, sub, iss)
	require.NoError(t, err)
	assert.Equal(t, u.UserID, got.UserID)

	um, err := user.Ctl.Get(ctx, u.UserID, &user.Option{WithOIDCInfo: true})
	require.NoError(t, err)
	require.NotNil(t, um.OIDCUserMeta)
	assert.Equal(t, sub+iss, um.OIDCUserMeta.SubIss)
	assert.NotEmpty(t, um.OIDCUserMeta.Secret)
	assert.NotEmpty(t, um.OIDCUserMeta.Token)

	// Same subject again (the concurrent-first-request race): resolves to the existing user.
	again, err := onboardFromIDToken(ctx, "raw.id.token", sub, iss, exp, info)
	require.NoError(t, err)
	assert.Equal(t, u.UserID, again.UserID)

	// Different subject, same username: refused, nothing created.
	other := *info
	other.Subject = sub + "-other"
	other.Email = "other-" + info.Email
	_, err = onboardFromIDToken(ctx, "raw.id.token", other.Subject, iss, exp, &other)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "different subject")
	_, err = user.Ctl.GetBySubIss(ctx, other.Subject, iss)
	assert.True(t, hasCode(err, errors.NotFoundCode))

	// Empty username claim: refused before touching the DB.
	empty := *info
	empty.Subject = sub + "-empty"
	empty.Username = ""
	_, err = onboardFromIDToken(ctx, "raw.id.token", empty.Subject, iss, exp, &empty)
	require.Error(t, err)
}
