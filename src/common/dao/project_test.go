// Copyright Project Harbor Authors
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
	"fmt"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
)

func TestDeleteProject(t *testing.T) {
	name := "project_for_test"
	project := models.Project{
		OwnerID: currentUser.UserID,
		Name:    name,
	}

	id, err := AddProject(project)
	if err != nil {
		t.Fatalf("failed to add project: %v", err)
	}
	defer func() {
		if err := delProjPermanent(id); err != nil {
			t.Errorf("failed to clear up project %d: %v", id, err)
		}
	}()

	if err = DeleteProject(id); err != nil {
		t.Fatalf("failed to delete project: %v", err)
	}

	p := &models.Project{}
	if err = GetOrmer().Raw(`select * from project where project_id = ?`, id).
		QueryRow(p); err != nil {
		t.Fatalf("failed to get project: %v", err)
	}

	if !p.Deleted {
		t.Errorf("unexpeced deleted column: %t != %t", p.Deleted, true)
	}

	deletedName := fmt.Sprintf("%s#%d", name, id)
	if p.Name != deletedName {
		t.Errorf("unexpected name: %s != %s", p.Name, deletedName)
	}

}

func delProjPermanent(id int64) error {
	_, err := GetOrmer().QueryTable("access_log").
		Filter("ProjectID", id).
		Delete()
	if err != nil {
		return err
	}

	_, err = GetOrmer().Raw(`delete from project_member 
		where project_id = ?`, id).Exec()
	if err != nil {
		return err
	}

	_, err = GetOrmer().QueryTable("project").
		Filter("ProjectID", id).
		Delete()
	return err
}

func Test_projectQueryConditions(t *testing.T) {
	type args struct {
		query *models.ProjectQueryParam
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []interface{}
	}{
		{"Query invalid projectID",
			args{query: &models.ProjectQueryParam{ProjectIDs: []int64{}, Owner: "admin"}},
			"from project as p where 1 = 0",
			[]interface{}{}},
		{"Query with valid projectID",
			args{query: &models.ProjectQueryParam{ProjectIDs: []int64{2, 3}, Owner: "admin"}},
			` from project as p join harbor_user u1
					on p.owner_id = u1.user_id where p.deleted=false and u1.username=? and p.project_id in ( ?,? )`,
			[]interface{}{2, 3}},
		{"Query with valid page and member",
			args{query: &models.ProjectQueryParam{ProjectIDs: []int64{2, 3}, Owner: "admin", Name: "sample", Member: &models.MemberQuery{Name: "name", Role: 1}, Pagination: &models.Pagination{Page: 1, Size: 20}}},
			` from project as p join harbor_user u1
					on p.owner_id = u1.user_id join project_member pm
					on p.project_id = pm.project_id and pm.entity_type = 'u'
					join harbor_user u2
					on pm.entity_id=u2.user_id where p.deleted=false and u1.username=? and p.name like ? and u2.username=? and pm.role = ? and p.project_id in ( ?,? )`,
			[]interface{}{1, []int64{2, 3}, 20, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := projectQueryConditions(tt.args.query)
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("projectQueryConditions() got = %v\n, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGroupProjects(t *testing.T) {
	prepareGroupTest()
	query := &models.ProjectQueryParam{Member: &models.MemberQuery{Name: "sample_group"}}
	type args struct {
		groupDNCondition string
		query            *models.ProjectQueryParam
	}
	tests := []struct {
		name     string
		args     args
		wantSize int
		wantErr  bool
	}{
		{"Verify correct sql", args{groupDNCondition: "'cn=harbor_user,dc=example,dc=com'", query: query}, 1, false},
		{"Verify missed sql", args{groupDNCondition: "'cn=another_user,dc=example,dc=com'", query: query}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGroupProjects(tt.args.groupDNCondition, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGroupProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantSize {
				t.Errorf("GetGroupProjects() = %v, want %v", got, tt.wantSize)
			}
		})
	}
}

func prepareGroupTest() {
	initSqls := []string{
		`insert into user_group (group_name, group_type, ldap_group_dn) values ('harbor_group_01', 1, 'cn=harbor_user,dc=example,dc=com')`,
		`insert into harbor_user (username, email, password, realname) values ('sample01', 'sample01@example.com', 'harbor12345', 'sample01')`,
		`insert into project (name, owner_id) values ('group_project', 1)`,
		`insert into project (name, owner_id) values ('group_project_private', 1)`,
		`insert into project_metadata (project_id, name, value) values ((select project_id from project where name = 'group_project'), 'public', 'false')`,
		`insert into project_metadata (project_id, name, value) values ((select project_id from project where name = 'group_project_private'), 'public', 'false')`,
		`insert into project_member (project_id, entity_id, entity_type, role) values ((select project_id from project where name = 'group_project'), (select id from user_group where group_name = 'harbor_group_01'),'g', 2)`,
	}

	clearSqls := []string{
		`delete from project_metadata where project_id in (select project_id from project where name in ('group_project', 'group_project_private'))`,
		`delete from project where name in ('group_project', 'group_project_private')`,
		`delete from project_member where project_id in (select project_id from project where name in ('group_project', 'group_project_private'))`,
		`delete from user_group where group_name = 'harbor_group_01'`,
		`delete from harbor_user where username = 'sample01'`,
	}
	PrepareTestData(clearSqls, initSqls)
}

func TestGetTotalGroupProjects(t *testing.T) {
	prepareGroupTest()
	query := &models.ProjectQueryParam{Member: &models.MemberQuery{Name: "sample_group"}}
	type args struct {
		groupDNCondition string
		query            *models.ProjectQueryParam
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"Verify correct sql", args{groupDNCondition: "'cn=harbor_user,dc=example,dc=com'", query: query}, 1, false},
		{"Verify missed sql", args{groupDNCondition: "'cn=another_user,dc=example,dc=com'", query: query}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTotalGroupProjects(tt.args.groupDNCondition, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTotalGroupProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTotalGroupProjects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRolesByLDAPGroup(t *testing.T) {
	prepareGroupTest()
	project, err := GetProjectByName("group_project")
	if err != nil {
		t.Errorf("Error occurred when Get project by name: %v", err)
	}
	privateProject, err := GetProjectByName("group_project_private")
	if err != nil {
		t.Errorf("Error occurred when Get project by name: %v", err)
	}
	type args struct {
		projectID        int64
		groupDNCondition string
	}
	tests := []struct {
		name     string
		args     args
		wantSize int
		wantErr  bool
	}{
		{"Check normal", args{project.ProjectID, "'cn=harbor_user,dc=example,dc=com'"}, 1, false},
		{"Check non exist", args{privateProject.ProjectID, "'cn=not_harbor_user,dc=example,dc=com'"}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRolesByLDAPGroup(tt.args.projectID, tt.args.groupDNCondition)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetRolesByLDAPGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantSize {
				t.Errorf("TestGetRolesByLDAPGroup() = %v, want %v", len(got), tt.wantSize)
			}
		})
	}
}

func TestProjetExistsByName(t *testing.T) {
	name := "project_exist_by_name_test"
	exist := ProjectExistsByName(name)
	if exist {
		t.Errorf("project %s expected to be not exist", name)
	}

	project := models.Project{
		OwnerID: currentUser.UserID,
		Name:    name,
	}
	id, err := AddProject(project)
	if err != nil {
		t.Fatalf("failed to add project: %v", err)
	}
	defer func() {
		if err := delProjPermanent(id); err != nil {
			t.Errorf("failed to clear up project %d: %v", id, err)
		}
	}()

	exist = ProjectExistsByName(name)
	if !exist {
		t.Errorf("project %s expected to be exist", name)
	}
}
