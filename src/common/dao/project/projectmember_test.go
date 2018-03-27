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

package project

import (
	"fmt"
	"os"
	"testing"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	_ "github.com/vmware/harbor/src/ui/auth/db"
	_ "github.com/vmware/harbor/src/ui/auth/ldap"
	cfg "github.com/vmware/harbor/src/ui/config"
)

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
		cfg.Init()
		result = m.Run()

		if result != 0 {
			os.Exit(result)
		}
	}

}

func TestDeleteProjectMemberByID(t *testing.T) {
	currentProject, err := dao.GetProjectByName("member_test_01")

	if currentProject == nil || err != nil {
		fmt.Println("Failed to load project!")
	} else {
		fmt.Printf("Load project %+v", currentProject)
	}
	var addMember = models.Member{
		ProjectID:  currentProject.ProjectID,
		EntityID:   1,
		EntityType: common.UserMember,
		Role:       models.DEVELOPER,
	}

	pmid, err := AddProjectMember(addMember)

	if err != nil {
		t.Fatalf("Failed to add project member error: %v", err)
	}

	type args struct {
		pmid int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Delete created", args{pmid}, false},
		{"Delete non exist", args{-13}, false},
		{"Delete non exist", args{13}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteProjectMemberByID(tt.args.pmid); (err != nil) != tt.wantErr {
				t.Errorf("DeleteProjectMemberByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
func TestAddProjectMember(t *testing.T) {

	currentProject, err := dao.GetProjectByName("member_test_01")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	member := models.Member{
		ProjectID:  currentProject.ProjectID,
		EntityID:   1,
		EntityType: common.UserMember,
		Role:       models.PROJECTADMIN,
	}

	log.Debugf("Current project id %v", currentProject.ProjectID)
	pmid, err := AddProjectMember(member)
	if err != nil {
		t.Errorf("Error occurred in AddProjectMember: %v", err)
	}
	if pmid == 0 {
		t.Errorf("Error add project member, pmid=0")
	}

	queryMember := models.Member{
		ProjectID: currentProject.ProjectID,
		ID:        pmid,
	}

	memberList, err := GetProjectMember(queryMember)
	if err != nil {
		t.Errorf("Failed to query project member, %v, error: %v", queryMember, err)
	}

	if len(memberList) == 0 {
		t.Errorf("Failed to query project member, %v", queryMember)
	}
}
func TestUpdateProjectMemberRole(t *testing.T) {
	currentProject, err := dao.GetProjectByName("member_test_01")
	user := models.User{
		Username: "pm_sample",
		Email:    "pm_sample@example.com",
		Realname: "pm_sample",
		Password: "1234567d",
	}
	o := dao.GetOrmer()
	userID, err := o.Insert(&user)
	if err != nil {
		t.Errorf("Error occurred when add user: %v", err)
	}
	member := models.Member{
		ProjectID:  currentProject.ProjectID,
		EntityID:   int(userID),
		EntityType: common.UserMember,
		Role:       models.PROJECTADMIN,
	}

	pmid, err := AddProjectMember(member)
	if err != nil {
		t.Errorf("Error occurred in UpdateProjectMember: %v", err)
	}

	UpdateProjectMemberRole(pmid, models.DEVELOPER)

	queryMember := models.Member{
		ProjectID:  currentProject.ProjectID,
		EntityID:   int(userID),
		EntityType: common.UserMember,
	}

	memberList, err := GetProjectMember(queryMember)
	if err != nil {
		t.Errorf("Error occurred in GetProjectMember: %v", err)
	}
	if len(memberList) != 1 {
		t.Errorf("Error occurred in Failed,  size: %d, condition:%+v", len(memberList), queryMember)
	}
	memberItem := memberList[0]
	if memberItem.Role != models.DEVELOPER || memberItem.Entityname != user.Username {
		t.Errorf("member doesn't match!")
	}

}

func TestGetProjectMember(t *testing.T) {
	currentProject, err := dao.GetProjectByName("member_test_01")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	var memberList1 = []*models.Member{
		&models.Member{
			ID:         346,
			Entityname: "admin",
			Rolename:   "projectAdmin",
			Role:       1,
			EntityID:   1,
			EntityType: "u"},
	}
	var memberList2 = []*models.Member{
		&models.Member{
			ID:         398,
			Entityname: "test_group_01",
			Rolename:   "projectAdmin",
			Role:       1,
			EntityType: "g"},
	}
	type args struct {
		queryMember models.Member
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.Member
		wantErr bool
	}{
		{"Query default project member", args{models.Member{ProjectID: currentProject.ProjectID, Entityname: "admin"}}, memberList1, false},
		{"Query default project member group", args{models.Member{ProjectID: currentProject.ProjectID, Entityname: "test_group_01"}}, memberList2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetProjectMember(tt.args.queryMember)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 1 {
				t.Errorf("Error occurred when query project member")
			}
			itemGot := got[0]
			itemWant := tt.want[0]

			if itemGot.Entityname != itemWant.Entityname || itemGot.Role != itemWant.Role || itemGot.EntityType != itemWant.EntityType {
				t.Errorf("test failed, got:%+v, want:%+v", itemGot, itemWant)
			}
		})
	}

}
