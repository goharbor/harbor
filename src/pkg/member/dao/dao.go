//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package dao

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/lib/q"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/member/models"
)

func init() {
	orm.RegisterModel(
		new(models.Member),
	)
}

// DAO the dao for project member
type DAO interface {
	// GetProjectMember gets all members of the project.
	GetProjectMember(ctx context.Context, queryMember models.Member, query *q.Query) ([]*models.Member, error)
	// GetTotalOfProjectMembers returns total of project members
	GetTotalOfProjectMembers(ctx context.Context, projectID int64, query *q.Query, roles ...int) (int, error)
	// AddProjectMember inserts a record to table project_member
	AddProjectMember(ctx context.Context, member models.Member) (int, error)
	// UpdateProjectMemberRole updates the record in table project_member, only role can be changed
	UpdateProjectMemberRole(ctx context.Context, projectID int64, pmID int, role int) error
	// DeleteProjectMemberByID - Delete Project Member by ID
	DeleteProjectMemberByID(ctx context.Context, projectID int64, pmid int) error
	// DeleteProjectMemberByUserID -- Delete project member by user id
	DeleteProjectMemberByUserID(ctx context.Context, uid int) error
	// SearchMemberByName search members of the project by entity_name
	SearchMemberByName(ctx context.Context, projectID int64, entityName string) ([]*models.Member, error)
	// ListRoles lists the roles of user for the specific project
	ListRoles(ctx context.Context, user *models.User, projectID int64) ([]int, error)
}

type dao struct {
}

// New ...
func New() DAO {
	return &dao{}
}

func (d *dao) GetProjectMember(ctx context.Context, queryMember models.Member, query *q.Query) ([]*models.Member, error) {
	log.Debugf("Query condition %+v", queryMember)
	if queryMember.ProjectID == 0 {
		return nil, fmt.Errorf("failed to query project member, query condition %v", queryMember)
	}
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	sql := ` select a.* from (select pm.id as id, pm.project_id as project_id, ug.id as entity_id, ug.group_name as entity_name, ug.creation_time, ug.update_time, r.name as rolename,
		r.role_id as role, pm.entity_type as entity_type from user_group ug join project_member pm
		on pm.project_id = ? and ug.id = pm.entity_id join role r on pm.role = r.role_id where  pm.entity_type = 'g'
		union
		select pm.id as id, pm.project_id as project_id, u.user_id as entity_id, u.username as entity_name, u.creation_time, u.update_time, r.name as rolename,
		r.role_id as role, pm.entity_type as entity_type from harbor_user u join project_member pm
		on pm.project_id = ? and u.user_id = pm.entity_id
		join role r on pm.role = r.role_id where pm.entity_type = 'u') as a where a.project_id = ? `

	queryParam := make([]interface{}, 1)
	// used ProjectID already
	queryParam = append(queryParam, queryMember.ProjectID)
	queryParam = append(queryParam, queryMember.ProjectID)
	queryParam = append(queryParam, queryMember.ProjectID)

	if len(queryMember.Entityname) > 0 {
		sql += " and a.entity_name = ? "
		queryParam = append(queryParam, queryMember.Entityname)
	}

	if len(queryMember.EntityType) == 1 {
		sql += " and a.entity_type = ? "
		queryParam = append(queryParam, queryMember.EntityType)
	}

	if queryMember.EntityID > 0 {
		sql += " and a.entity_id = ? "
		queryParam = append(queryParam, queryMember.EntityID)
	}
	if queryMember.ID > 0 {
		sql += " and a.id = ? "
		queryParam = append(queryParam, queryMember.ID)
	}
	sql += ` order by entity_name `
	sql, queryParam = orm.PaginationOnRawSQL(query, sql, queryParam)
	members := []*models.Member{}
	_, err = o.Raw(sql, queryParam).QueryRows(&members)

	return members, err
}

func (d *dao) GetTotalOfProjectMembers(ctx context.Context, projectID int64, query *q.Query, roles ...int) (int, error) {
	log.Debugf("Query condition %+v", projectID)
	if projectID == 0 {
		return 0, fmt.Errorf("failed to get total of project members, project id required %v", projectID)
	}

	sql := "SELECT COUNT(1) FROM project_member WHERE project_id = ?"

	queryParam := []interface{}{projectID}

	if len(roles) > 0 {
		sql += " AND role = ?"
		queryParam = append(queryParam, roles[0])
	}

	var count int
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	o.Raw(sql, queryParam).QueryRow(&count)
	return count, err
}

func (d *dao) AddProjectMember(ctx context.Context, member models.Member) (int, error) {
	log.Debugf("Adding project member %+v", member)
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	if member.EntityID <= 0 {
		return 0, fmt.Errorf("invalid entity_id, member: %+v", member)
	}

	if member.ProjectID <= 0 {
		return 0, fmt.Errorf("invalid project_id, member: %+v", member)
	}

	delSQL := "delete from project_member where project_id = ? and entity_id = ? and entity_type = ? "
	_, err = o.Raw(delSQL, member.ProjectID, member.EntityID, member.EntityType).Exec()
	if err != nil {
		return 0, err
	}

	var pmid int
	sql := "insert into project_member (project_id, entity_id , role, entity_type) values (?, ?, ?, ?) RETURNING id"
	err = o.Raw(sql, member.ProjectID, member.EntityID, member.Role, member.EntityType).QueryRow(&pmid)
	if err != nil {
		return 0, err
	}
	return pmid, err
}

func (d *dao) UpdateProjectMemberRole(ctx context.Context, projectID int64, pmID int, role int) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	sql := "update project_member set role = ? where project_id = ? and id = ?  "
	_, err = o.Raw(sql, role, projectID, pmID).Exec()
	return err
}

func (d *dao) DeleteProjectMemberByID(ctx context.Context, projectID int64, pmid int) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	sql := "delete from project_member where project_id = ? and id = ?"
	if _, err := o.Raw(sql, projectID, pmid).Exec(); err != nil {
		return err
	}
	return nil
}

func (d *dao) DeleteProjectMemberByUserID(ctx context.Context, uid int) error {
	sql := "delete from project_member where entity_type = 'u' and entity_id = ? "
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = o.Raw(sql, uid).Exec()
	return err
}

func (d *dao) SearchMemberByName(ctx context.Context, projectID int64, entityName string) ([]*models.Member, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	sql := `select pm.id, pm.project_id,
	               u.username as entity_name,
	               r.name as rolename,
			       pm.role, pm.entity_id, pm.entity_type
			  from project_member pm
         left join harbor_user u on pm.entity_id = u.user_id and pm.entity_type = 'u'
		 left join role r on pm.role = r.role_id
			 where pm.project_id = ? and u.username like ?
			union
		   select pm.id, pm.project_id,
			       ug.group_name as entity_name,
				   r.name as rolename,
				   pm.role, pm.entity_id, pm.entity_type
		      from project_member pm
	     left join user_group ug on pm.entity_id = ug.id and pm.entity_type = 'g'
	     left join role r on pm.role = r.role_id
			 where pm.project_id = ? and ug.group_name like ?
			 order by entity_name  `
	queryParam := make([]interface{}, 4)
	queryParam = append(queryParam, projectID)
	queryParam = append(queryParam, "%"+orm.Escape(entityName)+"%")
	queryParam = append(queryParam, projectID)
	queryParam = append(queryParam, "%"+orm.Escape(entityName)+"%")
	members := []*models.Member{}
	log.Debugf("Query sql: %v", sql)
	_, err = o.Raw(sql, queryParam).QueryRows(&members)
	return members, err
}

func (d *dao) ListRoles(ctx context.Context, user *models.User, projectID int64) ([]int, error) {
	if user == nil {
		return nil, nil
	}
	var params []interface{}
	sql := `
		select role
			from project_member
			where entity_type = 'u' and entity_id = ? and project_id = ? `
	params = append(params, user.UserID, projectID)
	if len(user.GroupIDs) > 0 {
		sql += fmt.Sprintf(`union
			select role
			from project_member
			where entity_type = 'g' and entity_id in ( %s ) and project_id = ? `, orm.ParamPlaceholderForIn(len(user.GroupIDs)))
		params = append(params, user.GroupIDs)
		params = append(params, projectID)
	}
	roles := []int{}
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	_, err = o.Raw(sql, params).QueryRows(&roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
