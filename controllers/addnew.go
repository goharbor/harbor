package controllers

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

// AddNewController handles requests to /add_new
type AddNewController struct {
	BaseController
}

// Get renders the add new page
func (anc *AddNewController) Get() {
	sessionUserID := anc.GetSession("userId")
	anc.Data["AddNew"] = false
	if sessionUserID != nil {
		isAdmin, err := dao.IsAdminRole(sessionUserID.(int))
		if err != nil {
			log.Errorf("Error occurred in IsAdminRole: %v", err)
			anc.CustomAbort(http.StatusInternalServerError, "")
		}
		if isAdmin && anc.AuthMode == "db_auth" {
			anc.Data["AddNew"] = true
			anc.Forward("page_title_add_new", "sign-up.htm")
			return
		}
	}
	anc.CustomAbort(http.StatusUnauthorized, "Status Unauthorized.")
}
