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
	"bytes"
	"os"
	"regexp"
	"text/template"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils"

	"github.com/astaxie/beego"
)

type ChangePasswordController struct {
	BaseController
}

func (cpc *ChangePasswordController) Get() {
	sessionUserId := cpc.GetSession("userId")
	if sessionUserId == nil {
		cpc.Redirect("/signIn", 302)
	}
	cpc.Data["Username"] = cpc.GetSession("username")
	cpc.ForwardTo("page_title_change_password", "change-password")
}

func (cpc *CommonController) UpdatePassword() {

	sessionUserId := cpc.GetSession("userId")
	sessionUsername := cpc.GetSession("username")

	if sessionUserId == nil || sessionUsername == nil {
		beego.Warning("User does not login.")
		cpc.CustomAbort(401, "please_login_first")
	}

	oldPassword := cpc.GetString("old_password")
	queryUser := models.User{UserId: sessionUserId.(int), Username: sessionUsername.(string), Password: oldPassword}
	user, err := dao.CheckUserPassword(queryUser)
	if err != nil {
		beego.Error("Error occurred in CheckUserPassword:", err)
		cpc.CustomAbort(500, "Internal error.")
	}

	if user == nil {
		beego.Warning("Password input is not correct")
		cpc.CustomAbort(403, "old_password_is_not_correct")
	}

	password := cpc.GetString("password")
	if password != "" {
		updateUser := models.User{UserId: sessionUserId.(int), Username: sessionUsername.(string), Password: password, Salt: user.Salt}
		dao.ChangeUserPassword(updateUser)
	} else {
		cpc.CustomAbort(404, "please_input_new_password")
	}
}

type ForgotPasswordController struct {
	BaseController
}

type MessageDetail struct {
	Hint string
	Url  string
	Uuid string
}

func (fpc *ForgotPasswordController) Get() {
	fpc.ForwardTo("page_title_forgot_password", "forgot-password")
}

func (fpc *CommonController) SendEmail() {

	email := fpc.GetString("email")

	if ok, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, email); ok {

		queryUser := models.User{Email: email}
		exist, err := dao.UserExists(queryUser, "email")
		if err != nil {
			beego.Error("Error occurred in UserExists:", err)
			fpc.CustomAbort(500, "Internal error.")
		}
		if !exist {
			fpc.CustomAbort(404, "email_does_not_exist")
		}

		messageTemplate, err := template.ParseFiles("views/reset-password-mail.tpl")
		if err != nil {
			beego.Error("Parse email template file failed:", err)
			fpc.CustomAbort(500, err.Error())
		}

		message := new(bytes.Buffer)

		harborUrl := os.Getenv("HARBOR_URL")
		if harborUrl == "" {
			harborUrl = "localhost"
		}
		uuid, err := dao.GenerateRandomString()
		if err != nil {
			beego.Error("Error occurred in GenerateRandomString:", err)
			fpc.CustomAbort(500, "Internal error.")
		}
		err = messageTemplate.Execute(message, MessageDetail{
			Hint: fpc.Tr("reset_email_hint"),
			Url:  harborUrl,
			Uuid: uuid,
		})

		if err != nil {
			beego.Error("message template error:", err)
			fpc.CustomAbort(500, "internal_error")
		}

		config, err := beego.AppConfig.GetSection("mail")
		if err != nil {
			beego.Error("Can not load app.conf:", err)
			fpc.CustomAbort(500, "internal_error")
		}

		mail := utils.Mail{
			From:    config["from"],
			To:      []string{email},
			Subject: fpc.Tr("reset_email_subject"),
			Message: message.String()}

		err = mail.SendMail()

		if err != nil {
			beego.Error("send email failed:", err)
			fpc.CustomAbort(500, "send_email_failed")
		}

		user := models.User{ResetUuid: uuid, Email: email}
		dao.UpdateUserResetUuid(user)

	} else {
		fpc.CustomAbort(409, "email_content_illegal")
	}

}

type ResetPasswordController struct {
	BaseController
}

func (rpc *ResetPasswordController) Get() {

	q := rpc.GetString("q")
	queryUser := models.User{ResetUuid: q}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		beego.Error("Error occurred in GetUser:", err)
		rpc.CustomAbort(500, "Internal error.")
	}

	if user != nil {
		rpc.Data["ResetUuid"] = user.ResetUuid
		rpc.ForwardTo("page_title_reset_password", "reset-password")
	} else {
		rpc.Redirect("/", 302)
	}
}

func (rpc *CommonController) ResetPassword() {

	resetUuid := rpc.GetString("reset_uuid")

	queryUser := models.User{ResetUuid: resetUuid}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		beego.Error("Error occurred in GetUser:", err)
		rpc.CustomAbort(500, "Internal error.")
	}

	password := rpc.GetString("password")

	if password != "" {
		user.Password = password
		dao.ResetUserPassword(*user)
	} else {
		rpc.CustomAbort(404, "password_is_required")
	}
}
