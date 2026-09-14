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
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/oidc"
)

const idTokenOnboardComment = "Onboarded via OIDC (ID token)" // harbor_user.comment is varchar(30)

type idToken struct{}

func (i *idToken) Generate(req *http.Request) security.Context {
	ctx := req.Context()
	log := log.G(ctx)
	if lib.GetAuthMode(ctx) != common.OIDCAuth {
		return nil
	}
	if !strings.HasPrefix(req.URL.Path, "/api") && req.URL.Path != "/service/token" {
		return nil
	}
	token := bearerToken(req)
	if len(token) == 0 {
		return nil
	}
	claims, err := oidc.VerifyToken(ctx, token)
	if err != nil {
		log.Warningf("failed to verify token: %v", err)
		return nil
	}
	setting, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("failed to get OIDC settings: %v", err)
		return nil
	}
	info, err := oidc.UserInfoFromIDToken(ctx, &oidc.Token{RawIDToken: token}, *setting)
	if err != nil {
		log.Errorf("Failed to get user info from ID token: %v", err)
		return nil
	}
	u, err := user.Ctl.GetBySubIss(ctx, claims.Subject, claims.Issuer)
	switch {
	case err == nil:
		renewStoredToken(ctx, u.UserID, token, claims.Expiry)
	case hasCode(err, errors.NotFoundCode):
		// Same rule as the browser callback (src/core/controllers/oidc.go):
		// a first-time user is created automatically only when the
		// administrator enabled "Automatic onboarding".
		if !setting.AutoOnboard {
			log.Warningf("user with subject %q is not onboarded and OIDC auto onboard is disabled", claims.Subject)
			return nil
		}
		u, err = onboardFromIDToken(ctx, token, claims.Subject, claims.Issuer, claims.Expiry, info)
		if err != nil {
			log.Warningf("failed to auto onboard user from ID token: %v", err)
			return nil
		}
		log.Infof("user %q auto onboarded from ID token", u.Username)
	default:
		log.Warningf("failed to get user based on token claims: %v", err)
		return nil
	}
	oidc.InjectGroupsToUser(info, u)
	log.Debugf("an ID token security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(u)
}

// onboardFromIDToken mirrors the auto-onboard branch of the browser callback
// for callers that present a verified ID token straight to the API (for
// example a portal that shares the OIDC provider with Harbor and forwards the
// user's token). Username, e-mail and groups come from the same claims the
// callback uses; the CLI secret is generated the same way.
func onboardFromIDToken(ctx context.Context, rawIDToken, sub, iss string, expiry time.Time, info *oidc.UserInfo) (*models.User, error) {
	username := strings.ReplaceAll(info.Username, " ", "_")
	if username == "" {
		return nil, fmt.Errorf("unable to recover username for auto onboard from the ID token claims")
	}
	key, err := config.SecretKey()
	if err != nil {
		return nil, err
	}
	encToken, err := encryptIDTokenAsToken(rawIDToken, expiry, key)
	if err != nil {
		return nil, err
	}
	secret, err := utils.ReversibleEncrypt(utils.GenerateRandomString(), key)
	if err != nil {
		return nil, err
	}
	u := &models.User{
		Username: username,
		Realname: username,
		Email:    info.Email,
		Comment:  idTokenOnboardComment,
		OIDCUserMeta: &models.OIDCUser{
			SubIss: sub + iss,
			Secret: secret,
			Token:  encToken,
		},
	}
	oidc.InjectGroupsToUser(info, u)
	if err := user.Ctl.OnboardOIDCUser(ctx, u); err != nil {
		if hasCode(err, errors.ConflictCode) {
			// Either a concurrent first request from the same user won the
			// race, or a stale record with this username/e-mail exists (for
			// example from a previous IdP connector, which changes `sub`).
			if existing, gerr := user.Ctl.GetBySubIss(ctx, sub, iss); gerr == nil {
				return existing, nil
			}
			return nil, fmt.Errorf("username %q or e-mail %q already exists in Harbor under a different subject; an administrator must remove or relink the stale user: %w", username, info.Email, err)
		}
		return nil, err
	}
	return u, nil
}

// renewStoredToken keeps the CLI secret usable for users whose stored OIDC
// token carries no refresh token, i.e. users onboarded from an ID token.
// CLI-secret verification (pkg/oidc/secret.go) re-validates the stored token
// and tries to refresh it once expired, which cannot succeed without a refresh
// token, so we store the freshest ID token seen on the API instead. Records
// that do hold a refresh token (browser onboarding) are left untouched.
func renewStoredToken(ctx context.Context, userID int, rawIDToken string, expiry time.Time) {
	log := log.G(ctx)
	um, err := user.Ctl.Get(ctx, userID, &user.Option{WithOIDCInfo: true})
	if err != nil || um == nil || um.OIDCUserMeta == nil {
		return
	}
	key, err := config.SecretKey()
	if err != nil {
		return
	}
	stored := &oidc.Token{}
	if plain, derr := utils.ReversibleDecrypt(um.OIDCUserMeta.Token, key); derr == nil {
		_ = json.Unmarshal([]byte(plain), stored)
	}
	if stored.RefreshToken != "" {
		return
	}
	if stored.Valid() && !stored.Expiry.Before(expiry) {
		return
	}
	enc, err := encryptIDTokenAsToken(rawIDToken, expiry, key)
	if err != nil {
		return
	}
	um.OIDCUserMeta.Token = enc
	if err := user.Ctl.UpdateOIDCMeta(ctx, um.OIDCUserMeta, "token"); err != nil {
		log.Warningf("failed to renew stored OIDC token for user %d: %v", userID, err)
	}
}

// encryptIDTokenAsToken wraps a raw ID token into the oidc.Token shape the
// rest of Harbor expects in oidc_user.token. AccessToken is set so that
// oauth2.Token.Valid() reflects the ID token's own expiry; userinfo lookups
// with it may fail, which UserInfoFromToken already tolerates by falling back
// to the ID token claims.
func encryptIDTokenAsToken(rawIDToken string, expiry time.Time, key string) (string, error) {
	tb, err := json.Marshal(&oidc.Token{
		Token:      oauth2.Token{AccessToken: rawIDToken, TokenType: "Bearer", Expiry: expiry},
		RawIDToken: rawIDToken,
	})
	if err != nil {
		return "", err
	}
	return utils.ReversibleEncrypt(string(tb), key)
}

// hasCode reports whether any error in the chain is a Harbor error with the
// given code. errors.IsErr only inspects the outermost *errors.Error, which
// is code-less after errors.Wrap.
func hasCode(err error, code string) bool {
	for err != nil {
		if e, ok := err.(*errors.Error); ok && e.Code == code {
			return true
		}
		err = stderrors.Unwrap(err)
	}
	return false
}
