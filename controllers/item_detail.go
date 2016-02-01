/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package controllers

import (
	"net/url"
	"os"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
)

type ItemDetailController struct {
	BaseController
}

var SYS_ADMIN int = 1
var PROJECT_ADMIN int = 2
var DEVELOPER int = 3
var GUEST int = 4

func CheckProjectRole(userId int, projectId int64) bool {
	if projectId == 0 {
		return false
	}
	userQuery := models.User{UserId: int(userId)}
	if userId == SYS_ADMIN {
		return true
	}
	roleList, err := dao.GetUserProjectRoles(userQuery, projectId)
	if err != nil {
		beego.Error("Error occurred in GetUserProjectRoles:", err)
		return false
	}
	return len(roleList) > 0
}

func CheckPublicProject(projectId int64) bool {
	projectQuery := models.Project{ProjectId: projectId}
	project, err := dao.GetProjectById(projectQuery)
	if err != nil {
		beego.Error("Error occurred in GetProjectById:", err)
		return false
	}
	if project != nil && project.Public == 1 {
		return true
	}
	return false
}

func (idc *ItemDetailController) Get() {

	sessionUserId := idc.GetSession("userId")
	projectId, _ := idc.GetInt64("project_id")

	if CheckPublicProject(projectId) == false && (sessionUserId == nil || !CheckProjectRole(sessionUserId.(int), projectId)) {
		idc.Redirect("/signIn?uri="+url.QueryEscape(idc.Ctx.Input.URI()), 302)
	}

	projectQuery := models.Project{ProjectId: projectId}
	project, err := dao.GetProjectById(projectQuery)

	if err != nil {
		beego.Error("Error occurred in GetProjectById:", err)
		idc.CustomAbort(500, "Internal error.")
	}

	if project == nil {
		idc.Redirect("/signIn", 302)
	}

	idc.Data["ProjectId"] = project.ProjectId
	idc.Data["ProjectName"] = project.Name
	idc.Data["OwnerName"] = project.OwnerName
	idc.Data["OwnerId"] = project.OwnerId

	if sessionUserId != nil {
		idc.Data["Username"] = idc.GetSession("username")
		idc.Data["UserId"] = sessionUserId.(int)
		roleList, err := dao.GetUserProjectRoles(models.User{UserId: sessionUserId.(int)}, projectId)
		if err != nil {
			beego.Error("Error occurred in GetUserProjectRoles:", err)
			idc.CustomAbort(500, "Internal error.")
		}
		if len(roleList) > 0 {
			idc.Data["RoleId"] = roleList[0].RoleId
		}
	}

	idc.Data["HarborRegUrl"] = os.Getenv("HARBOR_REG_URL")
	idc.Data["RepoName"] = idc.GetString("repo_name")

	idc.ForwardTo("page_title_item_details", "item-detail")

}
