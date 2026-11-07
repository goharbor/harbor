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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	cfgmodels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/encrypt"
	liberrors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/oidc"
)

const testIssuer = "https://issuer.test/dex"

// fakeIDToken makes verifyIDToken and userInfoFromIDToken answer as if the
// provider had verified a token for the given subject, and restores the real
// functions when the test ends.
func fakeIDToken(t *testing.T, sub string, exp time.Time, info *oidc.UserInfo, verifyErr, infoErr error) {
	t.Helper()
	origVerify, origInfo := verifyIDToken, userInfoFromIDToken
	verifyIDToken = func(_ context.Context, _ string) (*gooidc.IDToken, error) {
		if verifyErr != nil {
			return nil, verifyErr
		}
		return &gooidc.IDToken{Issuer: testIssuer, Subject: sub, Expiry: exp}, nil
	}
	userInfoFromIDToken = func(_ context.Context, _ *oidc.Token, _ cfgmodels.OIDCSetting) (*oidc.UserInfo, error) {
		if infoErr != nil {
			return nil, infoErr
		}
		return info, nil
	}
	t.Cleanup(func() { verifyIDToken, userInfoFromIDToken = origVerify, origInfo })
}

func setupOIDCConfig(t *testing.T, autoOnboard bool) {
	t.Helper()
	keyFile := filepath.Join(t.TempDir(), "key")
	require.NoError(t, os.WriteFile(keyFile, []byte(testSecretKey), 0o600))
	config.InitWithSettings(map[string]any{
		common.AUTHMode:        common.OIDCAuth,
		common.OIDCAutoOnboard: autoOnboard,
		common.OIDCGroupsClaim: "groups",
		common.OIDCGroupFilter: "",
	}, encrypt.NewFileKeyProvider(keyFile))
}

func apiRequest(t *testing.T, token string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/v2.0/projects", nil)
	require.NoError(t, err)
	req = req.WithContext(lib.WithAuthMode(orm.Context(), common.OIDCAuth))
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func newInfo(sub, username string) *oidc.UserInfo {
	return &oidc.UserInfo{
		Subject:  sub,
		Issuer:   testIssuer,
		Username: username,
		Email:    username + "@example.test",
		Groups:   []string{"idtoken-test-group"},
	}
}

func storedToken(t *testing.T, userID int) *oidc.Token {
	t.Helper()
	um, err := user.Ctl.Get(orm.Context(), userID, &user.Option{WithOIDCInfo: true})
	require.NoError(t, err)
	require.NotNil(t, um.OIDCUserMeta)
	plain, err := utils.ReversibleDecrypt(um.OIDCUserMeta.Token, testSecretKey)
	require.NoError(t, err)
	tok := &oidc.Token{}
	require.NoError(t, json.Unmarshal([]byte(plain), tok))
	return tok
}

func TestIDTokenGenerateAutoOnboards(t *testing.T) {
	setupOIDCConfig(t, true)
	sub := fmt.Sprintf("gen-onboard-%d", time.Now().UnixNano())
	username := "gen onboard " + sub[len(sub)-6:]
	fakeIDToken(t, sub, time.Now().Add(time.Hour), newInfo(sub, username), nil, nil)

	ctx := (&idToken{}).Generate(apiRequest(t, "raw.id.token"))
	require.NotNil(t, ctx, "a first contact with auto onboard on must yield a security context")
	assert.True(t, ctx.IsAuthenticated())

	u, err := user.Ctl.GetBySubIss(orm.Context(), sub, testIssuer)
	require.NoError(t, err)
	defer func() { _ = user.Ctl.Delete(orm.Context(), u.UserID) }()
	assert.Equal(t, ctx.GetUsername(), u.Username)
	assert.Equal(t, idTokenOnboardComment, u.Comment)
	assert.NotContains(t, u.Username, " ")

	// The stored token is the ID token itself, valid until its own expiry.
	tok := storedToken(t, u.UserID)
	assert.Equal(t, "raw.id.token", tok.RawIDToken)
	assert.True(t, tok.Valid())
	assert.Empty(t, tok.RefreshToken)

	// Second request from the same user: no second onboarding, same user.
	ctx2 := (&idToken{}).Generate(apiRequest(t, "raw.id.token"))
	require.NotNil(t, ctx2)
	assert.Equal(t, ctx.GetUsername(), ctx2.GetUsername())
}

func TestIDTokenGenerateRespectsAutoOnboardOff(t *testing.T) {
	setupOIDCConfig(t, false)
	sub := fmt.Sprintf("gen-off-%d", time.Now().UnixNano())
	fakeIDToken(t, sub, time.Now().Add(time.Hour), newInfo(sub, "gen off"), nil, nil)

	assert.Nil(t, (&idToken{}).Generate(apiRequest(t, "raw.id.token")), "stock behaviour: unknown subject is unauthenticated")
	_, err := user.Ctl.GetBySubIss(orm.Context(), sub, testIssuer)
	assert.True(t, hasCode(err, liberrors.NotFoundCode), "nothing must be created")
}

func TestIDTokenGenerateRenewsIDTokenRecord(t *testing.T) {
	setupOIDCConfig(t, true)
	sub := fmt.Sprintf("gen-renew-%d", time.Now().UnixNano())
	info := newInfo(sub, "gen renew "+sub[len(sub)-6:])

	// Onboard with a token that expires soon.
	fakeIDToken(t, sub, time.Now().Add(time.Minute), info, nil, nil)
	require.NotNil(t, (&idToken{}).Generate(apiRequest(t, "old.id.token")))
	u, err := user.Ctl.GetBySubIss(orm.Context(), sub, testIssuer)
	require.NoError(t, err)
	defer func() { _ = user.Ctl.Delete(orm.Context(), u.UserID) }()
	assert.Equal(t, "old.id.token", storedToken(t, u.UserID).RawIDToken)

	// A later request with a fresher token replaces the stored one.
	fakeIDToken(t, sub, time.Now().Add(2*time.Hour), info, nil, nil)
	require.NotNil(t, (&idToken{}).Generate(apiRequest(t, "new.id.token")))
	tok := storedToken(t, u.UserID)
	assert.Equal(t, "new.id.token", tok.RawIDToken)
	assert.True(t, tok.Valid())

	// An older token does not roll the record back.
	fakeIDToken(t, sub, time.Now().Add(time.Minute), info, nil, nil)
	require.NotNil(t, (&idToken{}).Generate(apiRequest(t, "older.id.token")))
	assert.Equal(t, "new.id.token", storedToken(t, u.UserID).RawIDToken)
}

func TestIDTokenGenerateLeavesBrowserRecordAlone(t *testing.T) {
	setupOIDCConfig(t, true)
	sub := fmt.Sprintf("gen-browser-%d", time.Now().UnixNano())
	info := newInfo(sub, "gen browser "+sub[len(sub)-6:])

	// A record as the browser callback writes it: full token set with a refresh token.
	browserTok, err := json.Marshal(&oidc.Token{
		Token:      oauth2.Token{AccessToken: "access", RefreshToken: "refresh", TokenType: "Bearer", Expiry: time.Now().Add(-time.Hour)},
		RawIDToken: "browser.id.token",
	})
	require.NoError(t, err)
	encTok, err := utils.ReversibleEncrypt(string(browserTok), testSecretKey)
	require.NoError(t, err)
	encSecret, err := utils.ReversibleEncrypt("cli-secret", testSecretKey)
	require.NoError(t, err)
	u := &models.User{
		Username:     "gen_browser_" + sub[len(sub)-6:],
		Realname:     "browser user",
		Email:        info.Email,
		OIDCUserMeta: &models.OIDCUser{SubIss: sub + testIssuer, Secret: encSecret, Token: encTok},
	}
	require.NoError(t, user.Ctl.OnboardOIDCUser(orm.Context(), u))
	defer func() { _ = user.Ctl.Delete(orm.Context(), u.UserID) }()

	// Even though the stored token is expired, a record holding a refresh
	// token is left for VerifySecret to refresh; the API must not overwrite it.
	fakeIDToken(t, sub, time.Now().Add(time.Hour), info, nil, nil)
	ctx := (&idToken{}).Generate(apiRequest(t, "fresh.id.token"))
	require.NotNil(t, ctx)
	assert.Equal(t, u.Username, ctx.GetUsername())
	tok := storedToken(t, u.UserID)
	assert.Equal(t, "browser.id.token", tok.RawIDToken)
	assert.Equal(t, "refresh", tok.RefreshToken)
}

func TestIDTokenGenerateRenewsUnreadableRecord(t *testing.T) {
	setupOIDCConfig(t, true)
	sub := fmt.Sprintf("gen-unreadable-%d", time.Now().UnixNano())
	info := newInfo(sub, "gen unreadable "+sub[len(sub)-6:])

	// A record whose token column holds nothing usable (for example provisioned
	// out of band) is treated like an expired one and renewed. The secret must
	// still be decryptable: Harbor decrypts it whenever it loads OIDC metadata.
	encSecret, err := utils.ReversibleEncrypt("cli-secret", testSecretKey)
	require.NoError(t, err)
	u := &models.User{
		Username:     "gen_unreadable_" + sub[len(sub)-6:],
		Realname:     "provisioned user",
		Email:        info.Email,
		OIDCUserMeta: &models.OIDCUser{SubIss: sub + testIssuer, Secret: encSecret, Token: ""},
	}
	require.NoError(t, user.Ctl.OnboardOIDCUser(orm.Context(), u))
	defer func() { _ = user.Ctl.Delete(orm.Context(), u.UserID) }()

	fakeIDToken(t, sub, time.Now().Add(time.Hour), info, nil, nil)
	require.NotNil(t, (&idToken{}).Generate(apiRequest(t, "fresh.id.token")))
	tok := storedToken(t, u.UserID)
	assert.Equal(t, "fresh.id.token", tok.RawIDToken)
	assert.True(t, tok.Valid())
}

func TestIDTokenGenerateRefusesStaleUsername(t *testing.T) {
	setupOIDCConfig(t, true)
	sub := fmt.Sprintf("gen-stale-%d", time.Now().UnixNano())
	info := newInfo(sub, "gen stale "+sub[len(sub)-6:])

	// Same username already exists under another subject (e.g. previous IdP connector).
	encSecret, err := utils.ReversibleEncrypt("cli-secret", testSecretKey)
	require.NoError(t, err)
	stale := &models.User{
		Username:     "gen_stale_" + sub[len(sub)-6:],
		Realname:     "stale user",
		Email:        "stale-" + info.Email,
		OIDCUserMeta: &models.OIDCUser{SubIss: "other-" + sub + testIssuer, Secret: encSecret, Token: ""},
	}
	require.NoError(t, user.Ctl.OnboardOIDCUser(orm.Context(), stale))
	defer func() { _ = user.Ctl.Delete(orm.Context(), stale.UserID) }()

	fakeIDToken(t, sub, time.Now().Add(time.Hour), info, nil, nil)
	assert.Nil(t, (&idToken{}).Generate(apiRequest(t, "raw.id.token")), "conflict must not authenticate as the stale user")
	_, err = user.Ctl.GetBySubIss(orm.Context(), sub, testIssuer)
	assert.True(t, hasCode(err, liberrors.NotFoundCode), "nothing half-created")
}

func TestIDTokenGenerateFailurePaths(t *testing.T) {
	setupOIDCConfig(t, true)
	sub := fmt.Sprintf("gen-fail-%d", time.Now().UnixNano())

	// Token the provider rejects.
	fakeIDToken(t, sub, time.Now().Add(time.Hour), nil, errors.New("bad signature"), nil)
	assert.Nil(t, (&idToken{}).Generate(apiRequest(t, "bad.token")))

	// Verified token but claims cannot be parsed.
	fakeIDToken(t, sub, time.Now().Add(time.Hour), nil, nil, errors.New("no claims"))
	assert.Nil(t, (&idToken{}).Generate(apiRequest(t, "raw.id.token")))

	// Verified token without a usable username claim: auto onboard cannot proceed.
	fakeIDToken(t, sub, time.Now().Add(time.Hour), &oidc.UserInfo{Subject: sub, Issuer: testIssuer}, nil, nil)
	assert.Nil(t, (&idToken{}).Generate(apiRequest(t, "raw.id.token")))
	_, err := user.Ctl.GetBySubIss(orm.Context(), sub, testIssuer)
	assert.True(t, hasCode(err, liberrors.NotFoundCode))

	// Requests outside the API paths are ignored regardless of the token.
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/c/oidc/login", nil)
	require.NoError(t, err)
	req = req.WithContext(lib.WithAuthMode(orm.Context(), common.OIDCAuth))
	req.Header.Set("Authorization", "Bearer raw.id.token")
	assert.Nil(t, (&idToken{}).Generate(req))
}
