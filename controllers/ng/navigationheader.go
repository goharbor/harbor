package ng

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

type NavigationHeaderController struct {
	BaseController
}

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
	nhc.TplName = "ng/navigation-header.htm"
	nhc.Render()
}
