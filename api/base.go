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

func (b *BaseAPI) DecodeJsonReq(v interface{}) {
	err := json.Unmarshal(b.Ctx.Input.CopyBody(1<<32), v)
	if err != nil {
		beego.Error("Error while decoding the json request:", err)
		b.CustomAbort(400, "Invalid json request")
	}
}

func (b *BaseAPI) ValidateUser() int {

	sessionUserId := b.GetSession("userId")
	if sessionUserId == nil {
		beego.Warning("No user id in session, canceling request")
		b.CustomAbort(401, "")
	}
	userId := sessionUserId.(int)
	u, err := dao.GetUser(models.User{UserId: userId})
	if err != nil {
		beego.Error("Error occurred in GetUser:", err)
		b.CustomAbort(500, "Internal error.")
	}
	if u == nil {
		beego.Warning("User was deleted already, user id: ", userId, " canceling request.")
		b.CustomAbort(401, "")
	}
	return userId
}
