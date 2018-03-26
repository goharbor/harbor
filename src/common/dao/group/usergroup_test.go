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

package group

import (
	"fmt"
	"os"
	"testing"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

var createdUserGroupID int

func TestMain(m *testing.M) {

	//databases := []string{"mysql", "sqlite"}
	databases := []string{"mysql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)

		result := 1
		switch database {
		case "mysql":
			dao.PrepareTestForMySQL()
		case "sqlite":
			dao.PrepareTestForSQLite()
		default:
			log.Fatalf("invalid database: %s", database)
		}

		//Extract to test utils
		initSqls := []string{
			"insert into user (username, email, password, realname)  values ('member_test_01', 'member_test_01@example.com', '123456', 'member_test_01')",
			"insert into project (name, owner_id) values ('member_test_01', 1)",
			"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_01', 1, 'CN=harbor_users,OU=sample,OU=vmware,DC=harbor,DC=com')",
			"update project set owner_id = (select user_id from user where username = 'member_test_01') where name = 'member_test_01'",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select user_id from user where username = 'member_test_01'), 'u', 1)",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select id from user_group where group_name = 'test_group_01'), 'g', 1)",
		}

		clearSqls := []string{
			"delete from project where name='member_test_01'",
			"delete from user where username='member_test_01' or username='pm_sample'",
			"delete from user_group",
			"delete from project_member",
		}
		dao.PrepareTestData(clearSqls, initSqls)

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
		{"Insert an ldap user group", args{userGroup: models.UserGroup{GroupName: "sample_group", GroupType: common.LdapGroupType, LdapGroupDN: "sample_ldap_dn_string"}}, 0, false},
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
		{"Query all ldap group", args{query: models.UserGroup{GroupType: common.LdapGroupType}}, 2, false},
		{"Query ldap group with group property", args{query: models.UserGroup{GroupType: common.LdapGroupType, LdapGroupDN: "CN=harbor_users,OU=sample,OU=vmware,DC=harbor,DC=com"}}, 1, false},
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
	userGroup := models.UserGroup{GroupName: "insert_group", GroupType: common.LdapGroupType, LdapGroupDN: "ldap_dn_string"}
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
			fmt.Printf("id=%v", createdUserGroupID)
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
				GroupType:   common.LdapGroupType}},
			false},
		{"OnBoardUserGroup second time",
			args{g: &models.UserGroup{
				GroupName:   "harbor_example",
				LdapGroupDN: "cn=harbor_example,ou=groups,dc=example,dc=com",
				GroupType:   common.LdapGroupType}},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := OnBoardUserGroup(tt.args.g, "LdapGroupDN", "GroupType"); (err != nil) != tt.wantErr {
				t.Errorf("OnBoardUserGroup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
