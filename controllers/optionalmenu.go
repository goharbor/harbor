package controllers

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// OptionalMenuController handles request to /optional_menu
type OptionalMenuController struct {
	BaseController
}

// Get renders optional menu, Admin user has "Add User" menu
func (omc *OptionalMenuController) Get() {
	sessionUserID := omc.GetSession("userId")

	var hasLoggedIn bool
	var allowAddNew bool

	if sessionUserID != nil {
		hasLoggedIn = true
		userID := sessionUserID.(int)
		u, err := dao.GetUser(models.User{UserID: userID})
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			omc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			log.Warningf("User was deleted already, user id: %d, canceling request.", userID)
			omc.CustomAbort(http.StatusUnauthorized, "")
		}
		omc.Data["Username"] = u.Username

		isAdmin, err := dao.IsAdminRole(sessionUserID.(int))
		if err != nil {
			log.Errorf("Error occurred in IsAdminRole: %v", err)
			omc.CustomAbort(http.StatusInternalServerError, "")
		}

		if isAdmin && omc.AuthMode == "db_auth" {
			allowAddNew = true
		}
	}
	omc.Data["AddNew"] = allowAddNew
	omc.Data["HasLoggedIn"] = hasLoggedIn
	omc.TplName = "optional-menu.htm"
	omc.Render()

}
