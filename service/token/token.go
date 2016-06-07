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

package token

import (
	"net/http"
	"time"

	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"

	"github.com/astaxie/beego"
	"github.com/docker/distribution/registry/auth/token"
)

// Handler handles request on /service/token, which is the auth provider for registry.
type Handler struct {
	beego.Controller
}

// Get handles GET request, it checks the http header for user credentials
// and parse service and scope based on docker registry v2 standard,
// checkes the permission agains local DB and generates jwt token.
func (h *Handler) Get() {

	var username, password string
	request := h.Ctx.Request
	service := h.GetString("service")
	scopes := h.GetStrings("scope")
	access := GetResourceActions(scopes)
	log.Infof("request url: %v", request.URL.String())

	if svc_utils.VerifySecret(request) {
		log.Debugf("Will grant all access as this request is from job service with legal secret.")
		username = "job-service-user"
	} else {
		username, password, _ = request.BasicAuth()
		authenticated := authenticate(username, password)

		if len(scopes) == 0 && !authenticated {
			log.Info("login request with invalid credentials")
			h.CustomAbort(http.StatusUnauthorized, "")
		}
		for _, a := range access {
			FilterAccess(username, authenticated, a)
		}
	}
	h.serveToken(username, service, access)
}

func (h *Handler) serveToken(username, service string, access []*token.ResourceActions) {
	writer := h.Ctx.ResponseWriter
	//create token
	rawToken, expiresIn, issuedAt, err := MakeToken(username, service, access)
	if err != nil {
		log.Errorf("Failed to make token, error: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	tk := make(map[string]interface{})
	tk["token"] = rawToken
	tk["expires_in"] = expiresIn
	tk["issued_at"] = issuedAt.Format(time.RFC3339)
	h.Data["json"] = tk
	h.ServeJSON()
}

func authenticate(principal, password string) bool {
	user, err := auth.Login(models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		return false
	}
	if user == nil {
		return false
	}

	return true
}
