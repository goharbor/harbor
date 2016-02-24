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
package test

import (
	//	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego/orm"
)

func execUpdate(o orm.Ormer, sql string, params interface{}) error {
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(params)
	if err != nil {
		return err
	}
	return nil
}

func clearUp(username string) {
	o := orm.NewOrm()
	o.Begin()
	err := execUpdate(o, `delete upr from user_project_role upr
	  left join project_role pr on upr.pr_id = pr.pr_id
	  left join project p on pr.project_id = p.project_id
	  left join user u on u.user_id = p.owner_id
	 where u.username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Println(err)
	}
	err = execUpdate(o, `delete pr from project_role pr
	  left join project p on pr.project_id = p.project_id
	  left join user u on u.user_id = p.owner_id
	 where u.username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Println(err)
	}
	err = execUpdate(o, `delete a from access_log a
	  left join user u on a.user_id = u.user_id
	 where u.username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Println(err)
	}
	err = execUpdate(o, `delete p from project p
	  left join user u on p.owner_id = u.user_id
	 where u.username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Println(err)
	}
	err = execUpdate(o, `delete u from user u
	 where u.username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Println(err)
	}
	o.Commit()
}

const USERNAME string = "Tester01"
const PROJECT_NAME string = "test_project"
const SYS_ADMIN int = 1
const PROJECT_ADMIN int = 2
const DEVELOPER int = 3
const GUEST int = 4

const PUBLICITY_ON = 1
const PUBLICITY_OFF = 0

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
	if len(dbPassword) == 0 {
		log.Fatalf("environment variable DB_PWD is not set")
	}

	os.Setenv("MYSQL_PORT_3306_TCP_ADDR", dbHost)
	os.Setenv("MYSQL_PORT_3306_TCP_PORT", dbPort)
	os.Setenv("MYSQL_USR", dbUser)
	os.Setenv("MYSQL_PWD", dbPassword)
	os.Setenv("AUTH_MODE", "db_auth")
	dao.InitDB()
	clearUp(USERNAME)
	os.Exit(m.Run())

}

func TestRegister(t *testing.T) {

	user := models.User{
		Username: USERNAME,
		Email:    "tester01@vmware.com",
		Password: "Abc12345",
		Realname: "tester01",
		Comment:  "register",
	}

	_, err := dao.Register(user)
	if err != nil {
		t.Errorf("Error occurred in Register: %v", err)
	}

	//Check if user registered successfully.
	queryUser := models.User{
		Username: USERNAME,
	}
	newUser, err := dao.GetUser(queryUser)
	if err != nil {
		t.Errorf("Error occurred in GetUser: %v", err)
	}

	if newUser.Username != USERNAME {
		t.Errorf("Username does not match, expected: %s, actual: %s", USERNAME, newUser.Username)
	}
	if newUser.Email != "tester01@vmware.com" {
		t.Errorf("Email does not match, expected: %s, actual: %s", "tester01@vmware.com", newUser.Email)
	}
}

func TestUserExists(t *testing.T) {
	var exists bool
	var err error

	exists, err = dao.UserExists(models.User{Username: USERNAME}, "username")
	if err != nil {
		t.Errorf("Error occurred in UserExists: %v", err)
	}
	if !exists {
		t.Errorf("User %s was inserted but does not exist", USERNAME)
	}
	exists, err = dao.UserExists(models.User{Email: "tester01@vmware.com"}, "email")

	if err != nil {
		t.Errorf("Error occurred in UserExists: %v", err)
	}
	if !exists {
		t.Errorf("User with email %s inserted but does not exist", "tester01@vmware.com")
	}
	exists, err = dao.UserExists(models.User{Username: "NOTHERE"}, "username")
	if err != nil {
		t.Errorf("Error occurred in UserExists: %v", err)
	}
	if exists {
		t.Errorf("User %s was not inserted but does exist", "NOTHERE")
	}
}

func TestLoginByUserName(t *testing.T) {

	userQuery := models.User{
		Username: USERNAME,
		Password: "Abc12345",
	}

	loginUser, err := dao.LoginByDb(models.AuthModel{userQuery.Username, userQuery.Password})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}
	if loginUser == nil {
		t.Errorf("No found for user logined by username and password: %v", userQuery)
	}

	if loginUser.Username != USERNAME {
		t.Errorf("User's username does not match after login, expected: %s, actual: %s", USERNAME, loginUser.Username)
	}
}

func TestLoginByEmail(t *testing.T) {

	userQuery := models.User{
		Email:    "tester01@vmware.com",
		Password: "Abc12345",
	}

	loginUser, err := dao.LoginByDb(models.AuthModel{userQuery.Email, userQuery.Password})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}
	if loginUser == nil {
		t.Errorf("No found for user logined by email and password : %v", userQuery)
	}
	if loginUser.Username != USERNAME {
		t.Errorf("User's username does not match after login, expected: %s, actual: %s", USERNAME, loginUser.Username)
	}
}

var currentUser *models.User

func TestGetUser(t *testing.T) {
	queryUser := models.User{
		Username: USERNAME,
	}
	var err error
	currentUser, err = dao.GetUser(queryUser)
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
	users, err := dao.ListUsers(models.User{})
	if err != nil {
		t.Errorf("Error occurred in ListUsers: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expect one user in list, but the acutal length is %d, the list: %+v", len(users), users)
	}
	users2, err := dao.ListUsers(models.User{Username: USERNAME})
	if len(users2) != 1 {
		t.Errorf("Expect one user in list, but the acutal length is %d, the list: %+v", len(users), users)
	}
	if users2[0].Username != USERNAME {
		t.Errorf("The username in result list does not match, expected: %s, actual: %s", USERNAME, users2[0].Username)
	}
}

func TestResetUserPassword(t *testing.T) {
	uuid, err := dao.GenerateRandomString()
	if err != nil {
		t.Errorf("Error occurred in GenerateRandomString: %v", err)
	}

	err = dao.UpdateUserResetUuid(models.User{ResetUuid: uuid, Email: currentUser.Email})
	if err != nil {
		t.Errorf("Error occurred in UpdateUserResetUuid: %v", err)
	}

	err = dao.ResetUserPassword(models.User{UserId: currentUser.UserId, Password: "HarborTester12345", ResetUuid: uuid, Salt: currentUser.Salt})
	if err != nil {
		t.Errorf("Error occurred in ResetUserPassword: %v", err)
	}

	loginedUser, err := dao.LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "HarborTester12345"})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}

	if loginedUser.Username != USERNAME {
		t.Errorf("The username returned by Login does not match, expected: %s, acutal: %s", USERNAME, loginedUser.Username)
	}
}

func TestChangeUserPassword(t *testing.T) {
	err := dao.ChangeUserPassword(models.User{UserId: currentUser.UserId, Password: "NewHarborTester12345", Salt: currentUser.Salt})
	if err != nil {
		t.Errorf("Error occurred in ChangeUserPassword: %v", err)
	}

	loginedUser, err := dao.LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "NewHarborTester12345"})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}

	if loginedUser.Username != USERNAME {
		t.Errorf("The username returned by Login does not match, expected: %s, acutal: %s", USERNAME, loginedUser.Username)
	}
}

func TestChangeUserPasswordWithOldPassword(t *testing.T) {
	err := dao.ChangeUserPassword(models.User{UserId: currentUser.UserId, Password: "NewerHarborTester12345", Salt: currentUser.Salt}, "NewHarborTester12345")
	if err != nil {
		t.Errorf("Error occurred in ChangeUserPassword: %v", err)
	}
	loginedUser, err := dao.LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "NewerHarborTester12345"})
	if err != nil {
		t.Errorf("Error occurred in LoginByDb: %v", err)
	}

	if loginedUser.Username != USERNAME {
		t.Errorf("The username returned by Login does not match, expected: %s, acutal: %s", USERNAME, loginedUser.Username)
	}
}

func TestQueryRelevantProjectsWhenNoProjectAdded(t *testing.T) {
	projects, err := dao.QueryRelevantProjects(currentUser.UserId)
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
		OwnerId:      currentUser.UserId,
		Name:         PROJECT_NAME,
		CreationTime: time.Now(),
		OwnerName:    currentUser.Username,
	}

	err := dao.AddProject(project)
	if err != nil {
		t.Errorf("Error occurred in AddProject: %v", err)
	}

	newProject, err := dao.GetProjectByName(PROJECT_NAME)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}
	if newProject == nil {
		t.Errorf("No project found queried by project name: %v", PROJECT_NAME)
	}
}

var currentProject *models.Project

func TestGetProject(t *testing.T) {
	var err error
	currentProject, err = dao.GetProjectByName(PROJECT_NAME)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}
	if currentProject == nil {
		t.Errorf("No project found queried by project name: %v", PROJECT_NAME)
	}
	if currentProject.Name != PROJECT_NAME {
		t.Errorf("Project name does not match, expected: %s, actual: %s", PROJECT_NAME, currentProject.Name)
	}
}

func getProjectRole(projectId int64) []models.Role {
	o := orm.NewOrm()
	var r []models.Role
	_, err := o.Raw(`select r.role_id, r.name
		from project_role pr
		 left join role r on pr.role_id = r.role_id
		where project_id = ?`, projectId).QueryRows(&r)
	if err != nil {
		log.Printf("Error occurred in querying project_role: %v", err)
	}
	return r
}

func TestCheckProjectRoles(t *testing.T) {
	r := getProjectRole(currentProject.ProjectId)
	if len(r) != 3 {
		t.Errorf("The length of project roles is not 3")
	}
	if r[1].RoleId != 3 {
		t.Errorf("The role id does not match, expected: 3, acutal: %d", r[1].RoleId)
	}
	if r[1].Name != "developer" {
		t.Errorf("The name of role id: 3 should be developer, actual:%s", r[1].Name)
	}
}

func TestGetAccessLog(t *testing.T) {
	queryAccessLog := models.AccessLog{
		UserId:    currentUser.UserId,
		ProjectId: currentProject.ProjectId,
	}
	accessLogs, err := dao.GetAccessLogs(queryAccessLog)
	if err != nil {
		t.Errorf("Error occurred in GetAccessLog: %v", err)
	}
	if len(accessLogs) != 1 {
		t.Errorf("The length of accesslog list should be 1, actual: %d", len(accessLogs))
	}
	if accessLogs[0].RepoName != PROJECT_NAME+"/" {
		t.Errorf("The project name does not match, expected: %s, actual: %s", PROJECT_NAME+"/", accessLogs[0].RepoName)
	}
}

func TestProjectExists(t *testing.T) {
	var exists bool
	var err error
	exists, err = dao.ProjectExists(currentProject.ProjectId)
	if err != nil {
		t.Errorf("Error occurred in ProjectExists: %v", err)
	}
	if !exists {
		t.Errorf("The project with id: %d, does not exist", currentProject.ProjectId)
	}
	exists, err = dao.ProjectExists(currentProject.Name)
	if err != nil {
		t.Errorf("Error occurred in ProjectExists: %v", err)
	}
	if !exists {
		t.Errorf("The project with name: %s, does not exist", currentProject.Name)
	}
}

func TestToggleProjectPublicity(t *testing.T) {
	err := dao.ToggleProjectPublicity(currentProject.ProjectId, PUBLICITY_ON)
	if err != nil {
		t.Errorf("Error occurred in ToggleProjectPublicity: %v", err)
	}

	currentProject, err = dao.GetProjectByName(PROJECT_NAME)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}
	if currentProject.Public != PUBLICITY_ON {
		t.Errorf("project, id: %d, its publicity is not on", currentProject.ProjectId)
	}
	err = dao.ToggleProjectPublicity(currentProject.ProjectId, PUBLICITY_OFF)
	if err != nil {
		t.Errorf("Error occurred in ToggleProjectPublicity: %v", err)
	}

	currentProject, err = dao.GetProjectByName(PROJECT_NAME)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}

	if currentProject.Public != PUBLICITY_OFF {
		t.Errorf("project, id: %d, its publicity is not off", currentProject.ProjectId)
	}
}

func getUserProjectRole(projectId int64, userId int) []models.Role {
	o := orm.NewOrm()
	var r []models.Role
	_, err := o.Raw(`select r.role_id, r.name
		from user_project_role upr
		 left join project_role pr on upr.pr_id = pr.pr_id
		 left join role r on r.role_id = pr.role_id
		where pr.project_id = ? and upr.user_id = ?`, projectId, userId).QueryRows(&r)
	if err != nil {
		log.Fatalf("Error occurred in querying user_project_role: %v", err)
	}
	return r
}

func TestGetUserProjectRole(t *testing.T) {
	r := getUserProjectRole(currentProject.ProjectId, currentUser.UserId)

	//Get the size of current user project role.
	if len(r) != 1 {
		t.Errorf("The user, id: %d, should only have one role in project, id: %d, but actual: %d", currentUser.UserId, currentProject.ProjectId, len(r))
	}

	if r[0].Name != "projectAdmin" {
		t.Errorf("the expected rolename is: projectAdmin, actual: %s", r[0].Name)
	}
}

func TestProjectPermission(t *testing.T) {
	roleCode, err := dao.GetPermission(currentUser.Username, currentProject.Name)
	if err != nil {
		t.Errorf("Error occurred in GetPermission: %v", err)
	}
	if roleCode != "MDRWS" {
		t.Errorf("The expected role code is MDRWS,but actual: %s", roleCode)
	}
}

func TestQueryRelevantProjects(t *testing.T) {
	projects, err := dao.QueryRelevantProjects(currentUser.UserId)
	if err != nil {
		t.Errorf("Error occurred in QueryRelevantProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("Expected length of relevant projects is 2, but actual: %d, the projects: %+v", len(projects), projects)
	}
	if projects[1].Name != PROJECT_NAME {
		t.Errorf("Expected project name in the list: %s, actual: %s", PROJECT_NAME, projects[1].Name)
	}
}

func TestAssignUserProjectRole(t *testing.T) {
	err := dao.AddUserProjectRole(currentUser.UserId, currentProject.ProjectId, DEVELOPER)
	if err != nil {
		t.Errorf("Error occurred in AddUserProjectRole: %v", err)
	}

	r := getUserProjectRole(currentProject.ProjectId, currentUser.UserId)

	//Get the size of current user project role info.
	if len(r) != 2 {
		t.Errorf("Expected length of role list is 2, actual: %d", len(r))
	}

	if r[1].RoleId != 3 {
		t.Errorf("Expected role id of the second role in list is 3, actual: %d", r[1].RoleId)
	}
}

func TestDeleteUserProjectRole(t *testing.T) {
	err := dao.DeleteUserProjectRoles(currentUser.UserId, currentProject.ProjectId)
	if err != nil {
		t.Errorf("Error occurred in DeleteUserProjectRoles: %v", err)
	}

	r := getUserProjectRole(currentProject.ProjectId, currentUser.UserId)
	//Get the size of current user project role.
	if len(r) != 0 {
		t.Errorf("Expected role list length is 0, actual: %d, role list: %+v", len(r), r)
	}
}

func TestDeleteUser(t *testing.T) {
	err := dao.DeleteUser(currentUser.UserId)
	if err != nil {
		t.Errorf("Error occurred in DeleteUser: %v", err)
	}
	user, err := dao.GetUser(*currentUser)
	if err != nil {
		t.Errorf("Error occurred in GetUser: %v", err)
	}
	if user != nil {
		t.Error("user is not nil after deletion, user: %+v", user)
	}
}
