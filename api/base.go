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

package api

import (
	"encoding/json"
	"net/http"

	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"

	"github.com/astaxie/beego"
)

// BaseAPI wraps common methods for controllers to host API
type BaseAPI struct {
	beego.Controller
}

// Render returns nil as it won't render template
func (b *BaseAPI) Render() error {
	return nil
}

// RenderError provides shortcut to render http error
func (b *BaseAPI) RenderError(code int, text string) {
	http.Error(b.Ctx.ResponseWriter, text, code)
}

// DecodeJSONReq decodes a json request
func (b *BaseAPI) DecodeJSONReq(v interface{}) {
	err := json.Unmarshal(b.Ctx.Input.CopyBody(1<<32), v)
	if err != nil {
		log.Errorf("Error while decoding the json request, error: %v", err)
		b.CustomAbort(http.StatusBadRequest, "Invalid json request")
	}
}

// ValidateUser checks if the request triggered by a valid user
func (b *BaseAPI) ValidateUser() int {

	username, password, ok := b.Ctx.Request.BasicAuth()
	if ok {
		log.Infof("Requst with Basic Authentication header, username: %s", username)
		user, err := auth.Login(models.AuthModel{
			Principal: username,
			Password:  password,
		})
		if err != nil {
			log.Errorf("Error while trying to login, username: %s, error: %v", username, err)
			user = nil
		}
		if user != nil {
			return user.UserID
		}
	}
	sessionUserID := b.GetSession("userId")
	if sessionUserID == nil {
		log.Warning("No user id in session, canceling request")
		b.CustomAbort(http.StatusUnauthorized, "")
	}
	userID := sessionUserID.(int)
	u, err := dao.GetUser(models.User{UserID: userID})
	if err != nil {
		log.Errorf("Error occurred in GetUser, error: %v", err)
		b.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if u == nil {
		log.Warningf("User was deleted already, user id: %d, canceling request.", userID)
		b.CustomAbort(http.StatusUnauthorized, "")
	}
	return userID
}

// Redirect does redirection to resource URI with http header status code.
func (b *BaseAPI) Redirect(statusCode int, resouceID string) {
	requestURI := b.Ctx.Request.RequestURI
	resoucreURI := requestURI + "/" + resouceID

	b.Ctx.Redirect(statusCode, resoucreURI)
}
