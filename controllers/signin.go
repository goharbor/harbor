package controllers

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// SignInController handles requests to /sign_in
type SignInController struct {
	BaseController
}

//Get renders sign_in page
func (sic *SignInController) Get() {
	sessionUserID := sic.GetSession("userId")
	var hasLoggedIn bool
	var username string
	if sessionUserID != nil {
		hasLoggedIn = true
		userID := sessionUserID.(int)
		u, err := dao.GetUser(models.User{UserID: userID})
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			sic.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			log.Warningf("User was deleted already, user id: %d, canceling request.", userID)
			sic.CustomAbort(http.StatusUnauthorized, "")
		}
		username = u.Username
	}
	sic.Data["AuthMode"] = sic.AuthMode
	sic.Data["Username"] = username
	sic.Data["HasLoggedIn"] = hasLoggedIn
	sic.TplName = "sign-in.htm"
	sic.Render()
}
