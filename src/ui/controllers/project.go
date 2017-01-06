package controllers

import (
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

// ProjectController handles requests to /project
type ProjectController struct {
	BaseController
}

// Get renders project page
func (pc *ProjectController) Get() {
	var err error
	isSysAdmin := false
	uid := pc.GetSession("userId")
	if uid != nil {
		isSysAdmin, err = dao.IsAdminRole(uid)
		if err != nil {
			log.Warningf("Error in checking Admin Role for user, id: %d, error: %v", uid, err)
			isSysAdmin = false
		}
	}
	pc.Data["CanCreate"] = !config.OnlyAdminCreateProject() || isSysAdmin
	pc.Forward("page_title_project", "project.htm")
}
