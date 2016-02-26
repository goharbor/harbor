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
package dao

import (
	"github.com/vmware/harbor/models"

	"errors"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

//TODO:transaction, return err

// AddProject adds a project to the database along with project roles information and access log records.
func AddProject(project models.Project) error {

	if isIllegalLength(project.Name, 4, 30) {
		return errors.New("project name is illegal in length. (greater than 4 or less than 30)")
	}
	if isContainIllegalChar(project.Name, []string{"~", "-", "$", "\\", "[", "]", "{", "}", "(", ")", "&", "^", "%", "*", "<", ">", "\"", "'", "/", "?", "@"}) {
		return errors.New("project name contains illegal characters")
	}

	o := orm.NewOrm()

	p, err := o.Raw("insert into project (owner_id, name, creation_time, deleted, public) values (?, ?, now(), ?, ?)").Prepare()
	if err != nil {
		return err
	}

	r, err := p.Exec(project.OwnerID, project.Name, project.Deleted, project.Public)
	if err != nil {
		return err
	}

	projectID, err := r.LastInsertId()
	if err != nil {
		return err
	}

	projectAdminRole := models.ProjectRole{ProjectID: projectID, RoleID: models.PROJECTADMIN}
	_, err = AddProjectRole(projectAdminRole)
	if err != nil {
		return err
	}

	projectDeveloperRole := models.ProjectRole{ProjectID: projectID, RoleID: models.DEVELOPER}
	_, err = AddProjectRole(projectDeveloperRole)
	if err != nil {
		return err
	}

	projectGuestRole := models.ProjectRole{ProjectID: projectID, RoleID: models.GUEST}
	_, err = AddProjectRole(projectGuestRole)
	if err != nil {
		return err
	}

	//Add all project roles, after that when assigning a user to a project just update the upr table
	err = AddUserProjectRole(project.OwnerID, projectID, models.PROJECTADMIN)
	if err != nil {
		return err
	}

	accessLog := models.AccessLog{UserID: project.OwnerID, ProjectID: projectID, RepoName: project.Name + "/", GUID: "N/A", Operation: "create", OpTime: time.Now()}
	err = AddAccessLog(accessLog)

	return err
}

// IsProjectPublic ...
func IsProjectPublic(projectName string) bool {
	project, err := GetProjectByName(projectName)
	if err != nil {
		beego.Error("Error occurred in GetProjectByName:", err)
		return false
	}
	if project == nil {
		return false
	}
	return project.Public == 1
}

// QueryProject querys the projects based on publicity and user, disregarding the names etc.
func QueryProject(query models.Project) ([]models.Project, error) {
	o := orm.NewOrm()

	sql := `select distinct
		p.project_id, p.owner_id, p.name,p.creation_time, p.public 
	 from project p 
		left join project_role pr on p.project_id = pr.project_id
	   left join user_project_role upr on upr.pr_id = pr.pr_id
	   left join user u on u.user_id = upr.user_id
	 where p.deleted = 0 `

	queryParam := make([]interface{}, 1)

	if query.Public == 1 {
		sql += ` and p.public = ?`
		queryParam = append(queryParam, query.Public)
	} else if isAdmin, _ := IsAdminRole(query.UserID); isAdmin == false {
		sql += ` and (p.owner_id = ? or u.user_id = ?) `
		queryParam = append(queryParam, query.UserID)
		queryParam = append(queryParam, query.UserID)
	}

	if query.Name != "" {
		sql += " and p.name like ? "
		queryParam = append(queryParam, query.Name)
	}

	sql += " order by p.name "

	var r []models.Project
	_, err := o.Raw(sql, queryParam).QueryRows(&r)

	if err != nil {
		return nil, err
	}
	return r, nil
}

//ProjectExists returns whether the project exists according to its name of ID.
func ProjectExists(nameOrID interface{}) (bool, error) {
	o := orm.NewOrm()
	type dummy struct{}
	sql := `select project_id from project where deleted = 0 and `
	switch nameOrID.(type) {
	case int64:
		sql += `project_id = ?`
	case string:
		sql += `name = ?`
	default:
		return false, fmt.Errorf("Invalid nameOrId: %v", nameOrID)
	}

	var d []dummy
	num, err := o.Raw(sql, nameOrID).QueryRows(&d)
	if err != nil {
		return false, err
	}
	return num > 0, nil

}

// GetProjectByID ...
func GetProjectByID(projectID int64) (*models.Project, error) {
	o := orm.NewOrm()

	sql := `select p.project_id, p.name, u.username as owner_name, p.owner_id, p.creation_time, p.public  
		from project p left join user u on p.owner_id = u.user_id where p.deleted = 0 and p.project_id = ?`
	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, projectID)

	p := []models.Project{}
	count, err := o.Raw(sql, queryParam).QueryRows(&p)

	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, nil
	} else {
		return &p[0], nil
	}
}

// GetProjectByName ...
func GetProjectByName(projectName string) (*models.Project, error) {
	o := orm.NewOrm()
	var p []models.Project
	n, err := o.Raw(`select project_id, owner_id, name, deleted, public from project where name = ? and deleted = 0`, projectName).QueryRows(&p)
	if err != nil {
		return nil, err
	} else if n == 0 {
		return nil, nil
	} else {
		return &p[0], nil
	}
}

// GetPermission gets roles that the user has according to the project.
func GetPermission(username, projectName string) (string, error) {
	o := orm.NewOrm()

	sql := "select r.role_code from role as r " +
		"inner join project_role as pr on r.role_id = pr.role_id " +
		"inner join user_project_role as ur on pr.pr_id = ur.pr_id " +
		"inner join user as u on u.user_id = ur.user_id " +
		"inner join project p on p.project_id = pr.project_id " +
		"where u.username = ? and p.name = ? and u.deleted = 0 and p.deleted = 0"

	var r []models.Role
	n, err := o.Raw(sql, username, projectName).QueryRows(&r)
	if err != nil {
		return "", err
	} else if n == 0 {
		return "", nil
	} else {
		return r[0].RoleCode, nil
	}
}

// ToggleProjectPublicity toggles the publicity of the project.
func ToggleProjectPublicity(projectID int64, publicity int) error {
	o := orm.NewOrm()
	sql := "update project set public = ? where project_id = ?"
	_, err := o.Raw(sql, publicity, projectID).Exec()
	return err
}

// QueryRelevantProjects returns all projects that the user is a member of.
func QueryRelevantProjects(userID int) ([]models.Project, error) {
	o := orm.NewOrm()
	sql := `SELECT distinct p.project_id, p.name, p.public FROM registry.project p 
		left join project_role pr on p.project_id = pr.project_id 
		left join user_project_role upr on upr.pr_id = pr.pr_id 
		where upr.user_id = ? or p.public = 1 and p.deleted = 0`
	var res []models.Project
	_, err := o.Raw(sql, userID).QueryRows(&res)
	if err != nil {
		return nil, err
	}
	return res, err
}
