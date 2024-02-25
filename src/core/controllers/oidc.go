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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	ctluser "github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/oidc"
)

const tokenKey = "oidc_token"
const stateKey = "oidc_state"
const userInfoKey = "oidc_user_info"
const redirectURLKey = "oidc_redirect_url"
const oidcUserComment = "Onboarded via OIDC provider"

// OIDCController handles requests for OIDC login, callback and user onboard
type OIDCController struct {
	api.BaseController
}

type onboardReq struct {
	Username string `json:"username"`
}

// Prepare include public code path for call request handler of OIDCController
func (oc *OIDCController) Prepare() {
	if mode, _ := config.AuthMode(oc.Context()); mode != common.OIDCAuth {
		oc.SendPreconditionFailedError(fmt.Errorf("auth mode: %s is not OIDC based", mode))
		return
	}
}

// RedirectLogin redirect user's browser to OIDC provider's login page
func (oc *OIDCController) RedirectLogin() {
	state := utils.GenerateRandomString()
	url, err := oidc.AuthCodeURL(oc.Context(), state)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	redirectURL := oc.Ctx.Request.URL.Query().Get("redirect_url")
	if !utils.IsLocalPath(redirectURL) {
		log.Errorf("invalid redirect url: %v", redirectURL)
		oc.SendBadRequestError(fmt.Errorf("cannot redirect to other site"))
		return
	}
	if err := oc.SetSession(redirectURLKey, redirectURL); err != nil {
		log.Errorf("failed to set session for key: %s, error: %v", redirectURLKey, err)
		oc.SendInternalServerError(err)
		return
	}
	if err := oc.SetSession(stateKey, state); err != nil {
		log.Errorf("failed to set session for key: %s, error: %v", stateKey, err)
		oc.SendInternalServerError(err)
		return
	}
	log.Debugf("State dumped to session: %s", state)
	// Force to use the func 'Redirect' of beego.Controller
	oc.Controller.Redirect(url, http.StatusFound)
}

// Callback handles redirection from OIDC provider.  It will exchange the token and
// kick off onboard if needed.
func (oc *OIDCController) Callback() {
	if oc.Ctx.Request.URL.Query().Get("state") != oc.GetSession(stateKey) {
		log.Errorf("State mismatch, in session: %s, in url: %s", oc.GetSession(stateKey),
			oc.Ctx.Request.URL.Query().Get("state"))
		oc.SendBadRequestError(errors.New("State mismatch"))
		return
	}

	errorCode := oc.Ctx.Request.URL.Query().Get("error")
	if errorCode != "" {
		errorDescription := oc.Ctx.Request.URL.Query().Get("error_description")
		log.Errorf("OIDC callback returned error: %s - %s", errorCode, errorDescription)
		oc.SendBadRequestError(errors.Errorf("OIDC callback returned error: %s - %s", errorCode, errorDescription))
		return
	}
	var redirectURLStr string
	redirectURL := oc.GetSession(redirectURLKey)
	if redirectURL != nil {
		redirectURLStr = redirectURL.(string)
		if err := oc.DelSession(redirectURLKey); err != nil {
			log.Errorf("failed to delete session for key:%s, error: %v", redirectURLKey, err)
			oc.SendInternalServerError(err)
			return
		}
	}
	code := oc.Ctx.Request.URL.Query().Get("code")
	ctx := oc.Ctx.Request.Context()
	token, err := oidc.ExchangeToken(ctx, code)
	if err != nil {
		log.Errorf("Failed to exchange token, error: %v", err)
		// Return a 4xx error so user can see the details in case it's due to misconfiguration.
		oc.SendBadRequestError(err)
		return
	}
	_, err = oidc.VerifyToken(ctx, token.RawIDToken)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	info, err := oidc.UserInfoFromToken(ctx, token)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	ouDataStr, err := json.Marshal(info)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	if err := oc.SetSession(tokenKey, tokenBytes); err != nil {
		log.Errorf("failed to set session for key: %s, error: %v", tokenKey, err)
		oc.SendInternalServerError(err)
		return
	}
	u, err := ctluser.Ctl.GetBySubIss(ctx, info.Subject, info.Issuer)
	if errors.IsNotFoundErr(err) { // User is not onboarded, kickoff the onboard flow
		// Recover the username from d.Username by default
		username := info.Username
		// Fix blanks in username
		username = strings.Replace(username, " ", "_", -1)
		oidcSettings, err := config.OIDCSetting(ctx)
		if err != nil {
			oc.SendInternalServerError(err)
			return
		}
		// If automatic onboard is enabled, skip the onboard page
		if oidcSettings.AutoOnboard {
			log.Debug("Doing automatic onboarding\n")
			if username == "" {
				oc.SendInternalServerError(fmt.Errorf("unable to recover username for auto onboard, username claim: %s",
					oidcSettings.UserClaim))
				return
			}
			userRec, onboarded := userOnboard(ctx, oc, info, username, tokenBytes)
			if !onboarded {
				log.Error("User not onboarded\n")
				return
			}
			log.Debug("User automatically onboarded\n")
			u = userRec
		} else {
			if err := oc.SetSession(userInfoKey, string(ouDataStr)); err != nil {
				log.Errorf("failed to set session for key: %s, error: %v", userInfoKey, err)
				oc.SendInternalServerError(err)
				return
			}
			oc.Controller.Redirect(fmt.Sprintf("/oidc-onboard?username=%s&redirect_url=%s", username, redirectURLStr), http.StatusFound)
			// Once redirected, no further actions are done
			return
		}
	} else if err != nil {
		oc.SendError(err)
		return
	}
	oidc.InjectGroupsToUser(info, u)
	um, err := ctluser.Ctl.Get(ctx, u.UserID, &ctluser.Option{WithOIDCInfo: true})
	if err != nil {
		oc.SendError(err)
		return
	}
	_, t, err := secretAndToken(tokenBytes)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	oidcUser := um.OIDCUserMeta
	oidcUser.Token = t
	if err := ctluser.Ctl.UpdateOIDCMeta(ctx, oidcUser); err != nil {
		oc.SendError(err)
		return
	}
	oc.PopulateUserSession(*u)

	if redirectURLStr == "" {
		redirectURLStr = "/"
	}
	oc.Controller.Redirect(redirectURLStr, http.StatusFound)
}

func userOnboard(ctx context.Context, oc *OIDCController, info *oidc.UserInfo, username string, tokenBytes []byte) (*models.User, bool) {
	s, t, err := secretAndToken(tokenBytes)
	if err != nil {
		oc.SendInternalServerError(err)
		return nil, false
	}
	oidcUser := models.OIDCUser{
		SubIss: info.Subject + info.Issuer,
		Secret: s,
		Token:  t,
	}

	user := &models.User{
		Username:     username,
		Realname:     username,
		Email:        info.Email,
		OIDCUserMeta: &oidcUser,
		Comment:      oidcUserComment,
	}
	oidc.InjectGroupsToUser(info, user)

	log.Debugf("User created: %v\n", user.Username)

	err = ctluser.Ctl.OnboardOIDCUser(ctx, user)
	if err != nil {
		oc.SendError(err)
		return nil, false
	}
	return user, true
}

// Onboard handles the request to onboard a user authenticated via OIDC provider
func (oc *OIDCController) Onboard() {
	u := &onboardReq{}
	if err := oc.DecodeJSONReq(u); err != nil {
		oc.SendBadRequestError(err)
		return
	}
	username := u.Username
	if utils.IsIllegalLength(username, 1, 255) {
		oc.SendBadRequestError(errors.New("username with illegal length"))
		return
	}
	if strings.ContainsAny(username, common.IllegalCharsInUsername) {
		oc.SendBadRequestError(errors.Errorf("username %v contains illegal characters: %v", username, common.IllegalCharsInUsername))
		return
	}

	userInfoStr, ok := oc.GetSession(userInfoKey).(string)
	if !ok {
		oc.SendBadRequestError(errors.New("Failed to get OIDC user info from session"))
		return
	}
	log.Debugf("User info string: %s\n", userInfoStr)
	tb, ok := oc.GetSession(tokenKey).([]byte)
	if !ok {
		oc.SendBadRequestError(errors.New("Failed to get OIDC token from session"))
		return
	}

	d := &oidc.UserInfo{}
	err := json.Unmarshal([]byte(userInfoStr), &d)
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}
	ctx := oc.Ctx.Request.Context()
	if user, onboarded := userOnboard(ctx, oc, d, username, tb); onboarded {
		user.OIDCUserMeta = nil
		if err := oc.DelSession(userInfoKey); err != nil {
			log.Errorf("failed to delete session for key:%s, error: %v", userInfoKey, err)
			oc.SendInternalServerError(err)
			return
		}
		oc.PopulateUserSession(*user)
	}
}

func secretAndToken(tokenBytes []byte) (string, string, error) {
	key, err := config.SecretKey()
	if err != nil {
		return "", "", err
	}
	token, err := utils.ReversibleEncrypt((string)(tokenBytes), key)
	if err != nil {
		return "", "", err
	}
	str := utils.GenerateRandomString()
	secret, err := utils.ReversibleEncrypt(str, key)
	if err != nil {
		return "", "", err
	}
	return secret, token, nil
}
