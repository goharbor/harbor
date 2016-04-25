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
	"net/http"

	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// IndexController handles request to /
type IndexController struct {
	BaseController
}

// Get renders the index page.
func (c *IndexController) Get() {
	c.Data["Username"] = c.GetSession("username")
	c.ForwardTo("page_title_index", "index")
}

// SignInController handles request to /signIn
type SignInController struct {
	BaseController
}

// Get renders Sign In page.
func (sic *SignInController) Get() {
	sic.ForwardTo("page_title_sign_in", "sign-in")
}

// Login handles login request from UI.
func (c *CommonController) Login() {
	principal := c.GetString("principal")
	password := c.GetString("password")

	user, err := auth.Login(models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		c.CustomAbort(http.StatusUnauthorized, "")
	}

	if user == nil {
		c.CustomAbort(http.StatusUnauthorized, "")
	}

	c.SetSession("userId", user.UserID)
	c.SetSession("username", user.Username)
}

// SwitchLanguage handles UI request to switch between different languages and re-render template based on language.
func (c *CommonController) SwitchLanguage() {
	lang := c.GetString("lang")
	if lang == "en-US" || lang == "zh-CN" || lang == "de-DE" {
		c.SetSession("lang", lang)
		c.Data["Lang"] = lang
	}
	c.Redirect(c.Ctx.Request.Header.Get("Referer"), http.StatusFound)
}

// Logout handles UI request to logout.
func (c *CommonController) Logout() {
	c.DestroySession()
}
