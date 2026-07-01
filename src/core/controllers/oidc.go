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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	ctluser "github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/oidc"

	"go.pinniped.dev/pkg/oidcclient/pkce"
)

const tokenKey = "oidc_token"
const stateKey = "oidc_state"
const pkceCodeKey = "oidc_pkce_code"
const userInfoKey = "oidc_user_info"
const redirectURLKey = "oidc_redirect_url"
const oidcUserComment = "Onboarded via OIDC provider"

const loginUserOperation = "login_user"

// OIDCController handles requests for OIDC login and callback
type OIDCController struct {
	api.BaseController
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
	pkceCode, err := pkce.Generate()
	if err != nil {
		log.Errorf("failed to generate PKCE code, error: %v", err)
		oc.SendInternalServerError(err)
		return
	}
	url, err := oidc.AuthCodeURL(oc.Context(), state, pkceCode)
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
	if err := oc.SetSession(pkceCodeKey, string(pkceCode)); err != nil {
		log.Errorf("failed to set session for key: %s, error: %v", pkceCodeKey, err)
		oc.SendInternalServerError(err)
		return
	}
	if err := oc.SetSession(stateKey, state); err != nil {
		log.Errorf("failed to set session for key: %s, error: %v", stateKey, err)
		oc.SendInternalServerError(err)
		return
	}
	log.Infof("State dumped to session: %s", state)
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
	pkceCode, _ := oc.GetSession(pkceCodeKey).(string)
	if err := oc.DelSession(pkceCodeKey); err != nil {
		log.Warningf("failed to delete session for key:%s, error: %v", pkceCodeKey, err)
	}
	code := oc.Ctx.Request.URL.Query().Get("code")
	ctx := oc.Ctx.Request.Context()
	token, err := oidc.ExchangeToken(ctx, code, pkce.Code(pkceCode))
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
	u, err := resolveOIDCUser(ctx, info, tokenBytes)
	if err != nil {
		if errors.IsErr(err, "FORBIDDEN") {
			oc.SendForbiddenError(err)
		} else {
			oc.SendError(err)
		}
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
	if oidcUser == nil {
		log.Errorf("OIDC user metadata is nil for user ID %d after retrieval; user may not have been properly linked or onboarded", u.UserID)
		oc.SendInternalServerError(errors.New("OIDC user metadata is missing; please contact your administrator"))
		return
	}
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
	// The log middleware can capture the OIDC user login event with the URL, but it cannot get the current username from security context because the security context is not ready yet.
	// need to create login event in the OIDC login call back logic
	// to avoid generate duplicate event in audit log ext, the PreCheck function of the login event intentionally bypass the OIDC user login event in log middleware
	// and OIDC's login callback function will create the login event and send it to notification.
	if config.AuditLogEventEnabled(ctx, loginUserOperation) {
		e := &commonevent.Metadata{
			Ctx:           ctx,
			Username:      u.Username,
			RequestMethod: oc.Ctx.Request.Method,
			RequestURL:    oc.Ctx.Request.URL.String(),
		}
		notification.AddEvent(e.Ctx, e, true)
	}
}

func (oc *OIDCController) RedirectLogout() {
	sessionData := oc.GetSession(tokenKey)
	ctx := oc.Ctx.Request.Context()
	if err := oc.DestroySession(); err != nil {
		log.Errorf("Error occurred in LogOut: %v", err)
		oc.SendInternalServerError(err)
		return
	}
	if sessionData == nil {
		log.Warningf("OIDC session token not found.")
		oc.Controller.Redirect("/account/sign-in", http.StatusFound)
		return
	}
	oidcSettings, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("Failed to get OIDC settings: %v", err)
		oc.SendInternalServerError(err)
		return
	}
	if oidcSettings == nil {
		log.Error("OIDC settings is missing.")
		oc.SendInternalServerError(fmt.Errorf("OIDC settings is missing"))
		return
	}
	if !oidcSettings.Logout {
		oc.Controller.Redirect("/account/sign-in", http.StatusFound)
		return
	}
	tk, ok := sessionData.([]byte)
	if !ok {
		log.Error("Invalid OIDC session data format.")
		oc.SendInternalServerError(fmt.Errorf("invalid OIDC session data format"))
		return
	}
	token := oidc.Token{}
	if err := json.Unmarshal(tk, &token); err != nil {
		log.Errorf("Error occurred in Unmarshal: %v", err)
		oc.SendInternalServerError(err)
		return
	}
	if token.RefreshToken != "" {
		sessionType, err := getSessionType(token.RefreshToken)
		if err == nil {
			// If the session is offline, try best to revoke the refresh token.
			if strings.ToLower(sessionType) == "offline" && oidc.EndpointsClaims.RevokeURL != "" {
				if err := oidc.RevokeOIDCRefreshToken(oidc.EndpointsClaims.RevokeURL, token.RefreshToken, oidcSettings.ClientID, oidcSettings.ClientSecret, oidcSettings.VerifyCert); err != nil {
					log.Warningf("Failed to revoke the offline session: %v", err)
				}
			}
		}
	}
	if token.RawIDToken == "" {
		log.Warning("Empty ID token for offline session.")
		oc.Controller.Redirect("/account/sign-in", http.StatusFound)
		return
	}
	if oidc.EndpointsClaims.EndSessionURL == "" {
		log.Warning("Unable to logout OIDC session since the 'end_session_point' is not set.")
		oc.Controller.Redirect("/account/sign-in", http.StatusFound)
		return
	}
	endSessionURL := oidc.EndpointsClaims.EndSessionURL
	baseURL, err := config.ExtEndpoint()
	if err != nil {
		log.Errorf("Failed to get external endpoint: %v", err)
		oc.SendInternalServerError(err)
		return
	}
	postRedirectURL := fmt.Sprintf("%s/account/sign-in", baseURL)
	logoutURL := fmt.Sprintf(
		"%s?id_token_hint=%s&post_logout_redirect_uri=%s",
		endSessionURL,
		url.QueryEscape(token.RawIDToken),
		url.QueryEscape(postRedirectURL),
	)
	log.Infof("Redirect user to logout page of OIDC provider: %s", logoutURL)
	oc.Controller.Redirect(logoutURL, http.StatusFound)
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

	log.Infof("User created: %v\n", user.Username)

	err = ctluser.Ctl.OnboardOIDCUser(ctx, user)
	if err != nil {
		oc.SendError(err)
		return nil, false
	}
	return user, true
}

// resolveOIDCUser determines which Harbor user should be used for the given OIDC identity.
// Returns the resolved user, or nil if the user cannot be provisioned (and the caller should reject with 403).
// This function encapsulates the user resolution logic for testing purposes.
func resolveOIDCUser(ctx context.Context, info *oidc.UserInfo, tokenBytes []byte) (*models.User, error) {
	// First, check if this OIDC identity is already linked
	u, err := ctluser.Ctl.GetBySubIss(ctx, info.Subject, info.Issuer)
	if err == nil {
		// Already linked; return the existing user
		return u, nil
	}

	if !errors.IsNotFoundErr(err) {
		// Some other error occurred (not a "not found")
		return nil, err
	}

	// User not found by sub/iss; try to find an existing local user by email
	existingUser, err := ctluser.Ctl.GetByEmail(ctx, info.Email)
	if err == nil && existingUser != nil {
		// Found an existing local user with matching email; link them
		s, t, err := secretAndToken(tokenBytes)
		if err != nil {
			return nil, err
		}
		if err := ctluser.Ctl.LinkExistingUserToOIDC(ctx, existingUser.UserID, info.Subject, info.Issuer, s, t); err != nil {
			return nil, err
		}
		// Retrieve the full user record with OIDC metadata
		um, err := ctluser.Ctl.Get(ctx, existingUser.UserID, &ctluser.Option{WithOIDCInfo: true})
		if err != nil {
			return nil, err
		}
		return um, nil
	}

	// No existing local user found; check AutoOnboard setting
	oidcSettings, err := config.OIDCSetting(ctx)
	if err != nil {
		return nil, err
	}

	if !oidcSettings.AutoOnboard {
		// AutoOnboard disabled; cannot provision new identity
		return nil, errors.ForbiddenError(nil).WithMessage("your account has not been provisioned; contact your administrator")
	}

	// AutoOnboard enabled; create new user from claims
	username := info.Username
	// Fix blanks in username
	username = strings.Replace(username, " ", "_", -1)
	if username == "" {
		return nil, fmt.Errorf("unable to recover username for auto onboard, username claim: %s", oidcSettings.UserClaim)
	}

	// Create the OIDC user record
	s, t, err := secretAndToken(tokenBytes)
	if err != nil {
		return nil, err
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

	err = ctluser.Ctl.OnboardOIDCUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
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

// getSessionType determines if the session is offline by decoding the refresh token or not
func getSessionType(refreshToken string) (string, error) {
	parts := strings.Split(refreshToken, ".")
	if len(parts) != 3 {
		return "", errors.Errorf("invalid refresh token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.Errorf("failed to decode refresh token: %v", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", errors.Errorf("failed to unmarshal refresh token: %v", err)
	}
	typ, ok := claims["typ"].(string)
	if !ok {
		return "", errors.New("missing 'typ' claim in refresh token")
	}
	return typ, nil
}
