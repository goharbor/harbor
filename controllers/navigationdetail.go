package controllers

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// NavigationDetailController handles requests to /navigation_detail
type NavigationDetailController struct {
	BaseController
}

// Get renders user's navigation details header
func (ndc *NavigationDetailController) Get() {
	sessionUserID := ndc.GetSession("userId")
	var isAdmin int
	if sessionUserID != nil {
		userID := sessionUserID.(int)
		u, err := dao.GetUser(models.User{UserID: userID})
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			ndc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			log.Warningf("User was deleted already, user id: %d, canceling request.", userID)
			ndc.CustomAbort(http.StatusUnauthorized, "")
		}
		isAdmin = u.HasAdminRole
	}
	ndc.Data["IsAdmin"] = isAdmin
	ndc.TplName = "navigation-detail.htm"
	ndc.Render()
}
