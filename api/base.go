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

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
)

type BaseAPI struct {
	beego.Controller
}

func (b *BaseAPI) Render() error {
	return nil
}

func (b *BaseAPI) RenderError(code int, text string) {
	http.Error(b.Ctx.ResponseWriter, text, code)
}

func (b *BaseAPI) DecodeJSONReq(v interface{}) {
	err := json.Unmarshal(b.Ctx.Input.CopyBody(1<<32), v)
	if err != nil {
		beego.Error("Error while decoding the json request:", err)
		b.CustomAbort(http.StatusBadRequest, "Invalid json request")
	}
}

func (b *BaseAPI) ValidateUser() int {

	sessionUserID := b.GetSession("userId")
	if sessionUserID == nil {
		beego.Warning("No user id in session, canceling request")
		b.CustomAbort(http.StatusUnauthorized, "")
	}
	userID := sessionUserID.(int)
	u, err := dao.GetUser(models.User{UserID: userID})
	if err != nil {
		beego.Error("Error occurred in GetUser:", err)
		b.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if u == nil {
		beego.Warning("User was deleted already, user id: ", userID, " canceling request.")
		b.CustomAbort(http.StatusUnauthorized, "")
	}
	return userID
}
