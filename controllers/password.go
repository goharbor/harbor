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
	sessionUserID := cpc.GetSession("userId")
	if sessionUserID == nil {
		cpc.Redirect("/signIn", http.StatusFound)
		return
	}
	cpc.Data["Username"] = cpc.GetSession("username")
	cpc.ForwardTo("page_title_change_password", "change-password")
}

func (cc *CommonController) UpdatePassword() {

	sessionUserID := cc.GetSession("userId")

	if sessionUserID == nil {
		beego.Warning("User does not login.")
		cc.CustomAbort(http.StatusUnauthorized, "please_login_first")
	}

	oldPassword := cc.GetString("old_password")
	if oldPassword == "" {
		beego.Error("Old password is blank")
		cc.CustomAbort(http.StatusBadRequest, "Old password is blank")
	}

	queryUser := models.User{UserID: sessionUserID.(int), Password: oldPassword}
	user, err := dao.CheckUserPassword(queryUser)
	if err != nil {
		beego.Error("Error occurred in CheckUserPassword:", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if user == nil {
		beego.Warning("Password input is not correct")
		cc.CustomAbort(http.StatusForbidden, "old_password_is_not_correct")
	}

	password := cc.GetString("password")
	if password != "" {
		updateUser := models.User{UserID: sessionUserID.(int), Password: password, Salt: user.Salt}
		err = dao.ChangeUserPassword(updateUser, oldPassword)
		if err != nil {
			beego.Error("Error occurred in ChangeUserPassword:", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	} else {
		cc.CustomAbort(http.StatusBadRequest, "please_input_new_password")
	}
}

type ForgotPasswordController struct {
	BaseController
}

type MessageDetail struct {
	Hint string
	URL  string
	UUID string
}

func (fpc *ForgotPasswordController) Get() {
	fpc.ForwardTo("page_title_forgot_password", "forgot-password")
}

func (cc *CommonController) SendEmail() {

	email := cc.GetString("email")

	pass, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, email)

	if !pass {
		cc.CustomAbort(http.StatusBadRequest, "email_content_illegal")
	} else {

		queryUser := models.User{Email: email}
		exist, err := dao.UserExists(queryUser, "email")
		if err != nil {
			beego.Error("Error occurred in UserExists:", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if !exist {
			cc.CustomAbort(http.StatusNotFound, "email_does_not_exist")
		}

		messageTemplate, err := template.ParseFiles("views/reset-password-mail.tpl")
		if err != nil {
			beego.Error("Parse email template file failed:", err)
			cc.CustomAbort(http.StatusInternalServerError, err.Error())
		}

		message := new(bytes.Buffer)

		harborURL := os.Getenv("HARBOR_URL")
		if harborURL == "" {
			harborURL = "localhost"
		}
		uuid, err := dao.GenerateRandomString()
		if err != nil {
			beego.Error("Error occurred in GenerateRandomString:", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		err = messageTemplate.Execute(message, MessageDetail{
			Hint: cc.Tr("reset_email_hint"),
			URL:  harborURL,
			UUID: uuid,
		})

		if err != nil {
			beego.Error("message template error:", err)
			cc.CustomAbort(http.StatusInternalServerError, "internal_error")
		}

		config, err := beego.AppConfig.GetSection("mail")
		if err != nil {
			beego.Error("Can not load app.conf:", err)
			cc.CustomAbort(http.StatusInternalServerError, "internal_error")
		}

		mail := utils.Mail{
			From:    config["from"],
			To:      []string{email},
			Subject: cc.Tr("reset_email_subject"),
			Message: message.String()}

		err = mail.SendMail()

		if err != nil {
			beego.Error("send email failed:", err)
			cc.CustomAbort(http.StatusInternalServerError, "send_email_failed")
		}

		user := models.User{ResetUUID: uuid, Email: email}
		dao.UpdateUserResetUUID(user)

	}

}

type ResetPasswordController struct {
	BaseController
}

func (rpc *ResetPasswordController) Get() {

	resetUUID := rpc.GetString("reset_uuid")
	if resetUUID == "" {
		beego.Error("Reset uuid is blank.")
		rpc.Redirect("/", http.StatusFound)
		return
	}

	queryUser := models.User{ResetUUID: resetUUID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		beego.Error("Error occurred in GetUser:", err)
		rpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if user != nil {
		rpc.Data["ResetUuid"] = user.ResetUUID
		rpc.ForwardTo("page_title_reset_password", "reset-password")
	} else {
		rpc.Redirect("/", http.StatusFound)
	}
}

func (cc *CommonController) ResetPassword() {

	resetUUID := cc.GetString("reset_uuid")
	if resetUUID == "" {
		cc.CustomAbort(http.StatusBadRequest, "Reset uuid is blank.")
	}

	queryUser := models.User{ResetUUID: resetUUID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		beego.Error("Error occurred in GetUser:", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if user == nil {
		beego.Error("User does not exist")
		cc.CustomAbort(http.StatusBadRequest, "User does not exist")
	}

	password := cc.GetString("password")

	if password != "" {
		user.Password = password
		err = dao.ResetUserPassword(*user)
		if err != nil {
			beego.Error("Error occurred in ResetUserPassword:", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	} else {
		cc.CustomAbort(http.StatusBadRequest, "password_is_required")
	}
}
