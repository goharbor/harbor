// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"github.com/vmware/harbor/src/common/models"

	"fmt"
	"time"

	"github.com/vmware/harbor/src/common/utils/log"
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

	sql :=
		`select distinct p.project_id, p.name, p.public, 
			p.owner_id, p.creation_time, p.update_time
		from project p 
		left join project_member pm 
		on p.project_id = pm.project_id 
		where (pm.user_id = ? or p.public = 1) 
		and p.deleted = 0 `

	var projects []models.Project

	if _, err := o.Raw(sql, userID).QueryRows(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

//GetTotalOfUserRelevantProjects returns the total count of
// user relevant projects
func GetTotalOfUserRelevantProjects(userID int, projectName string) (int64, error) {
	o := GetOrmer()
	sql := `select count(*) from project p 
			left join project_member pm 
			on p.project_id = pm.project_id
	 		where p.deleted = 0 and pm.user_id= ?`

	queryParam := []interface{}{}
	queryParam = append(queryParam, userID)
	if projectName != "" {
		sql += " and p.name like ? "
		queryParam = append(queryParam, "%"+escape(projectName)+"%")
	}

	var total int64
	err := o.Raw(sql, queryParam).QueryRow(&total)

	return total, err
}

// GetUserRelevantProjects returns the user relevant projects
// args[0]: public, args[1]: limit, args[2]: offset
func GetUserRelevantProjects(userID int, projectName string, args ...int64) ([]models.Project, error) {
	return getProjects(userID, projectName, args...)
}

// GetTotalOfProjects returns the total count of projects
func GetTotalOfProjects(name string, public ...int) (int64, error) {
	qs := GetOrmer().
		QueryTable(new(models.Project)).
		Filter("Deleted", 0)

	if len(name) > 0 {
		qs = qs.Filter("Name__icontains", name)
	}

	if len(public) > 0 {
		qs = qs.Filter("Public", public[0])
	}

	return qs.Count()
}

// GetProjects returns project list
// args[0]: public, args[1]: limit, args[2]: offset
func GetProjects(name string, args ...int64) ([]models.Project, error) {
	return getProjects(0, name, args...)
}

func getProjects(userID int, name string, args ...int64) ([]models.Project, error) {
	projects := []models.Project{}

	o := GetOrmer()
	sql := ""
	queryParam := []interface{}{}

	if userID != 0 { //get user's projects
		sql = `select distinct p.project_id, p.owner_id, p.name, 
					p.creation_time, p.update_time, p.public, pm.role role 
			from project p 
			left join project_member pm 
			on p.project_id = pm.project_id
	 		where p.deleted = 0 and pm.user_id= ?`
		queryParam = append(queryParam, userID)
	} else { // get all projects
		sql = `select * from project p where p.deleted = 0 `
	}

	if name != "" {
		sql += ` and p.name like ? `
		queryParam = append(queryParam, "%"+escape(name)+"%")
	}

	switch len(args) {
	case 1:
		sql += ` and p.public = ?`
		queryParam = append(queryParam, args[0])
		sql += ` order by p.name `
	case 2:
		sql += ` order by p.name `
		sql = paginateForRawSQL(sql, args[0], args[1])
	case 3:
		sql += ` and p.public = ?`
		queryParam = append(queryParam, args[0])
		sql += ` order by p.name `
		sql = paginateForRawSQL(sql, args[1], args[2])
	}

	_, err := o.Raw(sql, queryParam).QueryRows(&projects)

	return projects, err
}

// DeleteProject ...
func DeleteProject(id int64) error {
	project, err := GetProjectByID(id)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s#%d", project.Name, project.ProjectID)

	sql := `update project 
		set deleted = 1, name = ? 
		where project_id = ?`
	_, err = GetOrmer().Raw(sql, name, id).Exec()
	return err
}
