/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package service

import (
	"net/http"

	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"

	"github.com/astaxie/beego"
	"github.com/docker/distribution/registry/auth/token"
)

// TokenHandler handles request on /service/token, which is the auth provider for registry.
type TokenHandler struct {
	beego.Controller
}

// Get handles GET request, it checks the http header for user credentials
// and parse service and scope based on docker registry v2 standard,
// checkes the permission agains local DB and generates jwt token.
func (a *TokenHandler) Get() {

	request := a.Ctx.Request
	log.Infof("request url: " + request.URL.String())
	username, password, _ := request.BasicAuth()
	authenticated := authenticate(username, password)
	service := a.GetString("service")
	scope := a.GetString("scope")

	if len(scope) == 0 && !authenticated {
		log.Info("login request with invalid credentials")
		a.CustomAbort(http.StatusUnauthorized, "")
	}
	access := svc_utils.GetResourceActions(scope)
	for _, a := range access {
		svc_utils.FilterAccess(username, authenticated, a)
	}
	a.serveToken(username, service, access)
}

func (a *TokenHandler) serveToken(username, service string, access []*token.ResourceActions) {
	writer := a.Ctx.ResponseWriter
	//create token
	rawToken, err := svc_utils.MakeToken(username, service, access)
	if err != nil {
		log.Errorf("Failed to make token, error: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	tk := make(map[string]string)
	tk["token"] = rawToken
	a.Data["json"] = tk
	a.ServeJSON()
}

func authenticate(principal, password string) bool {
	user, err := auth.Login(models.AuthModel{principal, password})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		return false
	}
	if user == nil {
		return false
	}

	return true
}
