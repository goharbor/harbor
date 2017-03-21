package controllers

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"regexp"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/email"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
)

// CommonController handles request from UI that doesn't expect a page, such as /SwitchLanguage /logout ...
type CommonController struct {
	beego.Controller
	i18n.Locale
}

// Render returns nil.
func (cc *CommonController) Render() error {
	return nil
}

type messageDetail struct {
	Hint string
	URL  string
	UUID string
}

// Login handles login request from UI.
func (cc *CommonController) Login() {
	principal := cc.GetString("principal")
	password := cc.GetString("password")

	user, err := auth.Login(models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	if user == nil {
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	cc.SetSession("userId", user.UserID)
	cc.SetSession("username", user.Username)
}

// LogOut Habor UI
func (cc *CommonController) LogOut() {
	cc.DestroySession()
}

// UserExists checks if user exists when user input value in sign in form.
func (cc *CommonController) UserExists() {
	target := cc.GetString("target")
	value := cc.GetString("value")

	user := models.User{}
	switch target {
	case "username":
		user.Username = value
	case "email":
		user.Email = value
	}

	exist, err := dao.UserExists(user, target)
	if err != nil {
		log.Errorf("Error occurred in UserExists: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	cc.Data["json"] = exist
	cc.ServeJSON()
}

// SendEmail verifies the Email address and contact SMTP server to send reset password Email.
func (cc *CommonController) SendEmail() {

	emailStr := cc.GetString("email")

	pass, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, emailStr)

	if !pass {
		cc.CustomAbort(http.StatusBadRequest, "email_content_illegal")
	} else {

		queryUser := models.User{Email: emailStr}
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

		harborURL := common.ExtEndpoint
		if harborURL == "" {
			harborURL = "localhost"
		}
		uuid := utils.GenerateRandomString()
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

		mail := email.Mail{
			From:    config["from"],
			To:      []string{emailStr},
			Subject: cc.Tr("reset_email_subject"),
			Message: message.String()}

		err = mail.SendMail()

		if err != nil {
			log.Errorf("Send email failed: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "send_email_failed")
		}

		user := models.User{ResetUUID: uuid, Email: emailStr}
		dao.UpdateUserResetUUID(user)

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

func init() {
	//conf/app.conf -> os.Getenv("config_path")
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		log.Infof("Config path: %s", configPath)
		if err := beego.LoadAppConfig("ini", configPath); err != nil {
			log.Errorf("failed to load app config: %v", err)
		}
	}

}
