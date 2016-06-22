package controllers

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

type AddNewController struct {
	BaseController
}

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
			anc.Forward("Add User", "sign-up.htm")
			return
		}
	}
	anc.CustomAbort(http.StatusUnauthorized, "Status Unauthorized.")
}
