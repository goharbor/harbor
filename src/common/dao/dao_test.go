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
	"github.com/goharbor/harbor/src/common/utils/log"
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

	err = execUpdate(o, `delete
		from access_log
		where username = ?
		`, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete
		from access_log
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
		"project_metadata", "access_log", "repository", "replication_policy",
		"registry", "replication_execution", "replication_task", "img_scan_job",
		"replication_schedule_job", "img_scan_overview", "clair_vuln_timestamp", "project", "harbor_user"}
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

	err = ResetUserPassword(models.User{UserID: currentUser.UserID, Password: "HarborTester12345", ResetUUID: uuid, Salt: currentUser.Salt})
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
	err = ChangeUserPassword(models.User{UserID: currentUser.UserID, Password: "NewHarborTester12345", Salt: currentUser.Salt})
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

func TestGetAccessLog(t *testing.T) {

	accessLog := models.AccessLog{
		Username:  currentUser.Username,
		ProjectID: currentProject.ProjectID,
		RepoName:  currentProject.Name + "/",
		RepoTag:   "N/A",
		GUID:      "N/A",
		Operation: "create",
		OpTime:    time.Now(),
	}
	if err := AddAccessLog(accessLog); err != nil {
		t.Errorf("failed to add access log: %v", err)
	}

	query := &models.LogQueryParam{
		Username:   currentUser.Username,
		ProjectIDs: []int64{currentProject.ProjectID},
	}
	accessLogs, err := GetAccessLogs(query)
	if err != nil {
		t.Errorf("Error occurred in GetAccessLog: %v", err)
	}
	if len(accessLogs) != 1 {
		t.Errorf("The length of accesslog list should be 1, actual: %d", len(accessLogs))
	}
	if accessLogs[0].RepoName != projectName+"/" {
		t.Errorf("The project name does not match, expected: %s, actual: %s", projectName+"/", accessLogs[0].RepoName)
	}
}

func TestGetTotalOfAccessLogs(t *testing.T) {
	query := &models.LogQueryParam{
		Username:   currentUser.Username,
		ProjectIDs: []int64{currentProject.ProjectID},
	}
	total, err := GetTotalOfAccessLogs(query)
	if err != nil {
		t.Fatalf("failed to get total of access log: %v", err)
	}

	if total != 1 {
		t.Errorf("unexpected total %d != %d", total, 1)
	}
}

func TestAddAccessLog(t *testing.T) {
	var err error
	var accessLogList []models.AccessLog
	accessLog := models.AccessLog{
		Username:  currentUser.Username,
		ProjectID: currentProject.ProjectID,
		RepoName:  currentProject.Name + "/",
		RepoTag:   repoTag,
		GUID:      "N/A",
		Operation: "create",
		OpTime:    time.Now(),
	}
	err = AddAccessLog(accessLog)
	if err != nil {
		t.Errorf("Error occurred in AddAccessLog: %v", err)
	}

	query := &models.LogQueryParam{
		Username:   accessLog.Username,
		ProjectIDs: []int64{accessLog.ProjectID},
		Repository: accessLog.RepoName,
		Tag:        accessLog.RepoTag,
		Operations: []string{accessLog.Operation},
	}
	accessLogList, err = GetAccessLogs(query)
	if err != nil {
		t.Errorf("Error occurred in GetAccessLog: %v", err)
	}
	if len(accessLogList) != 1 {
		t.Errorf("The length of accesslog list should be 1, actual: %d", len(accessLogList))
	}
	if accessLogList[0].RepoName != projectName+"/" {
		t.Errorf("The project name does not match, expected: %s, actual: %s", projectName+"/", accessLogList[0].RepoName)
	}
	if accessLogList[0].RepoTag != repoTag {
		t.Errorf("The repo tag does not match, expected: %s, actual: %s", repoTag, accessLogList[0].RepoTag)
	}
}

func TestCountPull(t *testing.T) {
	var err error
	if err = AddAccessLog(models.AccessLog{
		Username:  currentUser.Username,
		ProjectID: currentProject.ProjectID,
		RepoName:  currentProject.Name + "/tomcat",
		RepoTag:   repoTag2,
		Operation: "pull",
		OpTime:    time.Now(),
	}); err != nil {
		t.Errorf("Error occurred in AccessLog: %v", err)
	}

	if err = AddAccessLog(models.AccessLog{
		Username:  currentUser.Username,
		ProjectID: currentProject.ProjectID,
		RepoName:  currentProject.Name + "/tomcat",
		RepoTag:   repoTag2,
		Operation: "pull",
		OpTime:    time.Now(),
	}); err != nil {
		t.Errorf("Error occurred in AccessLog: %v", err)
	}

	if err = AddAccessLog(models.AccessLog{
		Username:  currentUser.Username,
		ProjectID: currentProject.ProjectID,
		RepoName:  currentProject.Name + "/tomcat",
		RepoTag:   repoTag2,
		Operation: "pull",
		OpTime:    time.Now(),
	}); err != nil {
		t.Errorf("Error occurred in AccessLog: %v", err)
	}

	pullCount, err := CountPull(currentProject.Name + "/tomcat")
	if err != nil {
		t.Errorf("Error occurred in CountPull: %v", err)
	}
	if pullCount != 3 {
		t.Errorf("The access log pull count does not match, expected: 3, actual: %d", pullCount)
	}
}

/*
func TestProjectExists(t *testing.T) {
	var exists bool
	var err error
	exists, err = ProjectExists(currentProject.ProjectID)
	if err != nil {
		t.Errorf("Error occurred in ProjectExists: %v", err)
	}
	if !exists {
		t.Errorf("The project with id: %d, does not exist", currentProject.ProjectID)
	}
	exists, err = ProjectExists(currentProject.Name)
	if err != nil {
		t.Errorf("Error occurred in ProjectExists: %v", err)
	}
	if !exists {
		t.Errorf("The project with name: %s, does not exist", currentProject.Name)
	}
}
*/
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

func TestGetUserProjectRoles(t *testing.T) {
	r, err := GetUserProjectRoles(currentUser.UserID, currentProject.ProjectID, common.UserMember)
	if err != nil {
		t.Errorf("Error happened in GetUserProjectRole: %v, userID: %+v, project Id: %d", err, currentUser.UserID, currentProject.ProjectID)
	}

	// Get the size of current user project role.
	if len(r) != 1 {
		t.Errorf("The user, id: %d, should only have one role in project, id: %d, but actual: %d", currentUser.UserID, currentProject.ProjectID, len(r))
	}

	if r[0].Name != "projectAdmin" {
		t.Errorf("the expected rolename is: projectAdmin, actual: %s", r[0].Name)
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

func TestGetRoleByID(t *testing.T) {
	r, err := GetRoleByID(models.PROJECTADMIN)
	if err != nil {
		t.Errorf("Failed to call GetRoleByID: %v", err)
	}
	if r == nil || r.Name != "projectAdmin" || r.RoleCode != "MDRWS" {
		t.Errorf("Role does not match for role id: %d, actual: %+v", models.PROJECTADMIN, r)
	}
	r, err = GetRoleByID(9999)
	if err != nil {
		t.Errorf("Failed to call GetRoleByID: %v", err)
	}
	if r != nil {
		t.Errorf("Role should nil for non-exist id 9999, actual: %+v", r)
	}
}

func TestToggleAdminRole(t *testing.T) {
	err := ToggleUserAdminRole(currentUser.UserID, true)
	if err != nil {
		t.Errorf("Error in toggle ToggleUserAdmin role: %v, user: %+v", err, currentUser)
	}
	isAdmin, err := IsAdminRole(currentUser.UserID)
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
	isAdmin, err = IsAdminRole(currentUser.UserID)
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

func TestIncreasePullCount(t *testing.T) {
	if err := IncreasePullCount(currentRepository.Name); err != nil {
		log.Errorf("Error happens when increasing pull count: %v", currentRepository.Name)
	}

	repository, err := GetRepositoryByName(currentRepository.Name)
	if err != nil {
		t.Errorf("Error occurred in GetRepositoryByName: %v", err)
	}

	if repository.PullCount != 1 {
		t.Errorf("repository pull count is not 1 after IncreasePullCount, expected: 1, actual: %d", repository.PullCount)
	}
}

func TestRepositoryExists(t *testing.T) {
	var exists bool
	exists = RepositoryExists(currentRepository.Name)
	if !exists {
		t.Errorf("The repository with name: %s, does not exist", currentRepository.Name)
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

var sj1 = models.ScanJob{
	Status:     models.JobPending,
	Repository: "library/ubuntu",
	Tag:        "14.04",
}

var sj2 = models.ScanJob{
	Status:     models.JobPending,
	Repository: "library/ubuntu",
	Tag:        "15.10",
	Digest:     "sha256:0204dc6e09fa57ab99ac40e415eb637d62c8b2571ecbbc9ca0eb5e2ad2b5c56f",
}

func TestAddScanJob(t *testing.T) {
	assert := assert.New(t)
	id, err := AddScanJob(sj1)
	assert.Nil(err)
	r1, err := GetScanJob(id)
	assert.Nil(err)
	assert.Equal(sj1.Tag, r1.Tag)
	assert.Equal(sj1.Status, r1.Status)
	assert.Equal(sj1.Repository, r1.Repository)
	err = ClearTable(models.ScanJobTable)
	assert.Nil(err)
}

func TestGetScanJobs(t *testing.T) {
	assert := assert.New(t)
	_, err := AddScanJob(sj1)
	assert.Nil(err)
	id2, err := AddScanJob(sj1)
	assert.Nil(err)
	_, err = AddScanJob(sj2)
	assert.Nil(err)
	r, err := GetScanJobsByImage("library/ubuntu", "14.04")
	assert.Nil(err)
	assert.Equal(2, len(r))
	assert.Equal(id2, r[0].ID)
	r, err = GetScanJobsByImage("library/ubuntu", "14.04", 1)
	assert.Nil(err)
	assert.Equal(1, len(r))
	r, err = GetScanJobsByDigest("sha256:nono")
	assert.Nil(err)
	assert.Equal(0, len(r))
	r, err = GetScanJobsByDigest(sj2.Digest)
	assert.Equal(1, len(r))
	assert.Equal(sj2.Tag, r[0].Tag)
	assert.Nil(err)
	err = ClearTable(models.ScanJobTable)
	assert.Nil(err)
}

func TestSetScanJobUUID(t *testing.T) {
	uuid := "u-scan-job-uuid"
	assert := assert.New(t)
	id, err := AddScanJob(sj1)
	assert.Nil(err)
	err = SetScanJobUUID(id, uuid)
	assert.Nil(err)
	j, err := GetScanJob(id)
	assert.Nil(err)
	assert.Equal(uuid, j.UUID)
	err = ClearTable(models.ScanJobTable)
	assert.Nil(err)

}

func TestUpdateScanJobStatus(t *testing.T) {
	assert := assert.New(t)
	id, err := AddScanJob(sj1)
	assert.Nil(err)
	err = UpdateScanJobStatus(id, "newstatus")
	assert.Nil(err)
	j, err := GetScanJob(id)
	assert.Nil(err)
	assert.Equal("newstatus", j.Status)
	err = ClearTable(models.ScanJobTable)
	assert.Nil(err)
}

func TestImgScanOverview(t *testing.T) {
	assert := assert.New(t)
	err := ClearTable(models.ScanOverviewTable)
	assert.Nil(err)
	digest := "sha256:0204dc6e09fa57ab99ac40e415eb637d62c8b2571ecbbc9ca0eb5e2ad2b5c56f"
	res, err := GetImgScanOverview(digest)
	assert.Nil(err)
	assert.Nil(res)
	err = SetScanJobForImg(digest, 33)
	assert.Nil(err)
	res, err = GetImgScanOverview(digest)
	assert.Nil(err)
	assert.Equal(int64(33), res.JobID)
	err = SetScanJobForImg(digest, 22)
	assert.Nil(err)
	res, err = GetImgScanOverview(digest)
	assert.Nil(err)
	assert.Equal(int64(22), res.JobID)
	pk := "22-sha256:sdfsdfarfwefwr23r43t34ggregergerger"
	comp := &models.ComponentsOverview{
		Total: 2,
		Summary: []*models.ComponentsOverviewEntry{
			{
				Sev:   int(models.SevMedium),
				Count: 2,
			},
		},
	}
	err = UpdateImgScanOverview(digest, pk, models.SevMedium, comp)
	assert.Nil(err)
	res, err = GetImgScanOverview(digest)
	assert.Nil(err)
	assert.Equal(pk, res.DetailsKey)
	assert.Equal(int(models.SevMedium), res.Sev)
	assert.Equal(2, res.CompOverview.Summary[0].Count)
}

func TestVulnTimestamp(t *testing.T) {

	assert := assert.New(t)
	err := ClearTable(models.ClairVulnTimestampTable)
	assert.Nil(err)
	ns := "ubuntu:14"
	res, err := ListClairVulnTimestamps()
	assert.Nil(err)
	assert.Equal(0, len(res))
	err = SetClairVulnTimestamp(ns, time.Now())
	assert.Nil(err)
	res, err = ListClairVulnTimestamps()
	assert.Nil(err)
	assert.Equal(1, len(res))
	assert.Equal(ns, res[0].Namespace)
	old := time.Now()
	t.Logf("Sleep 3 seconds")
	time.Sleep(3 * time.Second)
	err = SetClairVulnTimestamp(ns, time.Now())
	assert.Nil(err)
	res, err = ListClairVulnTimestamps()
	assert.Nil(err)
	assert.Equal(1, len(res))

	d := res[0].LastUpdate.Sub(old)
	if d < 2*time.Second {
		t.Errorf("Delta should be larger than 2 seconds! old: %v, lastupdate: %v", old, res[0].LastUpdate)
	}
}

func TestListScanOverviews(t *testing.T) {
	assert := assert.New(t)
	err := ClearTable(models.ScanOverviewTable)
	assert.Nil(err)
	l, err := ListImgScanOverviews()
	assert.Nil(err)
	assert.Equal(0, len(l))
	err = ClearTable(models.ScanOverviewTable)
	assert.Nil(err)
}

func TestGetScanJobsByStatus(t *testing.T) {
	assert := assert.New(t)
	err := ClearTable(models.ScanOverviewTable)
	assert.Nil(err)
	id, err := AddScanJob(sj1)
	assert.Nil(err)
	err = UpdateScanJobStatus(id, models.JobRunning)
	assert.Nil(err)
	r1, err := GetScanJobsByStatus(models.JobPending, models.JobCanceled)
	assert.Nil(err)
	assert.Equal(0, len(r1))
	r2, err := GetScanJobsByStatus(models.JobPending, models.JobRunning)
	assert.Nil(err)
	assert.Equal(1, len(r2))
	assert.Equal(sj1.Repository, r2[0].Repository)
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
	assert.True(t, isDupRecErr(fmt.Errorf("pq: duplicate key value violates unique constraint \"properties_k_key\"")))
	assert.False(t, isDupRecErr(fmt.Errorf("other error")))
}

func TestWithTransaction(t *testing.T) {
	quota := models.Quota{
		Reference:   "project",
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
		quota.Reference = "project"
		quota.ReferenceID = "1"
		err := GetOrmer().Read(&quota, "reference", "reference_id")
		assert.Error(err)
		assert.False(quota.ID != 0)
	}

	if assert.Nil(WithTransaction(success)) {
		var quota models.Quota
		quota.Reference = "project"
		quota.ReferenceID = "1"
		err := GetOrmer().Read(&quota, "reference", "reference_id")
		assert.Nil(err)
		assert.True(quota.ID != 0)
		assert.Equal(quotaID, quota.ID)

		GetOrmer().Delete(&models.Quota{ID: quotaID}, "id")
	}
}
