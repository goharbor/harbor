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
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"

	"fmt"
	"time"
)

// AddProject adds a project to the database along with project roles information and access log records.
func AddProject(project models.Project) (int64, error) {

	o := GetOrmer()
	p, err := o.Raw("insert into project (owner_id, name, creation_time, update_time, deleted) values (?, ?, ?, ?, ?)").Prepare()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	r, err := p.Exec(project.OwnerID, project.Name, now, now, project.Deleted)
	if err != nil {
		return 0, err
	}

	projectID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	pmID, err := addProjectMember(models.Member{
		ProjectID:  projectID,
		EntityID:   project.OwnerID,
		Role:       models.PROJECTADMIN,
		EntityType: common.UserMember,
	})
	if err != nil {
		return 0, err
	}
	if pmID == 0 {
		return projectID, fmt.Errorf("Failed to add project member, pmid=0")
	}
	return projectID, err
}

func addProjectMember(member models.Member) (int, error) {

	log.Debugf("Adding project member %+v", member)

	o := GetOrmer()

	if member.EntityID <= 0 {
		return 0, fmt.Errorf("Invalid entity_id, member: %+v", member)
	}

	if member.ProjectID <= 0 {
		return 0, fmt.Errorf("Invalid project_id, member: %+v", member)
	}

	sql := "insert into project_member (project_id, entity_id , role, entity_type) values (?, ?, ?, ?)"
	r, err := o.Raw(sql, member.ProjectID, member.EntityID, member.Role, member.EntityType).Exec()
	if err != nil {
		return 0, err
	}
	pmid, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(pmid), err
}

// GetProjectByID ...
func GetProjectByID(id int64) (*models.Project, error) {
	o := GetOrmer()

	sql := `select p.project_id, p.name, u.username as owner_name, p.owner_id, p.creation_time, p.update_time  
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

// GetTotalOfProjects returns the total count of projects
// according to the query conditions
func GetTotalOfProjects(query *models.ProjectQueryParam) (int64, error) {
	var pagination *models.Pagination
	if query != nil {
		pagination = query.Pagination
		query.Pagination = nil
	}
	sql, params := projectQueryConditions(query)
	if query != nil {
		query.Pagination = pagination
	}

	sql = `select count(*) ` + sql

	var total int64
	err := GetOrmer().Raw(sql, params).QueryRow(&total)
	return total, err
}

// GetProjects returns a project list according to the query conditions
func GetProjects(query *models.ProjectQueryParam) ([]*models.Project, error) {
	sql, params := projectQueryConditions(query)

	sql = `select distinct p.project_id, p.name, p.owner_id, 
				p.creation_time, p.update_time ` + sql

	var projects []*models.Project
	_, err := GetOrmer().Raw(sql, params).QueryRows(&projects)
	return projects, err
}

func projectQueryConditions(query *models.ProjectQueryParam) (string, []interface{}) {
	params := []interface{}{}

	sql := ` from project as p`

	if query == nil {
		sql += ` where p.deleted=0 order by p.name`
		return sql, params
	}

	// if query.ProjectIDs is not nil but has no element, the query will returns no rows
	if query.ProjectIDs != nil && len(query.ProjectIDs) == 0 {
		sql += ` where 1 = 0`
		return sql, params
	}

	if len(query.Owner) != 0 {
		sql += ` join user u1
					on p.owner_id = u1.user_id`
	}

	if query.Member != nil && len(query.Member.Name) != 0 {
		sql += ` join project_member pm
					on p.project_id = pm.project_id
					join user u2
					on pm.entity_id=u2.user_id`
	}
	sql += ` where p.deleted=0`

	if len(query.Owner) != 0 {
		sql += ` and u1.username=?`
		params = append(params, query.Owner)
	}

	if len(query.Name) != 0 {
		sql += ` and p.name like ?`
		params = append(params, "%"+Escape(query.Name)+"%")
	}

	if query.Member != nil && len(query.Member.Name) != 0 {
		sql += ` and u2.username=?`
		params = append(params, query.Member.Name)

		if query.Member.Role > 0 {
			sql += ` and pm.role = ?`
			roleID := 0
			switch query.Member.Role {
			case common.RoleProjectAdmin:
				roleID = 1
			case common.RoleDeveloper:
				roleID = 2
			case common.RoleGuest:
				roleID = 3

			}
			params = append(params, roleID)
		}
	}

	if len(query.ProjectIDs) > 0 {
		sql += fmt.Sprintf(` and p.project_id in ( %s )`,
			paramPlaceholder(len(query.ProjectIDs)))
		params = append(params, query.ProjectIDs)
	}

	sql += ` order by p.name`

	if query.Pagination != nil && query.Pagination.Size > 0 {
		sql += ` limit ?`
		params = append(params, query.Pagination.Size)

		if query.Pagination.Page > 0 {
			sql += ` offset ?`
			params = append(params, (query.Pagination.Page-1)*query.Pagination.Size)
		}
	}

	return sql, params
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
