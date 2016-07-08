package controllers

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// NavigationHeaderController handles requests to /navigation_header
type NavigationHeaderController struct {
	BaseController
}

// Get renders user's navigation header
func (nhc *NavigationHeaderController) Get() {
	sessionUserID := nhc.GetSession("userId")
	var hasLoggedIn bool
	var isAdmin int
	if sessionUserID != nil {
		hasLoggedIn = true
		userID := sessionUserID.(int)
		u, err := dao.GetUser(models.User{UserID: userID})
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			nhc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			log.Warningf("User was deleted already, user id: %d, canceling request.", userID)
			nhc.CustomAbort(http.StatusUnauthorized, "")
		}
		isAdmin = u.HasAdminRole
	}
	nhc.Data["HasLoggedIn"] = hasLoggedIn
	nhc.Data["IsAdmin"] = isAdmin
	nhc.TplName = "navigation-header.htm"
	nhc.Render()
}
