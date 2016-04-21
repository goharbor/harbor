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
	"github.com/vmware/harbor/utils/log"

	"github.com/astaxie/beego"
)

// ChangePasswordController handles request to /changePassword
type ChangePasswordController struct {
	BaseController
}

// Get renders the page for user to change password.
func (cpc *ChangePasswordController) Get() {
	sessionUserID := cpc.GetSession("userId")
	if sessionUserID == nil {
		cpc.Redirect("/signIn", http.StatusFound)
		return
	}
	cpc.Data["Username"] = cpc.GetSession("username")
	cpc.ForwardTo("page_title_change_password", "change-password")
}

// ForgotPasswordController handles request to /forgotPassword
type ForgotPasswordController struct {
	BaseController
}

// Get Renders the page for user to input Email to reset password.
func (fpc *ForgotPasswordController) Get() {
	fpc.ForwardTo("page_title_forgot_password", "forgot-password")
}

type messageDetail struct {
	Hint string
	URL  string
	UUID string
}

// SendEmail verifies the Email address and contact SMTP server to send reset password Email.
func (cc *CommonController) SendEmail() {

	email := cc.GetString("email")

	pass, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, email)

	if !pass {
		cc.CustomAbort(http.StatusBadRequest, "email_content_illegal")
	} else {

		queryUser := models.User{Email: email}
		exist, err := dao.UserExists(queryUser, "email")
		if err != nil {
			log.Errorf("Error occurred in UserExists: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if !exist {
			cc.CustomAbort(http.StatusNotFound, "email_does_not_exist")
		}

		messageTemplate, err := template.ParseFiles("views/reset-password-mail.tpl")
		if err != nil {
			log.Errorf("Parse email template file failed: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, err.Error())
		}

		message := new(bytes.Buffer)

		harborURL := os.Getenv("HARBOR_URL")
		if harborURL == "" {
			harborURL = "localhost"
		}
		uuid, err := dao.GenerateRandomString()
		if err != nil {
			log.Errorf("Error occurred in GenerateRandomString: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		err = messageTemplate.Execute(message, messageDetail{
			Hint: cc.Tr("reset_email_hint"),
			URL:  harborURL,
			UUID: uuid,
		})

		if err != nil {
			log.Errorf("Message template error: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "internal_error")
		}

		config, err := beego.AppConfig.GetSection("mail")
		if err != nil {
			log.Errorf("Can not load app.conf: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "internal_error")
		}

		mail := utils.Mail{
			From:    config["from"],
			To:      []string{email},
			Subject: cc.Tr("reset_email_subject"),
			Message: message.String()}

		err = mail.SendMail()

		if err != nil {
			log.Errorf("Send email failed: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "send_email_failed")
		}

		user := models.User{ResetUUID: uuid, Email: email}
		dao.UpdateUserResetUUID(user)

	}

}

// ResetPasswordController handles request to /resetPassword
type ResetPasswordController struct {
	BaseController
}

// Get checks if reset_uuid in the reset link is valid and render the result page for user to reset password.
func (rpc *ResetPasswordController) Get() {

	resetUUID := rpc.GetString("reset_uuid")
	if resetUUID == "" {
		log.Error("Reset uuid is blank.")
		rpc.Redirect("/", http.StatusFound)
		return
	}

	queryUser := models.User{ResetUUID: resetUUID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		log.Errorf("Error occurred in GetUser: %v", err)
		rpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if user != nil {
		rpc.Data["ResetUuid"] = user.ResetUUID
		rpc.ForwardTo("page_title_reset_password", "reset-password")
	} else {
		rpc.Redirect("/", http.StatusFound)
	}
}

// ResetPassword handles request from the reset page and reset password
func (cc *CommonController) ResetPassword() {

	resetUUID := cc.GetString("reset_uuid")
	if resetUUID == "" {
		cc.CustomAbort(http.StatusBadRequest, "Reset uuid is blank.")
	}

	queryUser := models.User{ResetUUID: resetUUID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		log.Errorf("Error occurred in GetUser: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if user == nil {
		log.Error("User does not exist")
		cc.CustomAbort(http.StatusBadRequest, "User does not exist")
	}

	password := cc.GetString("password")

	if password != "" {
		user.Password = password
		err = dao.ResetUserPassword(*user)
		if err != nil {
			log.Errorf("Error occurred in ResetUserPassword: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	} else {
		cc.CustomAbort(http.StatusBadRequest, "password_is_required")
	}
}
