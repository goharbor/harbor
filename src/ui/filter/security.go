// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package filter

import (
	"context"
	"fmt"
	"net/http"

	beegoctx "github.com/astaxie/beego/context"
	"github.com/vmware/harbor/src/common/models"
	secstore "github.com/vmware/harbor/src/common/secret"
	"github.com/vmware/harbor/src/common/security"
	"github.com/vmware/harbor/src/common/security/admiral"
	"github.com/vmware/harbor/src/common/security/authcontext"
	"github.com/vmware/harbor/src/common/security/local"
	"github.com/vmware/harbor/src/common/security/secret"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/projectmanager"
	"github.com/vmware/harbor/src/ui/projectmanager/pms"
)

type key string

const (
	securCtxKey key = "harbor_security_context"
	pmKey       key = "harbor_project_manager"
)

var (
	reqCtxModifiers []ReqCtxModifier
)

// Init ReqCtxMofiers list
func Init() {
	// integration with admiral
	if config.WithAdmiral() {
		reqCtxModifiers = []ReqCtxModifier{
			&secretReqCtxModifier{config.SecretStore},
			&tokenReqCtxModifier{},
			&basicAuthReqCtxModifier{},
			&unauthorizedReqCtxModifier{}}
		return
	}

	// standalone
	reqCtxModifiers = []ReqCtxModifier{
		&secretReqCtxModifier{config.SecretStore},
		&basicAuthReqCtxModifier{},
		&sessionReqCtxModifier{},
		&unauthorizedReqCtxModifier{}}
}

// SecurityFilter authenticates the request and passes a security context
// and a project manager with it which can be used to do some authN & authZ
func SecurityFilter(ctx *beegoctx.Context) {
	if ctx == nil {
		return
	}

	req := ctx.Request
	if req == nil {
		return
	}

	// add security context and project manager to request context
	for _, modifier := range reqCtxModifiers {
		if modifier.Modify(ctx) {
			break
		}
	}
}

// ReqCtxModifier modifies the context of request
type ReqCtxModifier interface {
	Modify(*beegoctx.Context) bool
}

type secretReqCtxModifier struct {
	store *secstore.Store
}

func (s *secretReqCtxModifier) Modify(ctx *beegoctx.Context) bool {
	scrt := ctx.GetCookie("secret")
	if len(scrt) == 0 {
		return false
	}

	log.Debug("got secret from request")

	var pm projectmanager.ProjectManager
	if config.WithAdmiral() {
		// TODO project manager with harbor service accout
	} else {
		log.Debug("using local database project manager")
		pm = config.GlobalProjectMgr
	}

	log.Debug("creating a secret security context...")
	securCtx := secret.NewSecurityContext(scrt, s.store)
	setSecurCtxAndPM(ctx.Request, securCtx, pm)

	return true
}

type basicAuthReqCtxModifier struct{}

func (b *basicAuthReqCtxModifier) Modify(ctx *beegoctx.Context) bool {
	username, password, ok := ctx.Request.BasicAuth()
	if !ok {
		return false
	}
	log.Debug("got user information via basic auth")

	// integration with admiral
	if config.WithAdmiral() {
		// Can't get a token from Admiral's login API, we can only
		// create a project manager with the token of the solution user.
		// That way may cause some wrong permission promotion in some API
		// calls, so we just handle the requests which are necessary
		if !filterReq(ctx.Request) {
			log.Debugf("basic auth is not supported for request %s %s, skip",
				ctx.Request.Method, ctx.Request.URL.Path)
			return false
		}

		authCtx, err := authcontext.Login(config.AdmiralClient,
			config.AdmiralEndpoint(), username, password)
		if err != nil {
			log.Errorf("failed to authenticate %s: %v", username, err)
			return false
		}

		log.Debug("using glocal project manager...")
		pm := config.GlobalProjectMgr
		log.Debug("creating admiral security context...")
		securCtx := admiral.NewSecurityContext(authCtx, pm)

		setSecurCtxAndPM(ctx.Request, securCtx, pm)
		return true
	}

	// standalone
	user, err := auth.Login(models.AuthModel{
		Principal: username,
		Password:  password,
	})
	if err != nil {
		log.Errorf("failed to authenticate %s: %v", username, err)
		return false
	}
	if user == nil {
		log.Debug("basic auth user is nil")
		return false
	}
	log.Debug("using local database project manager")
	pm := config.GlobalProjectMgr
	log.Debug("creating local database security context...")
	securCtx := local.NewSecurityContext(user, pm)

	setSecurCtxAndPM(ctx.Request, securCtx, pm)
	return true
}

func filterReq(req *http.Request) bool {
	path := req.URL.Path
	if path == "/api/projects" && req.Method == http.MethodPost ||
		path == "/service/token" && req.Method == http.MethodGet {
		return true
	}
	return false
}

type sessionReqCtxModifier struct{}

func (s *sessionReqCtxModifier) Modify(ctx *beegoctx.Context) bool {
	username := ctx.Input.Session("username")
	if username == nil {
		return false
	}

	log.Debug("got user information from session")
	user := &models.User{
		Username: username.(string),
	}
	isSysAdmin := ctx.Input.Session("isSysAdmin")
	if isSysAdmin != nil && isSysAdmin.(bool) {
		user.HasAdminRole = 1
	}

	log.Debug("using local database project manager")
	pm := config.GlobalProjectMgr
	log.Debug("creating local database security context...")
	securCtx := local.NewSecurityContext(user, pm)

	setSecurCtxAndPM(ctx.Request, securCtx, pm)

	return true
}

type tokenReqCtxModifier struct{}

func (t *tokenReqCtxModifier) Modify(ctx *beegoctx.Context) bool {
	token := ctx.Request.Header.Get(authcontext.AuthTokenHeader)
	if len(token) == 0 {
		return false
	}

	log.Debug("got token from request")

	authContext, err := authcontext.GetAuthCtx(config.AdmiralClient,
		config.AdmiralEndpoint(), token)
	if err != nil {
		log.Errorf("failed to get auth context: %v", err)
		return false
	}

	log.Debug("creating PMS project manager...")
	pm := pms.NewProjectManager(config.AdmiralClient,
		config.AdmiralEndpoint(), &pms.RawTokenReader{
			Token: token,
		})
	log.Debug("creating admiral security context...")
	securCtx := admiral.NewSecurityContext(authContext, pm)
	setSecurCtxAndPM(ctx.Request, securCtx, pm)

	return true
}

// use this one as the last modifier in the modifier list for unauthorized request
type unauthorizedReqCtxModifier struct{}

func (u *unauthorizedReqCtxModifier) Modify(ctx *beegoctx.Context) bool {
	log.Debug("user information is nil")

	var securCtx security.Context
	var pm projectmanager.ProjectManager
	if config.WithAdmiral() {
		// integration with admiral
		log.Debug("creating PMS project manager...")
		pm = pms.NewProjectManager(config.AdmiralClient,
			config.AdmiralEndpoint(), nil)
		log.Debug("creating admiral security context...")
		securCtx = admiral.NewSecurityContext(nil, pm)
	} else {
		// standalone
		log.Debug("using local database project manager")
		pm = config.GlobalProjectMgr
		log.Debug("creating local database security context...")
		securCtx = local.NewSecurityContext(nil, pm)
	}
	setSecurCtxAndPM(ctx.Request, securCtx, pm)
	return true
}

func setSecurCtxAndPM(req *http.Request, ctx security.Context, pm projectmanager.ProjectManager) {
	addToReqContext(req, securCtxKey, ctx)
	addToReqContext(req, pmKey, pm)
}

func addToReqContext(req *http.Request, key, value interface{}) {
	*req = *(req.WithContext(context.WithValue(req.Context(), key, value)))
}

// GetSecurityContext tries to get security context from request and returns it
func GetSecurityContext(req *http.Request) (security.Context, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	ctx := req.Context().Value(securCtxKey)
	if ctx == nil {
		return nil, fmt.Errorf("the security context got from request is nil")
	}

	c, ok := ctx.(security.Context)
	if !ok {
		return nil, fmt.Errorf("the variable got from request is not security context type")
	}

	return c, nil
}

// GetProjectManager tries to get project manager from request and returns it
func GetProjectManager(req *http.Request) (projectmanager.ProjectManager, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	pm := req.Context().Value(pmKey)
	if pm == nil {
		return nil, fmt.Errorf("the project manager got from request is nil")
	}

	p, ok := pm.(projectmanager.ProjectManager)
	if !ok {
		return nil, fmt.Errorf("the variable got from request is not project manager type")
	}

	return p, nil
}
