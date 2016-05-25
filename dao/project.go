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

	"fmt"
	"time"

	"github.com/vmware/harbor/utils/log"
)

//TODO:transaction, return err

// AddProject adds a project to the database along with project roles information and access log records.
func AddProject(project models.Project) (int64, error) {

	o := GetOrmer()
	p, err := o.Raw("insert into project (owner_id, name, creation_time, update_time, deleted, public) values (?, ?, ?, ?, ?, ?)").Prepare()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	r, err := p.Exec(project.OwnerID, project.Name, now, now, project.Deleted, project.Public)
	if err != nil {
		return 0, err
	}

	projectID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = AddProjectMember(projectID, project.OwnerID, models.PROJECTADMIN); err != nil {
		return projectID, err
	}

	accessLog := models.AccessLog{UserID: project.OwnerID, ProjectID: projectID, RepoName: project.Name + "/", RepoTag: "N/A", GUID: "N/A", Operation: "create", OpTime: time.Now()}
	err = AddAccessLog(accessLog)

	return projectID, err
}

// IsProjectPublic ...
func IsProjectPublic(projectName string) bool {
	project, err := GetProjectByName(projectName)
	if err != nil {
		log.Errorf("Error occurred in GetProjectByName: %v", err)
		return false
	}
	if project == nil {
		return false
	}
	return project.Public == 1
}

//ProjectExists returns whether the project exists according to its name of ID.
func ProjectExists(nameOrID interface{}) (bool, error) {
	o := GetOrmer()
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
func GetProjectByID(id int64) (*models.Project, error) {
	o := GetOrmer()

	sql := `select p.project_id, p.name, u.username as owner_name, p.owner_id, p.creation_time, p.update_time, p.public  
		from project p left join user u on p.owner_id = u.user_id where p.deleted = 0 and p.project_id = ?`
	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, id)

	p := []models.Project{}
	count, err := o.Raw(sql, queryParam).QueryRows(&p)

	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	return &p[0], nil
}

// GetProjectByName ...
func GetProjectByName(name string) (*models.Project, error) {
	o := GetOrmer()
	var p []models.Project
	n, err := o.Raw(`select * from project where name = ? and deleted = 0`, name).QueryRows(&p)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, nil
	}

	return &p[0], nil
}

// GetPermission gets roles that the user has according to the project.
func GetPermission(username, projectName string) (string, error) {
	o := GetOrmer()

	sql := `select r.role_code from role as r
		inner join project_member as pm on r.role_id = pm.role
		inner join user as u on u.user_id = pm.user_id
		inner join project p on p.project_id = pm.project_id
		where u.username = ? and p.name = ? and u.deleted = 0 and p.deleted = 0`

	var r []models.Role
	n, err := o.Raw(sql, username, projectName).QueryRows(&r)
	if err != nil {
		return "", err
	}

	if n == 0 {
		return "", nil
	}

	return r[0].RoleCode, nil
}

// ToggleProjectPublicity toggles the publicity of the project.
func ToggleProjectPublicity(projectID int64, publicity int) error {
	o := GetOrmer()
	sql := "update project set public = ? where project_id = ?"
	_, err := o.Raw(sql, publicity, projectID).Exec()
	return err
}

// SearchProjects returns a project list,
// which satisfies the following conditions:
// 1. the project is not deleted
// 2. the prject is public or the user is a member of the project
func SearchProjects(userID int) ([]models.Project, error) {
	o := GetOrmer()
	sql := `select distinct p.project_id, p.name, p.public 
		from project p 
		left join project_member pm on p.project_id = pm.project_id 
		where (pm.user_id = ? or p.public = 1) and p.deleted = 0`

	var projects []models.Project

	if _, err := o.Raw(sql, userID).QueryRows(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// GetUserRelevantProjects returns the projects of the user which are not deleted and name like projectName
func GetUserRelevantProjects(userID int, projectName string) ([]models.Project, error) {
	o := GetOrmer()
	sql := `select distinct
		p.project_id, p.owner_id, p.name,p.creation_time, p.update_time, p.public, pm.role role 
	 from project p 
		left join project_member pm on p.project_id = pm.project_id
	 where p.deleted = 0 and pm.user_id= ?`

	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, userID)
	if projectName != "" {
		sql += " and p.name like ? "
		queryParam = append(queryParam, projectName)
	}
	sql += " order by p.name "
	var r []models.Project
	_, err := o.Raw(sql, queryParam).QueryRows(&r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//GetPublicProjects returns all public projects whose name like projectName
func GetPublicProjects(projectName string) ([]models.Project, error) {
	publicProjects, err := getProjects(1, projectName)
	if err != nil {
		return nil, err
	}
	return publicProjects, nil
}

// GetAllProjects returns all projects which are not deleted and name like projectName
func GetAllProjects(projectName string) ([]models.Project, error) {
	allProjects, err := getProjects(0, projectName)
	if err != nil {
		return nil, err
	}
	return allProjects, nil
}

func getProjects(public int, projectName string) ([]models.Project, error) {
	o := GetOrmer()
	sql := `select project_id, owner_id, creation_time, update_time, name, public 
		from project
		where deleted = 0`
	queryParam := make([]interface{}, 1)
	if public == 1 {
		sql += " and public = ? "
		queryParam = append(queryParam, public)
	}
	if len(projectName) > 0 {
		sql += " and name like ? "
		queryParam = append(queryParam, projectName)
	}
	sql += " order by name "
	var projects []models.Project
	if _, err := o.Raw(sql, queryParam).QueryRows(&projects); err != nil {
		return nil, err
	}
	return projects, nil
}
