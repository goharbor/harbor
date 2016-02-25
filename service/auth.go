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
	"log"
	"net/http"

	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils"

	"github.com/astaxie/beego"
	"github.com/docker/distribution/registry/auth/token"
)

type AuthController struct {
	beego.Controller
}

//handle request
func (a *AuthController) Auth() {

	request := a.Ctx.Request

	log.Println("request url: " + request.URL.String())
	authorization := request.Header["Authorization"]
	log.Println("authorization:", authorization)
	username, password := utils.ParseBasicAuth(authorization)
	authenticated := authenticate(username, password)

	service := a.GetString("service")
	scope := a.GetString("scope")

	if len(scope) == 0 && !authenticated {
		log.Printf("login request with invalid credentials")
		a.CustomAbort(http.StatusUnauthorized, "")
	}
	access := svc_utils.GetResourceActions(scope)
	for _, a := range access {
		svc_utils.FilterAccess(username, authenticated, a)
	}
	a.serveToken(username, service, access)
}

func (a *AuthController) serveToken(username, service string, access []*token.ResourceActions) {
	writer := a.Ctx.ResponseWriter
	//create token
	rawToken, err := svc_utils.MakeToken(username, service, access)
	if err != nil {
		log.Printf("Failed to make token, error: %v", err)
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
		log.Printf("Error occurred in UserLogin: %v", err)
		return false
	}
	if user == nil {
		return false
	}

	return true
}
