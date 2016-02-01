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
package controllers

import (
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/opt_auth"

	"github.com/astaxie/beego"
)

type IndexController struct {
	BaseController
}

func (c *IndexController) Get() {
	c.Data["Username"] = c.GetSession("username")
	c.ForwardTo("page_title_index", "index")
}

type SignInController struct {
	BaseController
}

func (sic *SignInController) Get() {
	sic.ForwardTo("page_title_sign_in", "sign-in")
}

func (c *CommonController) Login() {
	principal := c.GetString("principal")
	password := c.GetString("password")

	user, err := opt_auth.Login(models.AuthModel{principal, password})
	if err != nil {
		beego.Error("Error occurred in UserLogin:", err)
		c.CustomAbort(500, "Internal error.")
	}

	if user == nil {
		c.CustomAbort(401, "")
	}

	c.SetSession("userId", user.UserId)
	c.SetSession("username", user.Username)
}

func (c *CommonController) SwitchLanguage() {
	lang := c.GetString("lang")
	if lang == "en-US" || lang == "zh-CN" {
		c.SetSession("lang", lang)
		c.Data["Lang"] = lang
	}
	c.Redirect(c.Ctx.Request.Header.Get("Referer"), 302)
}

func (c *CommonController) Logout() {
	c.DestroySession()
}
