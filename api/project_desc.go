package api

import (
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

type ProjectDescController struct {
	BaseAPI
}

type ProjectDescReq struct {
 	ProjectId int     `json:"project_id"`
	Name      string   `json:"name"`
}

/*
 Method: Put
 https://registry.51yixiao.com/api/project_desc/3
 param: name
 */
func (c *ProjectDescController) UpdateProject() {

	var req ProjectDescReq
	c.DecodeJSONReq(&req)

	m := models.ProjectDesc{}
	m.ProjectId = req.ProjectId
	m.Name = req.Name

	log.Infof("m: %+v",m)
	if err := dao.UpdateProjectById(m); err == nil {
			c.Data["json"] = "OK"
	} else {
		c.Data["json"] = err
	}

	c.ServeJSON()
}
