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

package group

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

var createdUserGroupID int

func TestMain(m *testing.M) {

	// databases := []string{"mysql", "sqlite"}
	databases := []string{"postgresql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)

		result := 1
		switch database {
		case "postgresql":
			dao.PrepareTestForPostgresSQL()
		default:
			log.Fatalf("invalid database: %s", database)
		}

		// Extract to test utils
		initSqls := []string{
			"insert into harbor_user (username, email, password, realname)  values ('member_test_01', 'member_test_01@example.com', '123456', 'member_test_01')",
			"insert into harbor_user (username, email, password, realname)  values ('grouptestu09', 'grouptestu09@example.com', '123456', 'grouptestu09')",
			"insert into project (name, owner_id) values ('member_test_01', 1)",
			`insert into project (name, owner_id) values ('group_project2', 1)`,
			`insert into project (name, owner_id) values ('group_project_private', 1)`,
			"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_01', 1, 'cn=harbor_users,ou=sample,ou=vmware,dc=harbor,dc=com')",
			"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_http_group', 2, '')",
			"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_myhttp_group', 2, '')",
			"update project set owner_id = (select user_id from harbor_user where username = 'member_test_01') where name = 'member_test_01'",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select user_id from harbor_user where username = 'member_test_01'), 'u', 1)",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select id from user_group where group_name = 'test_group_01'), 'g', 1)",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select id from user_group where group_name = 'test_http_group'), 'g', 4)",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select id from user_group where group_name = 'test_myhttp_group'), 'g', 4)",
		}

		clearSqls := []string{
			"delete from project where name='member_test_01'",
			"delete from project where name='group_project2'",
			"delete from project where name='group_project_private'",
			"delete from harbor_user where username='member_test_01' or username='pm_sample' or username='grouptestu09'",
			"delete from user_group",
			"delete from project_member",
		}
		dao.ExecuteBatchSQL(initSqls)
		defer dao.ExecuteBatchSQL(clearSqls)

		result = m.Run()

		if result != 0 {
			os.Exit(result)
		}
	}

}

func TestAddUserGroup(t *testing.T) {
	type args struct {
		userGroup models.UserGroup
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"Insert an ldap user group", args{userGroup: models.UserGroup{GroupName: "sample_group", GroupType: common.LDAPGroupType, LdapGroupDN: "sample_ldap_dn_string"}}, 0, false},
		{"Insert other user group", args{userGroup: models.UserGroup{GroupName: "other_group", GroupType: 3, LdapGroupDN: "other information"}}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddUserGroup(tt.args.userGroup)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got <= 0 {
				t.Errorf("Failed to add user group")
			}
		})
	}
}

func TestQueryUserGroup(t *testing.T) {
	type args struct {
		query models.UserGroup
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"Query all user group", args{query: models.UserGroup{GroupName: "test_group_01"}}, 1, false},
		{"Query all ldap group", args{query: models.UserGroup{GroupType: common.LDAPGroupType}}, 2, false},
		{"Query ldap group with group property", args{query: models.UserGroup{GroupType: common.LDAPGroupType, LdapGroupDN: "CN=harbor_users,OU=sample,OU=vmware,DC=harbor,DC=com"}}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QueryUserGroup(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("QueryUserGroup() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestGetUserGroup(t *testing.T) {
	userGroup := models.UserGroup{GroupName: "insert_group", GroupType: common.LDAPGroupType, LdapGroupDN: "ldap_dn_string"}
	result, err := AddUserGroup(userGroup)
	if err != nil {
		t.Errorf("Error occurred when AddUserGroup: %v", err)
	}
	createdUserGroupID = result
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Get User Group", args{id: result}, "insert_group", false},
		{"Get User Group does not exist", args{id: 9999}, "insert_group", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserGroup(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && got.GroupName != tt.want {
				t.Errorf("GetUserGroup() = %v, want %v", got.GroupName, tt.want)
			}
		})
	}
}
func TestUpdateUserGroup(t *testing.T) {
	if createdUserGroupID == 0 {
		fmt.Println("User group doesn't created, skip to test!")
		return
	}
	type args struct {
		id        int
		groupName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Update user group", args{id: createdUserGroupID, groupName: "updated_groupname"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("id=%v\n", createdUserGroupID)
			if err := UpdateUserGroupName(tt.args.id, tt.args.groupName); (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				userGroup, err := GetUserGroup(tt.args.id)
				if err != nil {
					t.Errorf("Error occurred when GetUserGroup: %v", err)
				}
				if userGroup == nil {
					t.Fatalf("Failed to get updated user group")
				}
				if userGroup.GroupName != tt.args.groupName {
					t.Fatalf("Failed to update user group")
				}
			}
		})
	}
}

func TestDeleteUserGroup(t *testing.T) {
	if createdUserGroupID == 0 {
		fmt.Println("User group doesn't created, skip to test!")
		return
	}

	type args struct {
		id int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Delete existing user group", args{id: createdUserGroupID}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteUserGroup(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteUserGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOnBoardUserGroup(t *testing.T) {
	type args struct {
		g *models.UserGroup
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"OnBoardUserGroup",
			args{g: &models.UserGroup{
				GroupName:   "harbor_example",
				LdapGroupDN: "cn=harbor_example,ou=groups,dc=example,dc=com",
				GroupType:   common.LDAPGroupType}},
			false},
		{"OnBoardUserGroup second time",
			args{g: &models.UserGroup{
				GroupName:   "harbor_example",
				LdapGroupDN: "cn=harbor_example,ou=groups,dc=example,dc=com",
				GroupType:   common.LDAPGroupType}},
			false},
		{"OnBoardUserGroup HTTP user group",
			args{g: &models.UserGroup{
				GroupName: "test_myhttp_group",
				GroupType: common.HTTPGroupType}},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := OnBoardUserGroup(tt.args.g); (err != nil) != tt.wantErr {
				t.Errorf("OnBoardUserGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetGroupProjects(t *testing.T) {
	projectID1, err := dao.AddProject(models.Project{
		Name:    "grouptest01",
		OwnerID: 1,
	})
	if err != nil {
		t.Errorf("Error occurred when AddProject: %v", err)
	}
	defer dao.DeleteProject(projectID1)
	projectID2, err := dao.AddProject(models.Project{
		Name:    "grouptest02",
		OwnerID: 1,
	})
	if err != nil {
		t.Errorf("Error occurred when AddProject: %v", err)
	}
	defer dao.DeleteProject(projectID2)
	groupID, err := AddUserGroup(models.UserGroup{
		GroupName:   "test_group_03",
		GroupType:   1,
		LdapGroupDN: "cn=harbor_users,ou=groups,dc=example,dc=com",
	})
	if err != nil {
		t.Errorf("Error occurred when AddUserGroup: %v", err)
	}
	defer DeleteUserGroup(groupID)
	pmid, err := project.AddProjectMember(models.Member{
		ProjectID:  projectID1,
		EntityID:   groupID,
		EntityType: "g",
	})
	defer project.DeleteProjectMemberByID(pmid)
	type args struct {
		query *models.ProjectQueryParam
	}
	member := &models.MemberQuery{
		Name: "grouptestu09",
	}
	tests := []struct {
		name     string
		args     args
		wantSize int
		wantErr  bool
	}{
		{"Query with group DN",
			args{&models.ProjectQueryParam{
				Member: member,
			}},
			1, false},
		{"Query without group DN",
			args{&models.ProjectQueryParam{}},
			1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.GetGroupProjects([]int{groupID}, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGroupProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < tt.wantSize {
				t.Errorf("GetGroupProjects() size: %v, want %v", len(got), tt.wantSize)
			}
		})
	}
}

func TestGetTotalGroupProjects(t *testing.T) {
	projectID1, err := dao.AddProject(models.Project{
		Name:    "grouptest01",
		OwnerID: 1,
	})
	if err != nil {
		t.Errorf("Error occurred when AddProject: %v", err)
	}
	defer dao.DeleteProject(projectID1)
	projectID2, err := dao.AddProject(models.Project{
		Name:    "grouptest02",
		OwnerID: 1,
	})
	if err != nil {
		t.Errorf("Error occurred when AddProject: %v", err)
	}
	defer dao.DeleteProject(projectID2)
	groupID, err := AddUserGroup(models.UserGroup{
		GroupName:   "test_group_05",
		GroupType:   1,
		LdapGroupDN: "cn=harbor_users,ou=groups,dc=example,dc=com",
	})
	if err != nil {
		t.Errorf("Error occurred when AddUserGroup: %v", err)
	}
	defer DeleteUserGroup(groupID)
	pmid, err := project.AddProjectMember(models.Member{
		ProjectID:  projectID1,
		EntityID:   groupID,
		EntityType: "g",
	})
	defer project.DeleteProjectMemberByID(pmid)
	type args struct {
		query *models.ProjectQueryParam
	}
	tests := []struct {
		name     string
		args     args
		wantSize int
		wantErr  bool
	}{
		{"Query with group ID",
			args{&models.ProjectQueryParam{}},
			1, false},
		{"Query without group ID",
			args{&models.ProjectQueryParam{}},
			1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.GetTotalGroupProjects([]int{groupID}, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGroupProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got < tt.wantSize {
				t.Errorf("GetGroupProjects() size: %v, want %v", got, tt.wantSize)
			}
		})
	}
}
func TestGetRolesByLDAPGroup(t *testing.T) {

	userGroupList, err := QueryUserGroup(models.UserGroup{LdapGroupDN: "cn=harbor_users,ou=sample,ou=vmware,dc=harbor,dc=com", GroupType: 1})
	if err != nil || len(userGroupList) < 1 {
		t.Errorf("failed to query user group, err %v", err)
	}
	gl2, err2 := GetGroupIDByGroupName([]string{"test_http_group", "test_myhttp_group"}, common.HTTPGroupType)
	if err2 != nil || len(gl2) != 2 {
		t.Errorf("failed to query http user group, err %v", err)
	}
	project, err := dao.GetProjectByName("member_test_01")
	if err != nil {
		t.Errorf("Error occurred when Get project by name: %v", err)
	}
	privateProject, err := dao.GetProjectByName("group_project_private")
	if err != nil {
		t.Errorf("Error occurred when Get project by name: %v", err)
	}

	type args struct {
		projectID int64
		groupIDs  []int
	}
	tests := []struct {
		name     string
		args     args
		wantSize int
		wantErr  bool
	}{
		{"Check normal", args{projectID: project.ProjectID, groupIDs: []int{userGroupList[0].ID, gl2[0], gl2[1]}}, 2, false},
		{"Check non exist", args{projectID: privateProject.ProjectID, groupIDs: []int{9999}}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.GetRolesByGroupID(tt.args.projectID, tt.args.groupIDs)
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

func TestGetGroupIDByGroupName(t *testing.T) {
	groupList, err := QueryUserGroup(models.UserGroup{GroupName: "test_http_group", GroupType: 2})
	if err != nil {
		t.Error(err)
	}
	if len(groupList) < 0 {
		t.Error(err)
	}
	groupList2, err := QueryUserGroup(models.UserGroup{GroupName: "test_myhttp_group", GroupType: 2})
	if err != nil {
		t.Error(err)
	}
	if len(groupList2) < 0 {
		t.Error(err)
	}
	var expectGroupID []int
	type args struct {
		groupName []string
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr bool
	}{
		{"empty query", args{groupName: []string{}}, expectGroupID, false},
		{"normal query", args{groupName: []string{"test_http_group", "test_myhttp_group"}}, []int{groupList[0].ID, groupList2[0].ID}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGroupIDByGroupName(tt.args.groupName, common.HTTPGroupType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHTTPGroupIDByGroupName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHTTPGroupIDByGroupName() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
