// Copyright 2018 Project Harbor Authors
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
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/oidc"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

const idTokenKey = "oidc_id_token"
const stateKey = "oidc_state"

// OIDCController handles requests for OIDC login, callback and user onboard
type OIDCController struct {
	api.BaseController
}

type oidcUserData struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Username string `json:"name"`
	Email    string `json:"email"`
}

// Prepare include public code path for call request handler of OIDCController
func (oc *OIDCController) Prepare() {
	if mode, _ := config.AuthMode(); mode != common.OIDCAuth {
		oc.CustomAbort(http.StatusPreconditionFailed, fmt.Sprintf("Auth Mode: %s is not OIDC based.", mode))
	}
}

// RedirectLogin redirect user's browser to OIDC provider's login page
func (oc *OIDCController) RedirectLogin() {
	state := utils.GenerateRandomString()
	url, err := oidc.AuthCodeURL(state)
	if err != nil {
		oc.RenderFormatedError(http.StatusInternalServerError, err)
		return
	}
	oc.SetSession(stateKey, state)
	// Force to use the func 'Redirect' of beego.Controller
	oc.Controller.Redirect(url, http.StatusFound)
}

// Callback handles redirection from OIDC provider.  It will exchange the token and
// kick off onboard if needed.
func (oc *OIDCController) Callback() {
	if oc.Ctx.Request.URL.Query().Get("state") != oc.GetSession(stateKey) {
		oc.RenderError(http.StatusBadRequest, "State mismatch.")
		return
	}
	code := oc.Ctx.Request.URL.Query().Get("code")
	ctx := oc.Ctx.Request.Context()
	token, err := oidc.ExchangeToken(ctx, code)
	if err != nil {
		oc.RenderFormatedError(http.StatusInternalServerError, err)
		return
	}
	idToken, err := oidc.VerifyToken(ctx, token.IDToken)
	if err != nil {
		oc.RenderFormatedError(http.StatusInternalServerError, err)
		return
	}
	d := &oidcUserData{}
	err = idToken.Claims(d)
	if err != nil {
		oc.RenderFormatedError(http.StatusInternalServerError, err)
		return
	}
	ouDataStr, err := json.Marshal(d)
	if err != nil {
		oc.RenderFormatedError(http.StatusInternalServerError, err)
		return
	}
	oc.SetSession(idTokenKey, string(ouDataStr))
	// TODO: check and trigger onboard popup or redirect user to project page
	oc.Data["json"] = d
	oc.ServeFormatted()
}

// Onboard handles the request to onboard an user authenticated via OIDC provider
func (oc *OIDCController) Onboard() {

	username := oc.GetString("username")
	if utils.IsIllegalLength(username, 1, 255) {
		oc.RenderFormatedError(http.StatusBadRequest, errors.New("username with illegal length"))
		return
	}
	if utils.IsContainIllegalChar(username, []string{",", "~", "#", "$", "%"}) {
		oc.RenderFormatedError(http.StatusBadRequest, errors.New("username contains illegal characters"))
		return
	}

	idTokenStr := oc.GetSession(idTokenKey)
	d := &oidcUserData{}
	err := json.Unmarshal([]byte(idTokenStr.(string)), &d)
	if err != nil {
		oc.RenderFormatedError(http.StatusInternalServerError, err)
		return
	}
	oidcUser := models.OIDCUser{
		SubIss: d.Subject + d.Issuer,
		// TODO: get secret with secret manager.
		Secret: utils.GenerateRandomString(),
	}

	var email string
	if d.Email == "" {
		email = utils.GenerateRandomString() + "@harbor.com"
	}
	user := models.User{
		Username:     username,
		Email:        email,
		OIDCUserMeta: &oidcUser,
	}

	err = dao.OnBoardOIDCUser(&user)
	if err != nil {
		if strings.Contains(err.Error(), dao.ErrDupUser.Error()) {
			oc.RenderFormatedError(http.StatusConflict, err)
			return
		}
		oc.RenderFormatedError(http.StatusInternalServerError, err)
		return
	}

	oc.Controller.Redirect(config.GetPortalURL(), http.StatusMovedPermanently)
}
