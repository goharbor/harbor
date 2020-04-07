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
	"os"
	"testing"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/assert"
)

func execUpdate(o orm.Ormer, sql string, params ...interface{}) error {
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(params...)
	if err != nil {
		return err
	}
	return nil
}

func cleanByUser(username string) {
	var err error

	o := GetOrmer()
	o.Begin()

	err = execUpdate(o, `delete
		from project_member
		where entity_id = (
			select user_id
			from harbor_user
			where username = ?
		) `, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete
		from project_member
		where project_id = (
			select project_id
			from project
			where name = ?
		)`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from project where name = ?`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from harbor_user where username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}
	err = execUpdate(o, `delete from replication_policy where id < 99`)
	if err != nil {
		log.Error(err)
	}
	err = execUpdate(o, `delete from registry where id < 99`)
	if err != nil {
		log.Error(err)
	}
	o.Commit()
}

const username string = "Tester01"
const password string = "Abc12345"
const projectName string = "test_project"
const repositoryName string = "test_repository"
const repoTag string = "test1.1"
const repoTag2 string = "test1.2"
const SysAdmin int = 1
const projectAdmin int = 2
const developer int = 3
const guest int = 4

const publicityOn = 1
const publicityOff = 0

func TestMain(m *testing.M) {
	databases := []string{"postgresql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)
		result := 1
		switch database {
		case "postgresql":
			PrepareTestForPostgresSQL()
			PrepareTestData([]string{"delete from admin_job"}, []string{})
		default:
			log.Fatalf("invalid database: %s", database)
		}
		result = testForAll(m)

		if result != 0 {
			os.Exit(result)
		}
	}
}

func testForAll(m *testing.M) int {
	cleanByUser(username)

	rc := m.Run()
	clearAll()
	return rc
}

func clearAll() {
	tables := []string{"project_member",
		"project_metadata", "repository", "replication_policy",
		"registry", "replication_execution", "replication_task",
		"replication_schedule_job", "project", "harbor_user"}
	for _, t := range tables {
		if err := ClearTable(t); err != nil {
			log.Errorf("Failed to clear table: %s,error: %v", t, err)
		}
	}
}

func TestRegister(t *testing.T) {

	user := models.User{
		Username: username,
		Email:    "tester01@vmware.com",
		Password: password,
		Realname: "tester01",
		Comment:  "register",
	}

	_, err := Register(user)
	if err != nil {
		t.Errorf("Error occurred in Register: %v", err)
	}

	// Check if user registered successfully.
	queryUser := models.User{
		Username: username,
	}
	newUser, err := GetUser(queryUser)
	if err != nil {
		t.Errorf("Error occurred in GetUser: %v", err)
	}

	if newUser.Username != username {
		t.Errorf("Username does not match, expected: %s, actual: %s", username, newUser.Username)
	}
	if newUser.Email != "tester01@vmware.com" {
		t.Errorf("Email does not match, expected: %s, actual: %s", "tester01@vmware.com", newUser.Email)
	}
}

func TestUserExists(t *testing.T) {
	var exists bool
	var err error

	exists, err = UserExists(models.User{Username: username}, "username")
	if err != nil {
		t.Errorf("Error occurred in UserExists: %v", err)
	}
	if !exists {
		t.Errorf("User %s was inserted but does not exist", username)
	}
	exists, err = UserExists(models.User{Email: "tester01@vmware.com"}, "email")

	if err != nil {
		t.Errorf("Error occurred in UserExists: %v", err)
	}
	if !exists {
		t.Errorf("User with email %s inserted but does not exist", "tester01@vmware.com")
	}
	exists, err = UserExists(models.User{Username: "NOTHERE"}, "username")
	if err != nil {
		t.Errorf("Error occurred in UserExists: %v", err)
	}
	if exists {
		t.Errorf("User %s was not inserted but does exist", "NOTHERE")
	}
}

func TestLoginByUserName(t *testing.T) {

	userQuery := models.User{
		Username: username,
		Password: "Abc12345",
	}

	loginUser, err := LoginByDb(models.AuthModel{
		Principal: userQuery.Username,
		Password:  userQuery.Password,
	})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}
	if loginUser == nil {
		t.Errorf("No found for user logined by username and password: %v", userQuery)
	}

	if loginUser.Username != username {
		t.Errorf("User's username does not match after login, expected: %s, actual: %s", username, loginUser.Username)
	}
}

func TestLoginByEmail(t *testing.T) {

	userQuery := models.User{
		Email:    "tester01@vmware.com",
		Password: "Abc12345",
	}

	loginUser, err := LoginByDb(models.AuthModel{
		Principal: userQuery.Email,
		Password:  userQuery.Password,
	})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}
	if loginUser == nil {
		t.Errorf("No found for user logined by email and password : %v", userQuery)
	}
	if loginUser.Username != username {
		t.Errorf("User's username does not match after login, expected: %s, actual: %s", username, loginUser.Username)
	}
}

var currentUser *models.User

func TestGetUser(t *testing.T) {
	queryUser := models.User{
		Username: username,
		Email:    "tester01@vmware.com",
	}
	var err error
	currentUser, err = GetUser(queryUser)
	if err != nil {
		t.Errorf("Error occurred in GetUser: %v", err)
	}
	if currentUser == nil {
		t.Errorf("No user found queried by user query: %+v", queryUser)
	}
	if currentUser.Email != "tester01@vmware.com" {
		t.Errorf("the user's email does not match, expected: tester01@vmware.com, actual: %s", currentUser.Email)
	}

	queryUser = models.User{}
	_, err = GetUser(queryUser)
	assert.NotNil(t, err)
}

func TestListUsers(t *testing.T) {
	users, err := ListUsers(nil)
	if err != nil {
		t.Errorf("Error occurred in ListUsers: %v", err)
	}
	users2, err := ListUsers(&models.UserQuery{Username: username})
	if len(users2) != 1 {
		t.Errorf("Expect one user in list, but the acutal length is %d, the list: %+v", len(users), users)
	}
	if users2[0].Username != username {
		t.Errorf("The username in result list does not match, expected: %s, actual: %s", username, users2[0].Username)
	}

	users3, err := ListUsers(&models.UserQuery{Username: username, Pagination: &models.Pagination{Page: 2, Size: 1}})
	if len(users3) != 0 {
		t.Errorf("Expect no user in list, but the acutal length is %d, the list: %+v", len(users3), users3)
	}
}

func TestResetUserPassword(t *testing.T) {
	uuid := utils.GenerateRandomString()

	err := UpdateUserResetUUID(models.User{ResetUUID: uuid, Email: currentUser.Email})
	if err != nil {
		t.Errorf("Error occurred in UpdateUserResetUuid: %v", err)
	}

	err = ResetUserPassword(
		models.User{
			UserID:          currentUser.UserID,
			PasswordVersion: utils.SHA256,
			ResetUUID:       uuid,
			Salt:            currentUser.Salt}, "HarborTester12345")
	if err != nil {
		t.Errorf("Error occurred in ResetUserPassword: %v", err)
	}

	loginedUser, err := LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "HarborTester12345"})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}

	if loginedUser.Username != username {
		t.Errorf("The username returned by Login does not match, expected: %s, acutal: %s", username, loginedUser.Username)
	}
}

func TestChangeUserPassword(t *testing.T) {
	user := models.User{UserID: currentUser.UserID}
	query, err := GetUser(user)
	if err != nil {
		t.Errorf("Error occurred when get user salt")
	}
	currentUser.Salt = query.Salt
	err = ChangeUserPassword(
		models.User{
			UserID:          currentUser.UserID,
			Password:        "NewHarborTester12345",
			PasswordVersion: utils.SHA256,
			Salt:            currentUser.Salt})
	if err != nil {
		t.Errorf("Error occurred in ChangeUserPassword: %v", err)
	}

	loginedUser, err := LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "NewHarborTester12345"})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}

	if loginedUser.Username != username {
		t.Errorf("The username returned by Login does not match, expected: %s, acutal: %s", username, loginedUser.Username)
	}
}
func TestAddProject(t *testing.T) {

	project := models.Project{
		OwnerID:      currentUser.UserID,
		Name:         projectName,
		CreationTime: time.Now(),
		OwnerName:    currentUser.Username,
	}

	_, err := AddProject(project)
	if err != nil {
		t.Errorf("Error occurred in AddProject: %v", err)
	}

	newProject, err := GetProjectByName(projectName)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}
	if newProject == nil {
		t.Errorf("No project found queried by project name: %v", projectName)
	}
}

var currentProject *models.Project

func TestGetProject(t *testing.T) {
	var err error
	currentProject, err = GetProjectByName(projectName)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}
	if currentProject == nil {
		t.Errorf("No project found queried by project name: %v", projectName)
	}
	if currentProject.Name != projectName {
		t.Errorf("Project name does not match, expected: %s, actual: %s", projectName, currentProject.Name)
	}
}

func TestGetProjectById(t *testing.T) {
	id := currentProject.ProjectID
	p, err := GetProjectByID(id)
	if err != nil {
		t.Errorf("Error in GetProjectById: %v, id: %d", err, id)
	}
	if p.Name != currentProject.Name {
		t.Errorf("project name does not match, expected: %s, actual: %s", currentProject.Name, p.Name)
	}
}

func TestGetTotalOfProjects(t *testing.T) {
	total, err := GetTotalOfProjects(nil)
	if err != nil {
		t.Fatalf("failed to get total of projects: %v", err)
	}

	if total != 2 {
		t.Errorf("unexpected total: %d != 2", total)
	}
}

func TestGetProjects(t *testing.T) {
	projects, err := GetProjects(nil)
	if err != nil {
		t.Errorf("Error occurred in GetProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("Expected length of projects is 2, but actual: %d, the projects: %+v", len(projects), projects)
	}
	if projects[1].Name != projectName {
		t.Errorf("Expected project name in the list: %s, actual: %s", projectName, projects[1].Name)
	}
}

// isAdminRole returns whether the user is admin.
func isAdminRole(userIDOrUsername interface{}) (bool, error) {
	u := models.User{}

	switch v := userIDOrUsername.(type) {
	case int:
		u.UserID = v
	case string:
		u.Username = v
	default:
		return false, fmt.Errorf("invalid parameter, only int and string are supported: %v", userIDOrUsername)
	}

	if u.UserID == NonExistUserID && len(u.Username) == 0 {
		return false, nil
	}

	user, err := GetUser(u)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return user.SysAdminFlag, nil
}

func TestToggleAdminRole(t *testing.T) {
	err := ToggleUserAdminRole(currentUser.UserID, true)
	if err != nil {
		t.Errorf("Error in toggle ToggleUserAdmin role: %v, user: %+v", err, currentUser)
	}
	isAdmin, err := isAdminRole(currentUser.UserID)
	if err != nil {
		t.Errorf("Error in IsAdminRole: %v, user id: %d", err, currentUser.UserID)
	}
	if !isAdmin {
		t.Errorf("User is not admin after toggled, user id: %d", currentUser.UserID)
	}
	err = ToggleUserAdminRole(currentUser.UserID, false)
	if err != nil {
		t.Errorf("Error in toggle ToggleUserAdmin role: %v, user: %+v", err, currentUser)
	}
	isAdmin, err = isAdminRole(currentUser.UserID)
	if err != nil {
		t.Errorf("Error in IsAdminRole: %v, user id: %d", err, currentUser.UserID)
	}
	if isAdmin {
		t.Errorf("User is still admin after toggled, user id: %d", currentUser.UserID)
	}
}

func TestChangeUserProfile(t *testing.T) {
	user := models.User{UserID: currentUser.UserID, Email: username + "@163.com", Realname: "test", Comment: "Unit Test"}
	err := ChangeUserProfile(user)
	if err != nil {
		t.Errorf("Error occurred in ChangeUserProfile: %v", err)
	}
	loginedUser, err := GetUser(models.User{UserID: currentUser.UserID})
	if err != nil {
		t.Errorf("Error occurred in GetUser: %v", err)
	}
	if loginedUser != nil {
		if loginedUser.Email != username+"@163.com" {
			t.Errorf("user email does not update, expected: %s, acutal: %s", username+"@163.com", loginedUser.Email)
		}
		if loginedUser.Realname != "test" {
			t.Errorf("user realname does not update, expected: %s, acutal: %s", "test", loginedUser.Realname)
		}
		if loginedUser.Comment != "Unit Test" {
			t.Errorf("user email does not update, expected: %s, acutal: %s", "Unit Test", loginedUser.Comment)
		}
	}
}

var targetID, policyID, policyID2, policyID3, jobID, jobID2, jobID3 int64

func TestGetOrmer(t *testing.T) {
	o := GetOrmer()
	if o == nil {
		t.Errorf("Error get ormer.")
	}
}

func TestAddRepository(t *testing.T) {
	repoRecord := models.RepoRecord{
		Name:        currentProject.Name + "/" + repositoryName,
		ProjectID:   currentProject.ProjectID,
		Description: "testing repo",
		PullCount:   0,
		StarCount:   0,
	}

	err := AddRepository(repoRecord)
	if err != nil {
		t.Errorf("Error occurred in AddRepository: %v", err)
	}

	newRepoRecord, err := GetRepositoryByName(currentProject.Name + "/" + repositoryName)
	if err != nil {
		t.Errorf("Error occurred in GetRepositoryByName: %v", err)
	}
	if newRepoRecord == nil {
		t.Errorf("No repository found queried by repository name: %v", currentProject.Name+"/"+repositoryName)
	}
}

var currentRepository *models.RepoRecord

func TestGetRepositoryByName(t *testing.T) {
	var err error
	currentRepository, err = GetRepositoryByName(currentProject.Name + "/" + repositoryName)
	if err != nil {
		t.Errorf("Error occurred in GetRepositoryByName: %v", err)
	}
	if currentRepository == nil {
		t.Errorf("No repository found queried by repository name: %v", currentProject.Name+"/"+repositoryName)
	}
	if currentRepository.Name != currentProject.Name+"/"+repositoryName {
		t.Errorf("Repository name does not match, expected: %s, actual: %s", currentProject.Name+"/"+repositoryName, currentProject.Name)
	}
}

func TestDeleteRepository(t *testing.T) {
	err := DeleteRepository(currentRepository.Name)
	if err != nil {
		t.Errorf("Error occurred in DeleteRepository: %v", err)
	}
	repository, err := GetRepositoryByName(currentRepository.Name)
	if err != nil {
		t.Errorf("Error occurred in GetRepositoryByName: %v", err)
	}
	if repository != nil {
		t.Errorf("repository is not nil after deletion, repository: %+v", repository)
	}
}

func TestIsSuperUser(t *testing.T) {
	assert := assert.New(t)
	assert.True(IsSuperUser("admin"))
	assert.False(IsSuperUser("none"))
}

func TestSaveConfigEntries(t *testing.T) {
	configEntries := []models.ConfigEntry{
		{
			Key:   "teststringkey",
			Value: "192.168.111.211",
		},
		{
			Key:   "testboolkey",
			Value: "true",
		},
		{
			Key:   "testnumberkey",
			Value: "5",
		},
		{
			Key:   common.CfgDriverDB,
			Value: "db",
		},
	}
	err := SaveConfigEntries(configEntries)
	if err != nil {
		t.Fatalf("failed to save configuration to database %v", err)
	}
	readEntries, err := GetConfigEntries()
	if err != nil {
		t.Fatalf("Failed to get configuration from database %v", err)
	}
	findItem := 0
	for _, entry := range readEntries {
		switch entry.Key {
		case "teststringkey":
			if "192.168.111.211" == entry.Value {
				findItem++
			}
		case "testnumberkey":
			if "5" == entry.Value {
				findItem++
			}
		case "testboolkey":
			if "true" == entry.Value {
				findItem++
			}
		default:
		}
	}
	if findItem != 3 {
		t.Fatalf("Should update 3 configuration but only update %d", findItem)
	}

	configEntries = []models.ConfigEntry{
		{
			Key:   "teststringkey",
			Value: "192.168.111.215",
		},
		{
			Key:   "testboolkey",
			Value: "false",
		},
		{
			Key:   "testnumberkey",
			Value: "7",
		},
		{
			Key:   common.CfgDriverDB,
			Value: "db",
		},
	}
	err = SaveConfigEntries(configEntries)
	if err != nil {
		t.Fatalf("failed to save configuration to database %v", err)
	}
	readEntries, err = GetConfigEntries()
	if err != nil {
		t.Fatalf("Failed to get configuration from database %v", err)
	}
	findItem = 0
	for _, entry := range readEntries {
		switch entry.Key {
		case "teststringkey":
			if "192.168.111.215" == entry.Value {
				findItem++
			}
		case "testnumberkey":
			if "7" == entry.Value {
				findItem++
			}
		case "testboolkey":
			if "false" == entry.Value {
				findItem++
			}
		default:
		}
	}
	if findItem != 3 {
		t.Fatalf("Should update 3 configuration but only update %d", findItem)
	}
}

func TestIsDupRecError(t *testing.T) {
	assert.True(t, IsDupRecErr(fmt.Errorf("pq: duplicate key value violates unique constraint \"properties_k_key\"")))
	assert.False(t, IsDupRecErr(fmt.Errorf("other error")))
}

func TestWithTransaction(t *testing.T) {
	reference := "transaction"

	quota := models.Quota{
		Reference:   reference,
		ReferenceID: "1",
		Hard:        "{}",
	}

	failed := func(o orm.Ormer) error {
		o.Insert(&quota)

		return fmt.Errorf("failed")
	}

	var quotaID int64
	success := func(o orm.Ormer) error {
		id, err := o.Insert(&quota)
		if err != nil {
			return err
		}

		quotaID = id
		return nil
	}

	assert := assert.New(t)

	if assert.Error(WithTransaction(failed)) {
		var quota models.Quota
		quota.Reference = reference
		quota.ReferenceID = "1"
		err := GetOrmer().Read(&quota, "reference", "reference_id")
		assert.Error(err)
		assert.False(quota.ID != 0)
	}

	if assert.Nil(WithTransaction(success)) {
		var quota models.Quota
		quota.Reference = reference
		quota.ReferenceID = "1"
		err := GetOrmer().Read(&quota, "reference", "reference_id")
		assert.Nil(err)
		assert.True(quota.ID != 0)
		assert.Equal(quotaID, quota.ID)

		GetOrmer().Delete(&models.Quota{ID: quotaID}, "id")
	}
}
