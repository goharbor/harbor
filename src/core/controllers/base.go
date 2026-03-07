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
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/i18n"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	pkguser "github.com/goharbor/harbor/src/pkg/user"
)

// CommonController handles request from UI that doesn't expect a page, such as /SwitchLanguage /logout ...
type CommonController struct {
	api.BaseController
	i18n.Locale
}

// Render returns nil.
func (cc *CommonController) Render() error {
	return nil
}

// Prepare overwrites the Prepare func in api.BaseController to ignore unnecessary steps
func (cc *CommonController) Prepare() {}

func redirectForOIDC(ctx context.Context, username string) bool {
	if lib.GetAuthMode(ctx) != common.OIDCAuth {
		return false
	}
	u, err := user.Ctl.GetByName(ctx, username)
	if err != nil {
		log.Warningf("Failed to get user by name: %s, error: %v", username, err)
	}
	if u == nil {
		return true
	}
	us, err := user.Ctl.Get(ctx, u.UserID, &user.Option{WithOIDCInfo: true})
	if err != nil {
		log.Debugf("Failed to get OIDC user info for user, id: %d, error: %v", u.UserID, err)
	}
	if us != nil && us.OIDCUserMeta != nil {
		return true
	}
	return false
}

// Login handles login request from UI.
func (cc *CommonController) Login() {
	principal := cc.GetString("principal")
	password := cc.GetString("password")
	if redirectForOIDC(cc.Ctx.Request.Context(), principal) {
		ep, err := config.ExtEndpoint()
		if err != nil {
			log.Errorf("Failed to get the external endpoint, error: %v", err)
			cc.CustomAbort(http.StatusUnauthorized, "")
		}
		url := strings.TrimSuffix(ep, "/") + common.OIDCLoginPath
		log.Debugf("Redirect user %s to login page of OIDC provider", principal)
		// Return a json to UI with status code 403, as it cannot handle status 302
		cc.Ctx.Output.Status = http.StatusForbidden
		err = cc.Ctx.Output.JSON(struct {
			Location string `json:"redirect_location"`
		}{url}, false, false)
		if err != nil {
			log.Errorf("Failed to write json to response body, error: %v", err)
		}
		return
	}

	user, err := auth.Login(cc.Context(), models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	if user == nil {
		cc.CustomAbort(http.StatusUnauthorized, "")
	}
	cc.PopulateUserSession(*user)
}

// LogOut Harbor UI
func (cc *CommonController) LogOut() {
	// redirect for OIDC logout.
	securityCtx, ok := security.FromContext(cc.Context())
	if !ok {
		log.Error("Failed to get security context")
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	principal := securityCtx.GetUsername()
	if principal != "" {
		if redirectForOIDC(cc.Ctx.Request.Context(), principal) {
			ep, err := config.ExtEndpoint()
			if err != nil {
				log.Errorf("Failed to get the external endpoint, error: %v", err)
				cc.CustomAbort(http.StatusUnauthorized, "")
			}
			url := strings.TrimSuffix(ep, "/") + common.OIDCLoginoutPath
			log.Debugf("Redirect user %s to logout page of OIDC provider", principal)
			// Return a json to UI with status code 403, as it cannot handle status 302
			cc.Ctx.Output.Status = http.StatusForbidden
			err = cc.Ctx.Output.JSON(struct {
				Location string `json:"redirect_location"`
			}{url}, false, false)
			if err != nil {
				log.Errorf("Failed to write json to response body, error: %v", err)
			}
			return
		}
	}

	if err := cc.DestroySession(); err != nil {
		log.Errorf("Error occurred in LogOut: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

// UserExists checks if user exists when user input value in sign in form.
func (cc *CommonController) UserExists() {
	ctx := cc.Context()
	flag, err := config.SelfRegistration(ctx)
	if err != nil {
		log.Errorf("Failed to get the status of self registration flag, error: %v, disabling user existence check", err)
	}
	securityCtx, ok := security.FromContext(ctx)
	isAdmin := ok && securityCtx.IsSysAdmin()
	if !flag && !isAdmin {
		cc.CustomAbort(http.StatusPreconditionFailed, "self registration deactivated, only sysadmin can check user existence")
	}

	target := cc.GetString("target")
	value := cc.GetString("value")

	var query *q.Query
	switch target {
	case "username":
		query = q.New(q.KeyWords{"Username": value})
	case "email":
		query = q.New(q.KeyWords{"Email": value})
	}

	n, err := user.Ctl.Count(ctx, query)
	if err != nil {
		log.Errorf("Error occurred in UserExists: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	cc.Data["json"] = n > 0
	if err := cc.ServeJSON(); err != nil {
		log.Errorf("failed to serve json: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

// SetupStatus returns whether the one-time admin setup is required.
// GET /c/setup/status
func (cc *CommonController) SetupStatus() {
	ctx := cc.Ctx.Request.Context()
	cfgMgr := config.GetCfgManager(ctx)
	initialized := cfgMgr.Get(ctx, common.AdminInitialized).GetBool()

	setupRequired := false
	if !initialized {
		// Double-check: admin DB record has no salt
		admin, err := pkguser.Mgr.Get(ctx, 1)
		if err != nil {
			log.Errorf("SetupStatus: failed to get admin user: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
			return
		}
		if admin.Salt == "" {
			setupRequired = true
		}
	}

	cc.Data["json"] = map[string]bool{"setup_required": setupRequired}
	if err := cc.ServeJSON(); err != nil {
		log.Errorf("failed to serve json: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

// setupRequest is the JSON body for POST /c/setup
type setupRequest struct {
	Password string `json:"password"`
}

// validSetupPassword validates password strength: 8-128 chars, at least 1 uppercase, 1 lowercase, 1 number.
func validSetupPassword(password string) bool {
	if len(password) < 8 || len(password) > 128 {
		return false
	}
	hasLower := regexp.MustCompile(`[a-z]`)
	hasUpper := regexp.MustCompile(`[A-Z]`)
	hasNumber := regexp.MustCompile(`[0-9]`)
	return hasLower.MatchString(password) && hasUpper.MatchString(password) && hasNumber.MatchString(password)
}

// Setup handles the one-time admin password setup.
// POST /c/setup
func (cc *CommonController) Setup() {
	ctx := cc.Ctx.Request.Context()
	cfgMgr := config.GetCfgManager(ctx)

	// Check precondition: admin not yet initialized
	if cfgMgr.Get(ctx, common.AdminInitialized).GetBool() {
		cc.CustomAbort(http.StatusForbidden, "Setup has already been completed.")
		return
	}

	// Check precondition: admin DB record has no password
	admin, err := pkguser.Mgr.Get(ctx, 1)
	if err != nil {
		log.Errorf("Setup: failed to get admin user: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		return
	}
	if admin.Salt != "" {
		cc.CustomAbort(http.StatusForbidden, "Admin password already exists.")
		return
	}

	// Read password from JSON body or form param
	var password string
	if strings.Contains(cc.Ctx.Input.Header("Content-Type"), "application/json") {
		var req setupRequest
		if err := json.Unmarshal(cc.Ctx.Input.RequestBody, &req); err != nil {
			cc.CustomAbort(http.StatusBadRequest, "Invalid request body.")
			return
		}
		password = req.Password
	} else {
		password = cc.GetString("password")
	}

	if password == "" {
		cc.CustomAbort(http.StatusBadRequest, "Password is required.")
		return
	}

	// Validate password strength
	if !validSetupPassword(password) {
		cc.CustomAbort(http.StatusBadRequest, "Password must be 8-128 characters long with at least 1 uppercase letter, 1 lowercase letter, and 1 number.")
		return
	}

	// Re-check preconditions to guard against race conditions
	admin, err = pkguser.Mgr.Get(ctx, 1)
	if err != nil {
		log.Errorf("Setup: failed to re-check admin user: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		return
	}
	if admin.Salt != "" {
		cc.CustomAbort(http.StatusConflict, "Admin password was set by another request.")
		return
	}

	// Update admin password
	if err := pkguser.Mgr.UpdatePassword(ctx, 1, password); err != nil {
		log.Errorf("Setup: failed to update admin password: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Failed to set admin password.")
		return
	}

	// Persist AdminInitialized=true
	cfgMgr.Set(ctx, common.AdminInitialized, true)
	if err := cfgMgr.Save(ctx); err != nil {
		log.Errorf("Setup: failed to persist AdminInitialized: %v", err)
		// Password was set but flag not saved; it will be corrected on next startup
	}

	log.Info("Admin password set via one-time setup. admin_initialized=true.")

	cc.Data["json"] = map[string]bool{"ok": true}
	if err := cc.ServeJSON(); err != nil {
		log.Errorf("failed to serve json: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

func init() {
	// conf/app.conf -> os.Getenv("config_path")
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		log.Infof("Config path: %s", configPath)
		if err := web.LoadAppConfig("ini", configPath); err != nil {
			log.Errorf("failed to load app config: %v", err)
		}
	}
}
