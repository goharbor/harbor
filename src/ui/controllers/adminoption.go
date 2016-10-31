package controllers

import (
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
)

// AdminOptionController handles requests to /admin_option
type AdminOptionController struct {
	BaseController
}

// Get renders the admin options  page
func (aoc *AdminOptionController) Get() {
	sessionUserID, ok := aoc.GetSession("userId").(int)
	if ok {
		isAdmin, err := dao.IsAdminRole(sessionUserID)
		if err != nil {
			log.Errorf("Error occurred in IsAdminRole: %v", err)
		}
		if isAdmin {
			aoc.Forward("page_title_admin_option", "admin-options.htm")
			return
		}
	}
	aoc.Redirect("/dashboard", 302)
}
