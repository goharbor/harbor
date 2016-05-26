package ng

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// DashboardController handles requests to /ng/dashboard
type DashboardController struct {
	BaseController
}

// Get renders the dashboard  page
func (dc *DashboardController) Get() {
	sessionUserID := dc.GetSession("userId")
	var isAdmin int

	if sessionUserID != nil {
		userID := sessionUserID.(int)
		u, err := dao.GetUser(models.User{UserID: userID})
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			dc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			log.Warningf("User was deleted already, user id: %d, canceling request.", userID)
			dc.CustomAbort(http.StatusUnauthorized, "")
		}
		isAdmin = u.HasAdminRole
	}

	dc.Data["IsAdmin"] = isAdmin

	dc.Forward("Dashboard", "dashboard.htm")
}
