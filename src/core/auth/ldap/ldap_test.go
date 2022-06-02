// Copyright 2018 Project Harbor Authors
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
package ldap

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/member"
	memberModels "github.com/goharbor/harbor/src/pkg/member/models"
	userpkg "github.com/goharbor/harbor/src/pkg/user"
	userDao "github.com/goharbor/harbor/src/pkg/user/dao"
	"github.com/goharbor/harbor/src/pkg/usergroup"
	ugModel "github.com/goharbor/harbor/src/pkg/usergroup/model"
	"github.com/stretchr/testify/assert"
)

var ldapTestConfig = map[string]interface{}{
	common.ExtEndpoint:        "host01.com",
	common.AUTHMode:           "ldap_auth",
	common.DatabaseType:       "postgresql",
	common.PostGreSQLHOST:     "127.0.0.1",
	common.PostGreSQLPort:     5432,
	common.PostGreSQLUsername: "postgres",
	common.PostGreSQLPassword: "root123",
	common.PostGreSQLDatabase: "registry",
	// config.SelfRegistration: true,
	common.LDAPURL:                "ldap://127.0.0.1",
	common.LDAPSearchDN:           "cn=admin,dc=example,dc=com",
	common.LDAPSearchPwd:          "admin",
	common.LDAPBaseDN:             "dc=example,dc=com",
	common.LDAPUID:                "uid",
	common.LDAPFilter:             "",
	common.LDAPScope:              2,
	common.LDAPTimeout:            30,
	common.AdminInitialPassword:   "password",
	common.LDAPGroupSearchFilter:  "objectclass=groupOfNames",
	common.LDAPGroupBaseDN:        "dc=example,dc=com",
	common.LDAPGroupAttributeName: "cn",
	common.LDAPGroupSearchScope:   2,
	common.LDAPGroupAdminDn:       "cn=harbor_users,ou=groups,dc=example,dc=com",
}

var authHelper *Auth

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	config.InitWithSettings(ldapTestConfig)

	secretKeyPath := "/tmp/secretkey"
	_, err := test.GenerateKey(secretKeyPath)
	if err != nil {
		log.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		log.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}

	authHelper = &Auth{
		userMgr: userpkg.New(),
	}

	// Extract to test utils
	initSqls := []string{
		"insert into harbor_user (username, email, password, realname)  values ('member_test_01', 'member_test_01@example.com', '123456', 'member_test_01')",
		"insert into project (name, owner_id) values ('member_test_01', 1)",
		"insert into project (name, owner_id) values ('member_test_02', 1)",
		"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_01', 1, 'CN=harbor_users,OU=sample,OU=vmware,DC=harbor,DC=com')",
		"update project set owner_id = (select user_id from harbor_user where username = 'member_test_01') where name = 'member_test_01'",
		"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select user_id from harbor_user where username = 'member_test_01'), 'u', 1)",
		"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select id from user_group where group_name = 'test_group_01'), 'g', 1)",
	}

	clearSqls := []string{
		"delete from project where name='member_test_01'",
		"delete from project where name='member_test_02'",
		"delete from harbor_user where username='member_test_01' or username='pm_sample'",
		"delete from user_group",
		"delete from project_member where id > 1",
	}
	dao.ExecuteBatchSQL(initSqls)
	defer dao.ExecuteBatchSQL(clearSqls)
	retCode := m.Run()
	os.Exit(retCode)
}

func TestAuthenticate(t *testing.T) {
	ctx := orm.Context()

	var person models.AuthModel
	person.Principal = "test"
	person.Password = "123456"
	user, err := authHelper.Authenticate(ctx, person)
	if err != nil {
		t.Errorf("unexpected ldap authenticate fail: %v", err)
	}
	if user.Username != "test" {
		t.Errorf("unexpected ldap user authenticate fail: %s = %s", "user.Username", user.Username)
	}
	person.Principal = "test"
	person.Password = "1"
	user, err = authHelper.Authenticate(ctx, person)

	if _, ok := err.(auth.ErrAuth); !ok {
		t.Errorf("Expected an ErrAuth on wrong password, but got: %v", err)
	}
	person.Principal = ""
	person.Password = ""
	user, err = authHelper.Authenticate(ctx, person)
	if _, ok := err.(auth.ErrAuth); !ok {
		t.Errorf("Expected an ErrAuth on empty credentials, but got: %v", err)
	}
	// authenticate the second time
	person2 := models.AuthModel{
		Principal: "test",
		Password:  "123456",
	}
	user2, err := authHelper.Authenticate(ctx, person2)

	if err != nil {
		t.Errorf("unexpected ldap error: %v", err)
	}

	if user2 == nil {
		t.Errorf("Can not login user with person2 %+v", person2)
	}
}

func TestSearchUser(t *testing.T) {
	ctx := orm.Context()

	var username = "test"
	user, err := authHelper.SearchUser(ctx, username)
	if err != nil {
		t.Errorf("Search user failed %v", err)
	}
	if user == nil {
		t.Errorf("Search user failed %v", user)
	}
}
func TestAuthenticateWithAdmin(t *testing.T) {
	ctx := orm.Context()

	var person models.AuthModel
	person.Principal = "mike"
	person.Password = "zhu88jie"
	user, err := authHelper.Authenticate(ctx, person)
	if err != nil {
		t.Errorf("unexpected ldap authenticate fail: %v", err)
	}
	if user.Username != "mike" {
		t.Errorf("unexpected ldap user authenticate fail: %s = %s", "user.Username", user.Username)
	}
	if !user.AdminRoleInAuth {
		t.Errorf("ldap user mike should have admin role!")
	}
}
func TestAuthenticateWithoutAdmin(t *testing.T) {
	ctx := orm.Context()

	var person models.AuthModel
	person.Principal = "user001"
	person.Password = "Test1@34"
	user, err := authHelper.Authenticate(ctx, person)
	if err != nil {
		t.Errorf("unexpected ldap authenticate fail: %v", err)
	}
	if user.Username != "user001" {
		t.Errorf("unexpected ldap user authenticate fail: %s = %s", "user.Username", user.Username)
	}
	if user.AdminRoleInAuth {
		t.Errorf("ldap user user001 should not have admin role!")
	}
}
func TestSearchUser_02(t *testing.T) {
	ctx := orm.Context()

	var username = "nonexist"
	user, _ := authHelper.SearchUser(ctx, username)
	if user != nil {
		t.Errorf("Should failed to search nonexist user")
	}

}

func TestOnBoardUser(t *testing.T) {
	ctx := orm.Context()

	user := &models.User{
		Username: "sample",
		Email:    "sample@example.com",
		Realname: "Sample",
	}
	err := authHelper.OnBoardUser(ctx, user)
	if err != nil {
		t.Errorf("Failed to onboard user")
	}
	if user.UserID <= 0 {
		t.Errorf("Failed to onboard user")
	}
	assert.Equal(t, "sample@example.com", user.Email)
}

func TestOnBoardUser_02(t *testing.T) {
	ctx := orm.Context()

	user := &models.User{
		Username: "sample02",
		Realname: "Sample02",
	}
	err := authHelper.OnBoardUser(ctx, user)
	if err != nil {
		t.Errorf("Failed to onboard user")
	}
	if user.UserID <= 0 {
		t.Errorf("Failed to onboard user")
	}

	assert.Equal(t, "", user.Email)
	userDao.New().Delete(ctx, user.UserID)
}

func TestOnBoardUser_03(t *testing.T) {
	ctx := orm.Context()

	user := &models.User{
		Username: "sample03@example.com",
		Realname: "Sample03",
	}
	err := authHelper.OnBoardUser(ctx, user)
	if err != nil {
		t.Errorf("Failed to onboard user")
	}
	if user.UserID <= 0 {
		t.Errorf("Failed to onboard user")
	}

	assert.Equal(t, "sample03@example.com", user.Email)
	userDao.New().Delete(ctx, user.UserID)
}

func TestAuthenticateHelperOnBoardUser(t *testing.T) {
	ctx := orm.Context()

	user := models.User{
		Username: "test01",
		Realname: "test01",
		Email:    "test01@example.com",
	}

	err := auth.OnBoardUser(ctx, &user)
	if err != nil {
		t.Errorf("Failed to onboard user error: %v", err)
	}

	if user.UserID <= 0 {
		t.Errorf("Failed to onboard user, userid: %v", user.UserID)
	}

}

func TestOnBoardGroup(t *testing.T) {
	ctx := orm.Context()

	group := ugModel.UserGroup{
		GroupName:   "harbor_group2",
		LdapGroupDN: "cn=harbor_group2,ou=groups,dc=example,dc=com",
	}
	newGroupName := "group_name123"
	err := auth.OnBoardGroup(ctx, &group, newGroupName)
	if err != nil {
		t.Errorf("Failed to OnBoardGroup, %+v", group)
	}
	if group.GroupName != "group_name123" {
		t.Errorf("The OnBoardGroup should have name %v", newGroupName)
	}
}

func TestAuthenticateHelperSearchUser(t *testing.T) {
	ctx := orm.Context()

	user, err := auth.SearchUser(ctx, "test")
	if err != nil {
		t.Error("Failed to search user, test")
	}

	if user == nil {
		t.Error("Failed to search user test")
	}
}

func TestPostAuthentication(t *testing.T) {
	ctx := orm.Context()

	assert := assert.New(t)
	user1 := &models.User{
		Username: "test003",
		Email:    "test003@vmware.com",
		Realname: "test003",
	}

	queryUsername := "test003"

	err := auth.OnBoardUser(ctx, user1)
	assert.Nil(err)

	user2 := &models.User{
		Username: "test003",
		Email:    "234invalidmail@@@@@",
	}

	auth.PostAuthenticate(ctx, user2)

	dbUser, err := userpkg.Mgr.GetByName(ctx, queryUsername)
	if err != nil {
		t.Fatalf("Failed to get user, error %v", err)
	}
	assert.EqualValues("test003@vmware.com", dbUser.Email)

	user3 := &models.User{
		Username: "test003",
	}

	auth.PostAuthenticate(ctx, user3)
	dbUser, err = userpkg.Mgr.GetByName(ctx, queryUsername)
	if err != nil {
		t.Fatalf("Failed to get user, error %v", err)
	}
	assert.EqualValues("test003@vmware.com", dbUser.Email)

	user4 := &models.User{
		Username: "test003",
		Email:    "test003@example.com",
	}

	auth.PostAuthenticate(ctx, user4)
	dbUser, err = userpkg.Mgr.GetByName(ctx, queryUsername)
	if err != nil {
		t.Fatalf("Failed to get user, error %v", err)
	}
	assert.EqualValues("test003@example.com", dbUser.Email)
	userDao.New().Delete(ctx, dbUser.UserID)
}

func TestSearchAndOnBoardUser(t *testing.T) {
	ctx := orm.Context()

	userID, err := auth.SearchAndOnBoardUser(ctx, "mike02")
	defer userDao.New().Delete(ctx, userID)
	if err != nil {
		t.Errorf("Error occurred when SearchAndOnBoardUser: %v", err)
	}
	if userID == 0 {
		t.Errorf("Can not search and onboard user %v", "mike")
	}
}
func TestAddProjectMemberWithLdapUser(t *testing.T) {
	memberMgr := member.Mgr
	ctx := orm.Context()
	currentProject, err := pkg.ProjectMgr.Get(ctx, "member_test_01")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	userID, err := auth.SearchAndOnBoardUser(ctx, "mike")
	member := memberModels.Member{
		ProjectID:  currentProject.ProjectID,
		EntityType: common.UserMember,
		Entityname: "mike",
		EntityID:   userID,
		Role:       common.RoleProjectAdmin,
	}
	pmid, err := memberMgr.AddProjectMember(ctx, member)
	if err != nil {
		t.Errorf("Error occurred in AddOrUpdateProjectMember: %v", err)
	}
	if pmid == 0 {
		t.Errorf("Error occurred in AddOrUpdateProjectMember: pmid:%v", pmid)
	}

	currentProject, err = pkg.ProjectMgr.Get(ctx, "member_test_02")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	member2 := memberModels.Member{
		ProjectID:  currentProject.ProjectID,
		EntityType: common.UserMember,
		Entityname: "mike",
		EntityID:   userID,
		Role:       common.RoleProjectAdmin,
	}
	pmid, err = memberMgr.AddProjectMember(ctx, member2)
	if err != nil {
		t.Errorf("Error occurred in AddOrUpdateProjectMember: %v", err)
	}
	if pmid == 0 {
		t.Errorf("Error occurred in AddOrUpdateProjectMember: pmid:%v", pmid)
	}
}
func TestAddProjectMemberWithLdapGroup(t *testing.T) {
	memberMgr := member.Mgr
	ctx := orm.Context()
	currentProject, err := pkg.ProjectMgr.Get(ctx, "member_test_01")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	userGroups := []ugModel.UserGroup{{GroupName: "cn=harbor_users,ou=groups,dc=example,dc=com", LdapGroupDN: "cn=harbor_users,ou=groups,dc=example,dc=com", GroupType: common.LDAPGroupType}}
	groupIds, err := usergroup.Mgr.Populate(ctx, userGroups)
	m := memberModels.Member{
		ProjectID:  currentProject.ProjectID,
		EntityType: common.GroupMember,
		EntityID:   groupIds[0],
		Role:       common.RoleProjectAdmin,
	}
	pmid, err := memberMgr.AddProjectMember(ctx, m)
	if err != nil {
		t.Errorf("Error occurred in AddOrUpdateProjectMember: %v", err)
	}
	if pmid == 0 {
		t.Errorf("Error occurred in AddOrUpdateProjectMember: pmid: %v", pmid)
	}
	queryMember := memberModels.Member{
		ProjectID: currentProject.ProjectID,
	}
	memberList, err := member.Mgr.List(ctx, queryMember, nil)
	if err != nil {
		t.Errorf("Failed to query project member, %v, error: %v", queryMember, err)
	}
	if len(memberList) == 0 {
		t.Errorf("Failed to query project member, %v", queryMember)
	}
}
