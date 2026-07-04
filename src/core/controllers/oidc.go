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
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	ctluser "github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/core/api"
	cachepkg "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/oidc"

	"go.pinniped.dev/pkg/oidcclient/pkce"
	"golang.org/x/oauth2"
)

const tokenKey = "oidc_token"
const stateKey = "oidc_state"
const pkceCodeKey = "oidc_pkce_code"
const userInfoKey = "oidc_user_info"
const redirectURLKey = "oidc_redirect_url"
const cliStateSessionKey = "oidc_cli_state"
const oidcUserComment = "Onboarded via OIDC provider"
const oidcCLIStatePrefix = "oidc_cli_state:"
const oidcCLIStatusPending = "pending"
const oidcCLIStatusReady = "ready"
const oidcCLIStatusFailed = "failed"

var oidcCLIPendingTTL = 10 * time.Minute
var oidcCLIResultTTL = 5 * time.Minute

const loginUserOperation = "login_user"

// OIDCController handles requests for OIDC login, callback and user onboard
type OIDCController struct {
	api.BaseController
}

type onboardReq struct {
	Username string `json:"username"`
}

type oidcCLILoginResponse struct {
	RedirectURL string `json:"redirect_url"`
	State       string `json:"state"`
}

type oidcCLITokenResponse struct {
	Status       string `json:"status"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Username     string `json:"username,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
	Error        string `json:"error,omitempty"`
}

type oidcCLIRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type oidcCLIRefreshResponse struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
	Error        string `json:"error,omitempty"`
}

type oidcCLIState struct {
	Status       string `json:"status"`
	PKCECode     string `json:"pkce_code,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Username     string `json:"username,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
	Error        string `json:"error,omitempty"`
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
	if oc.isCLILogin() {
		entry := &oidcCLIState{
			Status:   oidcCLIStatusPending,
			PKCECode: string(pkceCode),
		}
		if err := saveOIDCCLIState(oc.Context(), state, entry, oidcCLIPendingTTL); err != nil {
			oc.SendInternalServerError(err)
			return
		}
		oc.writeJSON(http.StatusOK, &oidcCLILoginResponse{
			RedirectURL: url,
			State:       state,
		})
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
	log.Debugf("State dumped to session: %s", state)
	// Force to use the func 'Redirect' of beego.Controller
	oc.Controller.Redirect(url, http.StatusFound)
}

// Callback handles redirection from OIDC provider.  It will exchange the token and
// kick off onboard if needed.
func (oc *OIDCController) Callback() {
	queryState := oc.Ctx.Request.URL.Query().Get("state")
	cliState, cliFlow, err := getOIDCCLIState(oc.Context(), queryState)
	if err != nil && !errors.Is(err, cachepkg.ErrNotFound) {
		oc.SendInternalServerError(err)
		return
	}
	if errors.Is(err, cachepkg.ErrNotFound) {
		cliFlow = false
	}
	if !cliFlow && queryState != oc.GetSession(stateKey) {
		log.Errorf("State mismatch, in session: %s, in url: %s", oc.GetSession(stateKey),
			queryState)
		oc.SendBadRequestError(errors.New("State mismatch"))
		return
	}
	errorCode := oc.Ctx.Request.URL.Query().Get("error")
	if errorCode != "" {
		errorDescription := oc.Ctx.Request.URL.Query().Get("error_description")
		log.Errorf("OIDC callback returned error: %s - %s", errorCode, errorDescription)
		if cliFlow {
			if err := saveOIDCCLIState(oc.Context(), queryState, &oidcCLIState{
				Status: oidcCLIStatusFailed,
				Error:  fmt.Sprintf("%s - %s", errorCode, errorDescription),
			}, oidcCLIResultTTL); err != nil {
				log.Errorf("failed to save OIDC CLI failure state, error: %v", err)
			}
		}
		oc.SendBadRequestError(errors.Errorf("OIDC callback returned error: %s - %s", errorCode, errorDescription))
		return
	}
	var redirectURLStr string
	if !cliFlow {
		redirectURL := oc.GetSession(redirectURLKey)
		if redirectURL != nil {
			redirectURLStr = redirectURL.(string)
			if err := oc.DelSession(redirectURLKey); err != nil {
				log.Errorf("failed to delete session for key:%s, error: %v", redirectURLKey, err)
				oc.SendInternalServerError(err)
				return
			}
		}
	}
	var pkceCode string
	if cliFlow {
		pkceCode = cliState.PKCECode
	} else {
		pkceCode, _ = oc.GetSession(pkceCodeKey).(string)
		if err := oc.DelSession(pkceCodeKey); err != nil {
			log.Warningf("failed to delete session for key:%s, error: %v", pkceCodeKey, err)
		}
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
	verifiedToken, err := oidc.VerifyToken(ctx, token.RawIDToken)
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
			if cliFlow {
				if err := oc.SetSession(cliStateSessionKey, queryState); err != nil {
					log.Errorf("failed to set session for key: %s, error: %v", cliStateSessionKey, err)
					oc.SendInternalServerError(err)
					return
				}
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
	if cliFlow {
		if err := saveOIDCCLIState(ctx, queryState, &oidcCLIState{
			Status:       oidcCLIStatusReady,
			IDToken:      token.RawIDToken,
			RefreshToken: token.RefreshToken,
			Username:     u.Username,
			ExpiresAt:    verifiedToken.Expiry.Unix(),
		}, oidcCLIResultTTL); err != nil {
			oc.SendInternalServerError(err)
			return
		}
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

func (oc *OIDCController) CLIToken() {
	state := oc.GetString("state")
	if state == "" {
		oc.SendBadRequestError(errors.New("missing state"))
		return
	}

	entry, _, err := getOIDCCLIState(oc.Context(), state)
	if errors.Is(err, cachepkg.ErrNotFound) {
		oc.writeJSON(http.StatusAccepted, &oidcCLITokenResponse{Status: oidcCLIStatusPending})
		return
	}
	if err != nil {
		oc.SendInternalServerError(err)
		return
	}

	switch entry.Status {
	case oidcCLIStatusReady:
		if err := deleteOIDCCLIState(oc.Context(), state); err != nil {
			oc.SendInternalServerError(err)
			return
		}
		oc.writeJSON(http.StatusOK, &oidcCLITokenResponse{
			Status:       oidcCLIStatusReady,
			IDToken:      entry.IDToken,
			RefreshToken: entry.RefreshToken,
			Username:     entry.Username,
			ExpiresAt:    entry.ExpiresAt,
		})
	case oidcCLIStatusFailed:
		oc.writeJSON(http.StatusBadRequest, &oidcCLITokenResponse{
			Status: oidcCLIStatusFailed,
			Error:  entry.Error,
		})
	default:
		oc.writeJSON(http.StatusAccepted, &oidcCLITokenResponse{Status: oidcCLIStatusPending})
	}
}

func (oc *OIDCController) Refresh() {
	req := &oidcCLIRefreshRequest{}
	if err := oc.DecodeJSONReq(req); err != nil {
		oc.SendBadRequestError(err)
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		oc.SendBadRequestError(errors.New("missing refresh_token"))
		return
	}

	ctx := oc.Ctx.Request.Context()
	token, err := oidc.RefreshToken(ctx, &oidc.Token{
		Token: oauth2.Token{
			RefreshToken: req.RefreshToken,
		},
	})
	if err != nil {
		log.Errorf("failed to refresh OIDC token: %v", err)
		oc.writeJSON(http.StatusBadRequest, &oidcCLIRefreshResponse{
			Error: "failed to refresh OIDC token",
		})
		return
	}
	if token.RawIDToken == "" {
		oc.writeJSON(http.StatusBadRequest, &oidcCLIRefreshResponse{
			Error: "OIDC provider did not return an id_token",
		})
		return
	}

	verifiedToken, err := oidc.VerifyToken(ctx, token.RawIDToken)
	if err != nil {
		log.Errorf("failed to verify refreshed OIDC token: %v", err)
		oc.writeJSON(http.StatusBadRequest, &oidcCLIRefreshResponse{
			Error: "failed to verify refreshed OIDC token",
		})
		return
	}

	oc.writeJSON(http.StatusOK, &oidcCLIRefreshResponse{
		IDToken:      token.RawIDToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    verifiedToken.Expiry.Unix(),
	})
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
	log.Debugf("Redirect user to logout page of OIDC provider: %s", logoutURL)
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
		if cliState, ok := oc.GetSession(cliStateSessionKey).(string); ok && cliState != "" {
			token, err := unmarshalOIDCToken(tb)
			if err != nil {
				oc.SendInternalServerError(err)
				return
			}
			verifiedToken, err := oidc.VerifyToken(ctx, token.RawIDToken)
			if err != nil {
				oc.SendInternalServerError(err)
				return
			}
			if err := saveOIDCCLIState(ctx, cliState, &oidcCLIState{
				Status:       oidcCLIStatusReady,
				IDToken:      token.RawIDToken,
				RefreshToken: token.RefreshToken,
				Username:     user.Username,
				ExpiresAt:    verifiedToken.Expiry.Unix(),
			}, oidcCLIResultTTL); err != nil {
				oc.SendInternalServerError(err)
				return
			}
			if err := oc.DelSession(cliStateSessionKey); err != nil {
				log.Warningf("failed to delete session for key:%s, error: %v", cliStateSessionKey, err)
			}
		}
	}
}

func (oc *OIDCController) isCLILogin() bool {
	return oc.Ctx.Request.URL.Query().Get("mode") == "cli"
}

func (oc *OIDCController) writeJSON(status int, payload any) {
	// These CLI endpoints can carry sensitive auth material (e.g. refresh tokens).
	// Prevent caching by browsers/proxies.
	oc.Ctx.Output.Header("Cache-Control", "no-store")
	oc.Ctx.Output.Header("Pragma", "no-cache")
	oc.Ctx.Output.SetStatus(status)
	oc.Data["json"] = payload
	if err := oc.ServeJSON(); err != nil {
		log.Errorf("failed to serve json, %v", err)
		oc.SendInternalServerError(err)
	}
}

func oidcCLIKey(state string) string {
	return oidcCLIStatePrefix + state
}

func oidcCLICache() (cachepkg.Cache, error) {
	c := cachepkg.Default()
	if c == nil {
		return nil, errors.New("cache is not initialized")
	}
	return c, nil
}

func saveOIDCCLIState(ctx context.Context, state string, entry *oidcCLIState, ttl time.Duration) error {
	c, err := oidcCLICache()
	if err != nil {
		return err
	}
	return c.Save(ctx, oidcCLIKey(state), entry, ttl)
}

func getOIDCCLIState(ctx context.Context, state string) (*oidcCLIState, bool, error) {
	c, err := oidcCLICache()
	if err != nil {
		return nil, false, err
	}
	entry := &oidcCLIState{}
	if err := c.Fetch(ctx, oidcCLIKey(state), entry); err != nil {
		if errors.Is(err, cachepkg.ErrNotFound) {
			return nil, false, err
		}
		return nil, false, err
	}
	return entry, true, nil
}

func deleteOIDCCLIState(ctx context.Context, state string) error {
	c, err := oidcCLICache()
	if err != nil {
		return err
	}
	return c.Delete(ctx, oidcCLIKey(state))
}

func unmarshalOIDCToken(tokenBytes []byte) (*oidc.Token, error) {
	token := &oidc.Token{}
	if err := json.Unmarshal(tokenBytes, token); err != nil {
		return nil, err
	}
	return token, nil
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
