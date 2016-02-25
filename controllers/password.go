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
	"net/http"
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
		cpc.Redirect("/signIn", http.StatusFound)
		return
	}
	cpc.Data["Username"] = cpc.GetSession("username")
	cpc.ForwardTo("page_title_change_password", "change-password")
}

func (cpc *CommonController) UpdatePassword() {

	sessionUserId := cpc.GetSession("userId")

	if sessionUserId == nil {
		beego.Warning("User does not login.")
		cpc.CustomAbort(http.StatusUnauthorized, "please_login_first")
	}

	oldPassword := cpc.GetString("old_password")
	if oldPassword == "" {
		beego.Error("Old password is blank")
		cpc.CustomAbort(http.StatusBadRequest, "Old password is blank")
	}

	queryUser := models.User{UserId: sessionUserId.(int), Password: oldPassword}
	user, err := dao.CheckUserPassword(queryUser)
	if err != nil {
		beego.Error("Error occurred in CheckUserPassword:", err)
		cpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if user == nil {
		beego.Warning("Password input is not correct")
		cpc.CustomAbort(http.StatusForbidden, "old_password_is_not_correct")
	}

	password := cpc.GetString("password")
	if password != "" {
		updateUser := models.User{UserId: sessionUserId.(int), Password: password, Salt: user.Salt}
		err = dao.ChangeUserPassword(updateUser, oldPassword)
		if err != nil {
			beego.Error("Error occurred in ChangeUserPassword:", err)
			cpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	} else {
		cpc.CustomAbort(http.StatusBadRequest, "please_input_new_password")
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

	pass, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, email)

	if !pass {
		fpc.CustomAbort(http.StatusBadRequest, "email_content_illegal")
	} else {

		queryUser := models.User{Email: email}
		exist, err := dao.UserExists(queryUser, "email")
		if err != nil {
			beego.Error("Error occurred in UserExists:", err)
			fpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if !exist {
			fpc.CustomAbort(http.StatusNotFound, "email_does_not_exist")
		}

		messageTemplate, err := template.ParseFiles("views/reset-password-mail.tpl")
		if err != nil {
			beego.Error("Parse email template file failed:", err)
			fpc.CustomAbort(http.StatusInternalServerError, err.Error())
		}

		message := new(bytes.Buffer)

		harborUrl := os.Getenv("HARBOR_URL")
		if harborUrl == "" {
			harborUrl = "localhost"
		}
		uuid, err := dao.GenerateRandomString()
		if err != nil {
			beego.Error("Error occurred in GenerateRandomString:", err)
			fpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		err = messageTemplate.Execute(message, MessageDetail{
			Hint: fpc.Tr("reset_email_hint"),
			Url:  harborUrl,
			Uuid: uuid,
		})

		if err != nil {
			beego.Error("message template error:", err)
			fpc.CustomAbort(http.StatusInternalServerError, "internal_error")
		}

		config, err := beego.AppConfig.GetSection("mail")
		if err != nil {
			beego.Error("Can not load app.conf:", err)
			fpc.CustomAbort(http.StatusInternalServerError, "internal_error")
		}

		mail := utils.Mail{
			From:    config["from"],
			To:      []string{email},
			Subject: fpc.Tr("reset_email_subject"),
			Message: message.String()}

		err = mail.SendMail()

		if err != nil {
			beego.Error("send email failed:", err)
			fpc.CustomAbort(http.StatusInternalServerError, "send_email_failed")
		}

		user := models.User{ResetUuid: uuid, Email: email}
		dao.UpdateUserResetUuid(user)

	}

}

type ResetPasswordController struct {
	BaseController
}

func (rpc *ResetPasswordController) Get() {

	resetUuid := rpc.GetString("reset_uuid")
	if resetUuid == "" {
		beego.Error("Reset uuid is blank.")
		rpc.Redirect("/", http.StatusFound)
		return
	}

	queryUser := models.User{ResetUuid: resetUuid}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		beego.Error("Error occurred in GetUser:", err)
		rpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if user != nil {
		rpc.Data["ResetUuid"] = user.ResetUuid
		rpc.ForwardTo("page_title_reset_password", "reset-password")
	} else {
		rpc.Redirect("/", http.StatusFound)
	}
}

func (rpc *CommonController) ResetPassword() {

	resetUuid := rpc.GetString("reset_uuid")
	if resetUuid == "" {
		rpc.CustomAbort(http.StatusBadRequest, "Reset uuid is blank.")
	}

	queryUser := models.User{ResetUuid: resetUuid}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		beego.Error("Error occurred in GetUser:", err)
		rpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if user == nil {
		beego.Error("User does not exist")
		rpc.CustomAbort(http.StatusBadRequest, "User does not exist")
	}

	password := rpc.GetString("password")

	if password != "" {
		user.Password = password
		err = dao.ResetUserPassword(*user)
		if err != nil {
			beego.Error("Error occurred in ResetUserPassword:", err)
			rpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	} else {
		rpc.CustomAbort(http.StatusBadRequest, "password_is_required")
	}
}
