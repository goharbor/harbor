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

package handler

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"golang.org/x/oauth2"

	"github.com/goharbor/harbor/src/common"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	oidcpkg "github.com/goharbor/harbor/src/pkg/oidc"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/oidc"
)

type oidcAPI struct {
	BaseAPI
	userCtl user.Controller
	// For testing
	getAuth      func(ctx context.Context) (string, error)
	oidcSetting  func(ctx context.Context) (*cfgModels.OIDCSetting, error)
	verifyToken  func(ctx context.Context, rawIDToken string) (*oidcpkg.Token, *oidcpkg.UserInfo, error)
	injectGroups func(info *oidcpkg.UserInfo, u *commonmodels.User)
}

func newOIDCAPI() *oidcAPI {
	return &oidcAPI{
		userCtl:     user.Ctl,
		getAuth:     config.AuthMode,
		oidcSetting: config.OIDCSetting,
		verifyToken: verifyIDToken,
		injectGroups: func(info *oidcpkg.UserInfo, u *commonmodels.User) {
			oidcpkg.InjectGroupsToUser(info, u)
		},
	}
}

func (o oidcAPI) PingOIDC(ctx context.Context, params oidc.PingOIDCParams) middleware.Responder {
	if err := o.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceConfiguration); err != nil {
		return o.SendError(ctx, err)
	}
	err := oidcpkg.TestEndpoint(oidcpkg.Conn{
		URL:        params.Endpoint.URL,
		VerifyCert: params.Endpoint.VerifyCert,
	})
	if err != nil {
		log.Errorf("Failed to verify connection: %+v, err: %v", params.Endpoint, err)
		return o.SendError(ctx, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("failed to verify connection"))
	}
	return oidc.NewPingOIDCOK()
}

func (o oidcAPI) OnboardOIDCUser(ctx context.Context, params oidc.OnboardOIDCUserParams) middleware.Responder {
	auth, err := o.getAuth(ctx)
	if err != nil {
		return o.SendError(ctx, err)
	}
	if auth != common.OIDCAuth {
		return o.SendError(ctx, errors.PreconditionFailedError(nil).WithMessagef("unable to onboard OIDC user under authmode: %s", auth))
	}
	rawIDToken := lib.BearerToken(params.HTTPRequest)
	if rawIDToken == "" {
		return o.SendError(ctx, errors.UnauthorizedError(nil).WithMessage("an ID token is required in the Authorization header using the Bearer scheme"))
	}
	token, info, err := o.verifyToken(ctx, rawIDToken)
	if err != nil {
		return o.SendError(ctx, err)
	}
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return o.SendError(ctx, err)
	}

	u, err := o.userCtl.GetBySubIss(ctx, info.Subject, info.Issuer)
	switch {
	case errors.IsNotFoundErr(err):
		u, err = o.onboardUser(ctx, params.Onboard, info, tokenBytes)
		switch {
		case errors.ErrCode(err) == errors.ConflictCode:
			// A concurrent request may have onboarded the same user, e.g. parallel CI jobs on their first run
			existing, getErr := o.userCtl.GetBySubIss(ctx, info.Subject, info.Issuer)
			if getErr != nil {
				return o.SendError(ctx, err)
			}
			if err := o.updateStoredToken(ctx, existing.UserID, tokenBytes); err != nil {
				return o.SendError(ctx, err)
			}
			u = existing
		case err != nil:
			return o.SendError(ctx, err)
		}
	case err != nil:
		return o.SendError(ctx, err)
	default:
		if err := o.updateStoredToken(ctx, u.UserID, tokenBytes); err != nil {
			return o.SendError(ctx, err)
		}
	}

	u, err = o.userCtl.Get(ctx, u.UserID, nil)
	if err != nil {
		return o.SendError(ctx, err)
	}
	o.injectGroups(info, u)
	// never expose the CLI secret or the token
	u.OIDCUserMeta = nil
	m := &model.User{User: u}
	return oidc.NewOnboardOIDCUserOK().WithPayload(m.ToUserResp())
}

func (o oidcAPI) onboardUser(ctx context.Context, req *models.OIDCOnboardReq, info *oidcpkg.UserInfo, tokenBytes []byte) (*commonmodels.User, error) {
	username, err := o.usernameForOnboarding(ctx, req, info)
	if err != nil {
		return nil, err
	}
	u, err := oidcpkg.NewOnboardUser(info, username, tokenBytes)
	if err != nil {
		return nil, err
	}
	if err := o.userCtl.OnboardOIDCUser(ctx, u); err != nil {
		return nil, err
	}
	log.G(ctx).Infof("OIDC user %s onboarded via ID token", u.Username)
	return u, nil
}

var usernameReplacer = strings.NewReplacer(
	" ", "_", // the same as the OIDC callback
	// common in the subject of CI tokens, e.g. GitHub's "repo:my-org/my-repo:ref:refs/heads/main",
	// but break HTTP Basic auth with the CLI secret.
	":", ".",
)

func (o oidcAPI) usernameForOnboarding(ctx context.Context, req *models.OIDCOnboardReq, info *oidcpkg.UserInfo) (string, error) {
	if req != nil && req.Username != "" {
		if err := oidcpkg.ValidateUsername(req.Username); err != nil {
			return "", err
		}
		if strings.Contains(req.Username, ":") {
			return "", errors.BadRequestError(nil).WithMessage("username must not contain ':', as it can not be used to authenticate with the CLI secret")
		}
		return req.Username, nil
	}
	setting, err := o.oidcSetting(ctx)
	if err != nil {
		return "", err
	}
	if !setting.AutoOnboard {
		return "", errors.BadRequestError(nil).WithMessage("username is required when automatic onboarding is disabled")
	}
	username := usernameReplacer.Replace(info.Username)
	if username == "" {
		return "", errors.BadRequestError(nil).WithMessagef("unable to recover username for auto onboard, username claim: %s", setting.UserClaim)
	}
	if err := oidcpkg.ValidateUsername(username); err != nil {
		return "", err
	}
	return username, nil
}

// updateStoredToken replaces the token persisted for the user, unless the stored token has a refresh token.
//
// A refresh token is only obtained via the normal web-based OIDC login flow (/c/oidc/callback).
// As Harbor can refresh such token on its own, it must not be replaced by an ID token which can not be refreshed.
func (o oidcAPI) updateStoredToken(ctx context.Context, userID int, tokenBytes []byte) error {
	u, err := o.userCtl.Get(ctx, userID, &user.Option{WithOIDCInfo: true})
	if err != nil {
		return err
	}
	meta := u.OIDCUserMeta
	if meta == nil {
		return errors.NotFoundError(nil).WithMessagef("OIDC info of user %d not found", userID)
	}
	stored, err := oidcpkg.DecryptToken(meta.Token)
	if err != nil {
		log.G(ctx).Warningf("failed to decrypt the stored token of user %d, replacing it: %v", userID, err)
	} else if stored.RefreshToken != "" {
		log.G(ctx).Debugf("the stored token of user %d has a refresh token, skip replacing it", userID)
		return nil
	}
	_, t, err := oidcpkg.EncryptSecretAndToken(tokenBytes)
	if err != nil {
		return err
	}
	meta.Token = t
	return o.userCtl.UpdateOIDCMeta(ctx, meta, "token")
}

func verifyIDToken(ctx context.Context, rawIDToken string) (*oidcpkg.Token, *oidcpkg.UserInfo, error) {
	idToken, err := oidcpkg.VerifyToken(ctx, rawIDToken)
	if err != nil {
		log.G(ctx).Warningf("failed to verify ID token: %v", err)
		return nil, nil, errors.UnauthorizedError(nil).WithMessage("invalid ID token")
	}
	token := &oidcpkg.Token{
		Token: oauth2.Token{
			AccessToken: rawIDToken,
			TokenType:   "Bearer",
			// Must set expiry, otherwise the token is considered valid forever (including the CLI secret as well)
			Expiry: idToken.Expiry,
		},
		RawIDToken: rawIDToken,
	}
	info, err := oidcpkg.UserInfoFromToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	return token, info, nil
}
