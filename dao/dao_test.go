/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package dao

import (
	"os"
	"testing"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
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

func clearUp(username string) {
	var err error

	o := orm.NewOrm()
	o.Begin()

	err = execUpdate(o, `delete pm 
		from project_member pm 
		join user u 
		on pm.user_id = u.user_id 
		where u.username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete pm 
		from project_member pm
		join project p 
		on pm.project_id = p.project_id 
		where p.name = ?`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete al 
		from access_log al
		join user u 
		on al.user_id = u.user_id 
		where u.username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete al 
		from access_log al
		join project p 
		on al.project_id = p.project_id 
		where p.name = ?`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from project where name = ?`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from user where username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from replication_job where id < 99`)
	if err != nil {
		log.Error(err)
	}
	err = execUpdate(o, `delete from replication_policy where id < 99`)
	if err != nil {
		log.Error(err)
	}
	err = execUpdate(o, `delete from replication_target where id < 99`)
	if err != nil {
		log.Error(err)
	}
	o.Commit()
}

const username string = "Tester01"
const projectName string = "test_project"
const repoTag string = "test1.1"
const repoTag2 string = "test1.2"
const SysAdmin int = 1
const projectAdmin int = 2
const developer int = 3
const guest int = 4

const publicityOn = 1
const publicityOff = 0

func TestMain(m *testing.M) {

	dbHost := os.Getenv("DB_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable DB_HOST is not set")
	}
	dbUser := os.Getenv("DB_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable DB_USR is not set")
	}
	dbPort := os.Getenv("DB_PORT")
	if len(dbPort) == 0 {
		log.Fatalf("environment variable DB_PORT is not set")
	}
	dbPassword := os.Getenv("DB_PWD")

	log.Infof("DB_HOST: %s, DB_USR: %s, DB_PORT: %s, DB_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	os.Setenv("MYSQL_HOST", dbHost)
	os.Setenv("MYSQL_PORT", dbPort)
	os.Setenv("MYSQL_USR", dbUser)
	os.Setenv("MYSQL_PWD", dbPassword)
	os.Setenv("AUTH_MODE", "db_auth")
	InitDB()
	clearUp(username)
	os.Exit(m.Run())

}

func TestRegister(t *testing.T) {

	user := models.User{
		Username: username,
		Email:    "tester01@vmware.com",
		Password: "Abc12345",
		Realname: "tester01",
		Comment:  "register",
	}

	_, err := Register(user)
	if err != nil {
		t.Errorf("Error occurred in Register: %v", err)
	}

	//Check if user registered successfully.
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
}

func TestListUsers(t *testing.T) {
	users, err := ListUsers(models.User{})
	if err != nil {
		t.Errorf("Error occurred in ListUsers: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expect one user in list, but the acutal length is %d, the list: %+v", len(users), users)
	}
	users2, err := ListUsers(models.User{Username: username})
	if len(users2) != 1 {
		t.Errorf("Expect one user in list, but the acutal length is %d, the list: %+v", len(users), users)
	}
	if users2[0].Username != username {
		t.Errorf("The username in result list does not match, expected: %s, actual: %s", username, users2[0].Username)
	}
}

func TestResetUserPassword(t *testing.T) {
	uuid, err := GenerateRandomString()
	if err != nil {
		t.Errorf("Error occurred in GenerateRandomString: %v", err)
	}

	err = UpdateUserResetUUID(models.User{ResetUUID: uuid, Email: currentUser.Email})
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
	err := ChangeUserPassword(models.User{UserID: currentUser.UserID, Password: "NewHarborTester12345", Salt: currentUser.Salt})
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

func TestChangeUserPasswordWithOldPassword(t *testing.T) {
	err := ChangeUserPassword(models.User{UserID: currentUser.UserID, Password: "NewerHarborTester12345", Salt: currentUser.Salt}, "NewHarborTester12345")
	if err != nil {
		t.Errorf("Error occurred in ChangeUserPassword: %v", err)
	}
	loginedUser, err := LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "NewerHarborTester12345"})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}
	if loginedUser.Username != username {
		t.Errorf("The username returned by Login does not match, expected: %s, acutal: %s", username, loginedUser.Username)
	}
}

func TestChangeUserPasswordWithIncorrectOldPassword(t *testing.T) {
	err := ChangeUserPassword(models.User{UserID: currentUser.UserID, Password: "NNewerHarborTester12345", Salt: currentUser.Salt}, "WrongNewerHarborTester12345")
	if err == nil {
		t.Errorf("Error does not occurred due to old password is incorrect.")
	}
	loginedUser, err := LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "NNewerHarborTester12345"})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}
	if loginedUser != nil {
		t.Errorf("The login user is not nil, acutal: %+v", loginedUser)
	}
}

func TestQueryRelevantProjectsWhenNoProjectAdded(t *testing.T) {
	projects, err := SearchProjects(currentUser.UserID)
	if err != nil {
		t.Errorf("Error occurred in QueryRelevantProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("Expected only one project in DB, but actual: %d", len(projects))
	}
	if projects[0].Name != "library" {
		t.Errorf("There name of the project does not match, expected: %s, actual: %s", "library", projects[0].Name)
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
	queryAccessLog := models.AccessLog{
		UserID:    currentUser.UserID,
		ProjectID: currentProject.ProjectID,
	}
	accessLogs, err := GetAccessLogs(queryAccessLog)
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

func TestAddAccessLog(t *testing.T) {
	var err error
	var accessLogList []models.AccessLog
	accessLog := models.AccessLog{
		UserID:    currentUser.UserID,
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
	accessLogList, err = GetAccessLogs(accessLog)
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

func TestAccessLog(t *testing.T) {
	var err error
	var accessLogList []models.AccessLog
	accessLog := models.AccessLog{
		UserID:    currentUser.UserID,
		ProjectID: currentProject.ProjectID,
		RepoName:  currentProject.Name + "/",
		RepoTag:   repoTag2,
		Operation: "create",
	}
	err = AccessLog(currentUser.Username, currentProject.Name, currentProject.Name+"/", repoTag2, "create")
	if err != nil {
		t.Errorf("Error occurred in AccessLog: %v", err)
	}
	accessLogList, err = GetAccessLogs(accessLog)
	if err != nil {
		t.Errorf("Error occurred in GetAccessLog: %v", err)
	}
	if len(accessLogList) != 1 {
		t.Errorf("The length of accesslog list should be 1, actual: %d", len(accessLogList))
	}
	if accessLogList[0].RepoName != projectName+"/" {
		t.Errorf("The project name does not match, expected: %s, actual: %s", projectName+"/", accessLogList[0].RepoName)
	}
	if accessLogList[0].RepoTag != repoTag2 {
		t.Errorf("The repo tag does not match, expected: %s, actual: %s", repoTag2, accessLogList[0].RepoTag)
	}
}

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

func TestGetUserByProject(t *testing.T) {
	pid := currentProject.ProjectID
	u1 := models.User{
		Username: "%%Tester%%",
	}
	u2 := models.User{
		Username: "nononono",
	}
	users, err := GetUserByProject(pid, u1)
	if err != nil {
		t.Errorf("Error happened in GetUserByProject: %v, project Id: %d, user: %+v", err, pid, u1)
	}
	if len(users) != 1 {
		t.Errorf("unexpected length of user list, expected: 1, the users list: %+v", users)
	}
	users, err = GetUserByProject(pid, u2)
	if err != nil {
		t.Errorf("Error happened in GetUserByProject: %v, project Id: %d, user: %+v", err, pid, u2)
	}
	if len(users) != 0 {
		t.Errorf("unexpected length of user list, expected: 0, the users list: %+v", users)
	}

}

func TestToggleProjectPublicity(t *testing.T) {
	err := ToggleProjectPublicity(currentProject.ProjectID, publicityOn)
	if err != nil {
		t.Errorf("Error occurred in ToggleProjectPublicity: %v", err)
	}

	currentProject, err = GetProjectByName(projectName)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}
	if currentProject.Public != publicityOn {
		t.Errorf("project, id: %d, its publicity is not on", currentProject.ProjectID)
	}
	err = ToggleProjectPublicity(currentProject.ProjectID, publicityOff)
	if err != nil {
		t.Errorf("Error occurred in ToggleProjectPublicity: %v", err)
	}

	currentProject, err = GetProjectByName(projectName)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}

	if currentProject.Public != publicityOff {
		t.Errorf("project, id: %d, its publicity is not off", currentProject.ProjectID)
	}

}

func TestIsProjectPublic(t *testing.T) {

	if isPublic := IsProjectPublic(projectName); isPublic {
		t.Errorf("project, id: %d, its publicity is not false after turning off", currentProject.ProjectID)
	}
}

func TestGetUserProjectRoles(t *testing.T) {
	r, err := GetUserProjectRoles(currentUser.UserID, currentProject.ProjectID)
	if err != nil {
		t.Errorf("Error happened in GetUserProjectRole: %v, userID: %+v, project Id: %d", err, currentUser.UserID, currentProject.ProjectID)
	}

	//Get the size of current user project role.
	if len(r) != 1 {
		t.Errorf("The user, id: %d, should only have one role in project, id: %d, but actual: %d", currentUser.UserID, currentProject.ProjectID, len(r))
	}

	if r[0].Name != "projectAdmin" {
		t.Errorf("the expected rolename is: projectAdmin, actual: %s", r[0].Name)
	}
}

func TestProjectPermission(t *testing.T) {
	roleCode, err := GetPermission(currentUser.Username, currentProject.Name)
	if err != nil {
		t.Errorf("Error occurred in GetPermission: %v", err)
	}
	if roleCode != "MDRWS" {
		t.Errorf("The expected role code is MDRWS,but actual: %s", roleCode)
	}
}

func TestGetUserRelevantProjects(t *testing.T) {
	projects, err := GetUserRelevantProjects(currentUser.UserID, "")
	if err != nil {
		t.Errorf("Error occurred in GetUserRelevantProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("Expected length of relevant projects is 1, but actual: %d, the projects: %+v", len(projects), projects)
	}
	if projects[0].Name != projectName {
		t.Errorf("Expected project name in the list: %s, actual: %s", projectName, projects[1].Name)
	}
}

func TestGetAllProjects(t *testing.T) {
	projects, err := GetAllProjects("")
	if err != nil {
		t.Errorf("Error occurred in GetAllProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("Expected length of projects is 2, but actual: %d, the projects: %+v", len(projects), projects)
	}
	if projects[1].Name != projectName {
		t.Errorf("Expected project name in the list: %s, actual: %s", projectName, projects[1].Name)
	}
}

func TestGetPublicProjects(t *testing.T) {
	projects, err := GetPublicProjects("")
	if err != nil {
		t.Errorf("Error occurred in getProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("Expected length of projects is 1, but actual: %d, the projects: %+v", len(projects), projects)
	}
	if projects[0].Name != "library" {
		t.Errorf("Expected project name in the list: %s, actual: %s", "library", projects[0].Name)
	}
}

func TestAddProjectMember(t *testing.T) {
	err := AddProjectMember(currentProject.ProjectID, 1, models.DEVELOPER)
	if err != nil {
		t.Errorf("Error occurred in AddProjectMember: %v", err)
	}

	roles, err := GetUserProjectRoles(1, currentProject.ProjectID)
	if err != nil {
		t.Errorf("Error occurred in GetUserProjectRoles: %v", err)
	}

	flag := false
	for _, role := range roles {
		if role.Name == "developer" {
			flag = true
			break
		}
	}

	if !flag {
		t.Errorf("the user which ID is 1 does not have developer privileges")
	}
}

func TestDeleteProjectMember(t *testing.T) {
	err := DeleteProjectMember(currentProject.ProjectID, 1)
	if err != nil {
		t.Errorf("Error occurred in DeleteProjectMember: %v", err)
	}

	roles, err := GetUserProjectRoles(1, currentProject.ProjectID)
	if err != nil {
		t.Errorf("Error occurred in GetUserProjectRoles: %v", err)
	}

	if len(roles) != 0 {
		t.Errorf("delete record failed from table project_member")
	}
}

func TestToggleAdminRole(t *testing.T) {
	err := ToggleUserAdminRole(*currentUser)
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
	err = ToggleUserAdminRole(*currentUser)
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

func TestDeleteUser(t *testing.T) {
	err := DeleteUser(currentUser.UserID)
	if err != nil {
		t.Errorf("Error occurred in DeleteUser: %v", err)
	}
	user, err := GetUser(*currentUser)
	if err != nil {
		t.Errorf("Error occurred in GetUser: %v", err)
	}
	if user != nil {
		t.Errorf("user is not nil after deletion, user: %+v", user)
	}
}

var targetID, policyID, policyID2, policyID3, jobID, jobID2, jobID3 int64

func TestAddRepTarget(t *testing.T) {
	target := models.RepTarget{
		Name:     "test",
		URL:      "127.0.0.1:5000",
		Username: "admin",
		Password: "admin",
	}
	//_, err := AddRepTarget(target)
	id, err := AddRepTarget(target)
	t.Logf("added target, id: %d", id)
	if err != nil {
		t.Errorf("Error occurred in AddRepTarget: %v", err)
	} else {
		targetID = id
	}
	id2 := id + 99
	tgt, err := GetRepTarget(id2)
	if err != nil {
		t.Errorf("Error occurred in GetTarget: %v, id: %d", err, id2)
	}
	if tgt != nil {
		t.Errorf("There should not be a target with id: %d", id2)
	}
	tgt, err = GetRepTarget(id)
	if err != nil {
		t.Errorf("Error occurred in GetTarget: %v, id: %d", err, id)
	}
	if tgt == nil {
		t.Errorf("Unable to find a target with id: %d", id)
	}
	if tgt.URL != "127.0.0.1:5000" {
		t.Errorf("Unexpected url in target: %s, expected 127.0.0.1:5000", tgt.URL)
	}
	if tgt.Username != "admin" {
		t.Errorf("Unexpected username in target: %s, expected admin", tgt.Username)
	}
}

func TestGetRepTargetByName(t *testing.T) {
	target, err := GetRepTarget(targetID)
	if err != nil {
		t.Fatalf("failed to get target %d: %v", targetID, err)
	}

	target2, err := GetRepTargetByName(target.Name)
	if err != nil {
		t.Fatalf("failed to get target %s: %v", target.Name, err)
	}

	if target.Name != target2.Name {
		t.Errorf("unexpected target name: %s, expected: %s", target2.Name, target.Name)
	}
}

func TestUpdateRepTarget(t *testing.T) {
	target := &models.RepTarget{
		Name:     "name",
		URL:      "http://url",
		Username: "username",
		Password: "password",
	}

	id, err := AddRepTarget(*target)
	if err != nil {
		t.Fatalf("failed to add target: %v", err)
	}
	defer func() {
		if err := DeleteRepTarget(id); err != nil {
			t.Logf("failed to delete target %d: %v", id, err)
		}
	}()

	target.ID = id
	target.Name = "new_name"
	target.URL = "http://new_url"
	target.Username = "new_username"
	target.Password = "new_password"

	if err = UpdateRepTarget(*target); err != nil {
		t.Fatalf("failed to update target: %v", err)
	}

	target, err = GetRepTarget(id)
	if err != nil {
		t.Fatalf("failed to get target %d: %v", id, err)
	}

	if target.Name != "new_name" {
		t.Errorf("unexpected name: %s, expected: %s", target.Name, "new_name")
	}

	if target.URL != "http://new_url" {
		t.Errorf("unexpected url: %s, expected: %s", target.URL, "http://new_url")
	}

	if target.Username != "new_username" {
		t.Errorf("unexpected username: %s, expected: %s", target.Username, "new_username")
	}

	if target.Password != "new_password" {
		t.Errorf("unexpected password: %s, expected: %s", target.Password, "new_password")
	}
}

func TestGetAllRepTargets(t *testing.T) {
	if _, err := GetAllRepTargets(); err != nil {
		t.Fatalf("failed to get all targets: %v", err)
	}
}

func TestFilterRepTargets(t *testing.T) {
	targets, err := FilterRepTargets("test")
	if err != nil {
		t.Fatalf("failed to get all targets: %v", err)
	}

	if len(targets) == 0 {
		t.Errorf("unexpected num of targets: %d, expected: %d", len(targets), 1)
	}
}

func TestAddRepPolicy(t *testing.T) {
	policy := models.RepPolicy{
		ProjectID:   1,
		Enabled:     1,
		TargetID:    targetID,
		Description: "whatever",
		Name:        "mypolicy",
	}
	id, err := AddRepPolicy(policy)
	t.Logf("added policy, id: %d", id)
	if err != nil {
		t.Errorf("Error occurred in AddRepPolicy: %v", err)
	} else {
		policyID = id
	}
	p, err := GetRepPolicy(id)
	if err != nil {
		t.Errorf("Error occurred in GetPolicy: %v, id: %d", err, id)
	}
	if p == nil {
		t.Errorf("Unable to find a policy with id: %d", id)
	}

	if p.Name != "mypolicy" || p.TargetID != targetID || p.Enabled != 1 || p.Description != "whatever" {
		t.Errorf("The data does not match, expected: Name: mypolicy, TargetID: %d, Enabled: 1, Description: whatever;\n result: Name: %s, TargetID: %d, Enabled: %d, Description: %s",
			targetID, p.Name, p.TargetID, p.Enabled, p.Description)
	}
	var tm = time.Now().AddDate(0, 0, -1)
	if !p.StartTime.After(tm) {
		t.Errorf("Unexpected start_time: %v", p.StartTime)
	}

}

func TestGetRepPolicyByName(t *testing.T) {
	policy, err := GetRepPolicy(policyID)
	if err != nil {
		t.Fatalf("failed to get policy %d: %v", policyID, err)
	}

	policy2, err := GetRepPolicyByName(policy.Name)
	if err != nil {
		t.Fatalf("failed to get policy %s: %v", policy.Name, err)
	}

	if policy.Name != policy2.Name {
		t.Errorf("unexpected name: %s, expected: %s", policy2.Name, policy.Name)
	}

}

func TestDisableRepPolicy(t *testing.T) {
	err := DisableRepPolicy(policyID)
	if err != nil {
		t.Errorf("Failed to disable policy, id: %d", policyID)
	}
	p, err := GetRepPolicy(policyID)
	if err != nil {
		t.Errorf("Error occurred in GetPolicy: %v, id: %d", err, policyID)
	}
	if p == nil {
		t.Errorf("Unable to find a policy with id: %d", policyID)
	}
	if p.Enabled == 1 {
		t.Errorf("The Enabled value of replication policy is still 1 after disabled, id: %d", policyID)
	}
}

func TestEnableRepPolicy(t *testing.T) {
	err := EnableRepPolicy(policyID)
	if err != nil {
		t.Errorf("Failed to disable policy, id: %d", policyID)
	}
	p, err := GetRepPolicy(policyID)
	if err != nil {
		t.Errorf("Error occurred in GetPolicy: %v, id: %d", err, policyID)
	}
	if p == nil {
		t.Errorf("Unable to find a policy with id: %d", policyID)
	}
	if p.Enabled == 0 {
		t.Errorf("The Enabled value of replication policy is still 0 after disabled, id: %d", policyID)
	}
}

func TestAddRepPolicy2(t *testing.T) {
	policy2 := models.RepPolicy{
		ProjectID:   3,
		Enabled:     0,
		TargetID:    3,
		Description: "whatever",
		Name:        "mypolicy",
	}
	policyID2, err := AddRepPolicy(policy2)
	t.Logf("added policy, id: %d", policyID2)
	if err != nil {
		t.Errorf("Error occurred in AddRepPolicy: %v", err)
	}
	p, err := GetRepPolicy(policyID2)
	if err != nil {
		t.Errorf("Error occurred in GetPolicy: %v, id: %d", err, policyID2)
	}
	if p == nil {
		t.Errorf("Unable to find a policy with id: %d", policyID2)
	}
	var tm time.Time
	if p.StartTime.After(tm) {
		t.Errorf("Unexpected start_time: %v", p.StartTime)
	}
}

func TestAddRepJob(t *testing.T) {
	job := models.RepJob{
		Repository: "library/ubuntu",
		PolicyID:   policyID,
		Operation:  "transfer",
		TagList:    []string{"12.01", "14.04", "latest"},
	}
	id, err := AddRepJob(job)
	if err != nil {
		t.Errorf("Error occurred in AddRepJob: %v", err)
		return
	}
	jobID = id

	j, err := GetRepJob(id)
	if err != nil {
		t.Errorf("Error occurred in GetRepJob: %v, id: %d", err, id)
		return
	}
	if j == nil {
		t.Errorf("Unable to find a job with id: %d", id)
		return
	}
	if j.Status != models.JobPending || j.Repository != "library/ubuntu" || j.PolicyID != policyID || j.Operation != "transfer" || len(j.TagList) != 3 {
		t.Errorf("Expected data of job, id: %d, Status: %s, Repository: library/ubuntu, PolicyID: %d, Operation: transfer, taglist length 3"+
			"but in returned data:, Status: %s, Repository: %s, Operation: %s, PolicyID: %d, TagList: %v", id, models.JobPending, policyID, j.Status, j.Repository, j.Operation, j.PolicyID, j.TagList)
		return
	}
}

func TestUpdateRepJobStatus(t *testing.T) {
	err := UpdateRepJobStatus(jobID, models.JobFinished)
	if err != nil {
		t.Errorf("Error occured in UpdateRepJobStatus, error: %v, id: %d", err, jobID)
		return
	}
	j, err := GetRepJob(jobID)
	if err != nil {
		t.Errorf("Error occurred in GetRepJob: %v, id: %d", err, jobID)
	}
	if j == nil {
		t.Errorf("Unable to find a job with id: %d", jobID)
	}
	if j.Status != models.JobFinished {
		t.Errorf("Job's status: %s, expected: %s, id: %d", j.Status, models.JobFinished, jobID)
	}
	err = UpdateRepJobStatus(jobID, models.JobPending)
	if err != nil {
		t.Errorf("Error occured in UpdateRepJobStatus when update it back to status pending, error: %v, id: %d", err, jobID)
		return
	}
}

func TestGetRepPolicyByProject(t *testing.T) {
	p1, err := GetRepPolicyByProject(99)
	if err != nil {
		t.Errorf("Error occured in GetRepPolicyByProject:%v, project ID: %d", err, 99)
		return
	}
	if len(p1) > 0 {
		t.Errorf("Unexpected length of policy list, expected: 0, in fact: %d, project id: %d", len(p1), 99)
		return
	}

	p2, err := GetRepPolicyByProject(1)
	if err != nil {
		t.Errorf("Error occuered in GetRepPolicyByProject:%v, project ID: %d", err, 2)
		return
	}
	if len(p2) != 1 {
		t.Errorf("Unexpected length of policy list, expected: 1, in fact: %d, project id: %d", len(p2), 1)
		return
	}
	if p2[0].ID != policyID {
		t.Errorf("Unexpecred policy id in result, expected: %d, in fact: %d", policyID, p2[0].ID)
		return
	}
}

func TestGetRepJobByPolicy(t *testing.T) {
	jobs, err := GetRepJobByPolicy(999)
	if err != nil {
		log.Errorf("Error occured in GetRepJobByPolicy: %v, policy ID: %d", err, 999)
		return
	}
	if len(jobs) > 0 {
		log.Errorf("Unexpected length of jobs, expected: 0, in fact: %d", len(jobs))
		return
	}
	jobs, err = GetRepJobByPolicy(policyID)
	if err != nil {
		log.Errorf("Error occured in GetRepJobByPolicy: %v, policy ID: %d", err, policyID)
		return
	}
	if len(jobs) != 1 {
		log.Errorf("Unexpected length of jobs, expected: 1, in fact: %d", len(jobs))
		return
	}
	if jobs[0].ID != jobID {
		log.Errorf("Unexpected job ID in the result, expected: %d, in fact: %d", jobID, jobs[0].ID)
		return
	}
}

func TestGetRepoJobToStop(t *testing.T) {
	jobs := [...]models.RepJob{
		models.RepJob{
			Repository: "library/ubuntu",
			PolicyID:   policyID,
			Operation:  "transfer",
			Status:     models.JobRunning,
		},
		models.RepJob{
			Repository: "library/ubuntu",
			PolicyID:   policyID,
			Operation:  "transfer",
			Status:     models.JobFinished,
		},
		models.RepJob{
			Repository: "library/ubuntu",
			PolicyID:   policyID,
			Operation:  "transfer",
			Status:     models.JobCanceled,
		},
	}
	var err error
	for _, j := range jobs {
		_, err = AddRepJob(j)
		if err != nil {
			log.Errorf("Failed to add Job: %+v, error: %v", j, err)
			return
		}
	}
	res, err := GetRepJobToStop(policyID)
	if err != nil {
		log.Errorf("Failed to Get Jobs, error: %v", err)
		return
	}
	//time.Sleep(15 * time.Second)
	if len(res) != 2 {
		log.Errorf("Expected length of stoppable jobs, expected:2, in fact: %d", len(res))
		return
	}
}

func TestDeleteRepTarget(t *testing.T) {
	err := DeleteRepTarget(targetID)
	if err != nil {
		t.Errorf("Error occured in DeleteRepTarget: %v, id: %d", err, targetID)
		return
	}
	t.Logf("deleted target, id: %d", targetID)
	tgt, err := GetRepTarget(targetID)
	if err != nil {
		t.Errorf("Error occurred in GetTarget: %v, id: %d", err, targetID)
	}
	if tgt != nil {
		t.Errorf("Able to find target after deletion, id: %d", targetID)
	}
}

func TestDeleteRepPolicy(t *testing.T) {
	err := DeleteRepPolicy(policyID)
	if err != nil {
		t.Errorf("Error occured in DeleteRepPolicy: %v, id: %d", err, policyID)
		return
	}
	t.Logf("delete rep policy, id: %d", policyID)
	p, err := GetRepPolicy(policyID)
	if err != nil {
		t.Errorf("Error occured in GetRepPolicy:%v", err)
	}
	if p != nil {
		t.Errorf("Able to find rep policy after deletion, id: %d", policyID)
	}
}

func TestDeleteRepJob(t *testing.T) {
	err := DeleteRepJob(jobID)
	if err != nil {
		t.Errorf("Error occured in DeleteRepJob: %v, id: %d", err, jobID)
		return
	}
	t.Logf("deleted rep job, id: %d", jobID)
	j, err := GetRepJob(jobID)
	if err != nil {
		t.Errorf("Error occured in GetRepJob:%v", err)
	}
	if j != nil {
		t.Errorf("Able to find rep job after deletion, id: %d", jobID)
	}
}

func TestGetOrmer(t *testing.T) {
	o := GetOrmer()
	if o == nil {
		t.Errorf("Error get ormer.")
	}
}
