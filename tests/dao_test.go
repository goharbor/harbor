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
	"flag"
	"fmt"
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

	//Create a custom flag set, let user to provide DB related configures for testing.
	fs := flag.NewFlagSet("DB related configures", 0)

	dbIp := fs.String("db_ip", "localhost", "IP address for connecting a test DB.")
	dbPort := fs.String("db_port", "3306", "Port number for connecting a test DB.")
	dbUser := fs.String("db_user", "root", "Username for logging in a test DB.")
	dbPassword := fs.String("db_password", "root", "Password for logging in a test DB.")

	fs.Parse([]string{"db_ip", "db_port", "db_user", "db_password"})

	if fs.NFlag() == 0 {
		fs.PrintDefaults()
		fmt.Println("Now, use DEFAULT values if omit to set all of flags.")
	}

	if fs.Parsed() {

		clearUp(USERNAME)

		os.Setenv("MYSQL_PORT_3306_TCP_ADDR", *dbIp)
		os.Setenv("MYSQL_PORT_3306_TCP_PORT", *dbPort)
		os.Setenv("MYSQL_USR", *dbUser)
		os.Setenv("MYSQL_PWD", *dbPassword)
		os.Setenv("AUTH_MODE", "db_auth")
		os.Exit(m.Run())
	}
}

func ExampleRegister() {

	user := models.User{
		Username: USERNAME,
		Email:    "tester01@vmware.com",
		Password: "Abc12345",
		Realname: "tester01",
		Comment:  "register",
	}

	_, err := dao.Register(user)
	if err != nil {
		log.Printf("Error occurred in Register: %v", err)
	}

	//Check if user registered successfully.
	queryUser := models.User{
		Username: USERNAME,
	}
	newUser, err := dao.GetUser(queryUser)
	if err != nil {
		log.Fatalf("Error occurred in GetUser: %v", err)
	}

	fmt.Println(newUser.Username)
	fmt.Println(newUser.Email)
	// Output:
	// Tester01
	// tester01@vmware.com

}

func ExampleUserExists() {
	var exists bool
	var err error

	exists, err = dao.UserExists(models.User{Username: "Tester01"}, "username")
	fmt.Println(exists)

	if err != nil {
		log.Fatalf("Error occurred in UserExists: %v", err)
	}
	exists, err = dao.UserExists(models.User{Email: "tester01@vmware.com"}, "email")
	fmt.Println(exists)

	if err != nil {
		log.Fatalf("Error occurred in UserExists: %v", err)
	}

	//Output:
	//true
	//true

}

func ExampleLoginByUserName() {

	userQuery := models.User{
		Username: USERNAME,
		Password: "Abc12345",
	}

	loginUser, err := dao.LoginByDb(models.AuthModel{userQuery.Username, userQuery.Password})
	if err != nil {
		log.Fatalf("Error occurred in LoginByDb: %v", err)
	}
	if loginUser == nil {
		log.Fatalf("No found for user logined by username and password: %v", userQuery)
	}

	fmt.Println(loginUser.Username)
	// Output:
	// Tester01
}

func ExampleLoginByEmail() {

	userQuery := models.User{
		Email:    "tester01@vmware.com",
		Password: "Abc12345",
	}

	loginUser, err := dao.LoginByDb(models.AuthModel{userQuery.Email, userQuery.Password})
	if err != nil {
		log.Fatalf("Error occurred in LoginByDb: %v", err)
	}
	if loginUser == nil {
		log.Fatalf("No found for user logined by email and password : %v", userQuery)
	}
	fmt.Println(loginUser.Username)
	// Output:
	// Tester01
}

var currentUser *models.User

func ExampleGetUser() {
	queryUser := models.User{
		Username: USERNAME,
	}
	var err error
	currentUser, err = dao.GetUser(queryUser)
	if err != nil {
		log.Fatalf("Error occurred in GetUser", err)
	}
	if currentUser == nil {
		log.Fatalf("No user found queried by username: %v", queryUser)
	}
	fmt.Println(currentUser.Username)
	//Output:
	//Tester01
}

func ExampleListUsers() {
	users, err := dao.ListUsers(models.User{Username: "tester01"})
	if err != nil {
		log.Fatalf("Error occurred in ListUsers: %v", err)
	}

	for _, u := range users {
		fmt.Println(u.Username)
	}
	//Output:
	//Tester01
}

func ExampleResetUserPassword() {
	uuid, err := dao.GenerateRandomString()
	if err != nil {
		log.Fatalf("Error occurred in GenerateRandomString: %v", err)
	}

	err = dao.UpdateUserResetUuid(models.User{ResetUuid: uuid, Email: currentUser.Email})
	if err != nil {
		log.Fatalf("Error occurred in UpdateUserResetUuid: %v", err)
	}

	err = dao.ResetUserPassword(models.User{UserId: currentUser.UserId, Password: "HarborTester12345", ResetUuid: uuid, Salt: currentUser.Salt})
	if err != nil {
		log.Fatalf("Error occurred in ResetUserPassword: %v", err)
	}

	loginedUser, err := dao.LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "HarborTester12345"})
	if err != nil {
		log.Fatalf("Error occurred in LoginByDb: %v", err)
	}

	fmt.Println(loginedUser.Username)
	//Output:
	//Tester01
}

func ExampleChangeUserPassword() {
	err := dao.ChangeUserPassword(models.User{UserId: currentUser.UserId, Password: "NewHarborTester12345", Salt: currentUser.Salt})
	if err != nil {
		log.Fatalf("Error occurred in ChangeUserPassword: %v", err)
	}

	loginedUser, err := dao.LoginByDb(models.AuthModel{Principal: currentUser.Username, Password: "NewHarborTester12345"})
	if err != nil {
		log.Fatalf("Error occurred in LoginByDb: %v", err)
	}

	fmt.Println(loginedUser.Username)
	//Output:
	//Tester01
}

func ExampleQueryRelevantProjectsWhenNoProjectAdded() {
	projects, err := dao.QueryRelevantProjects(currentUser.UserId)
	if err != nil {
		log.Fatalf("Error occurred in QueryRelevantProjects: %v", err)
	}
	fmt.Println(len(projects))
	for _, p := range projects {
		fmt.Println(p.Name)
	}
	//Output:
	//1
	//library
}

func ExampleAddProject() {

	project := models.Project{
		OwnerId:      currentUser.UserId,
		Name:         PROJECT_NAME,
		CreationTime: time.Now(),
		OwnerName:    currentUser.Username,
	}

	err := dao.AddProject(project)
	if err != nil {
		log.Fatalf("Error occurred in AddProject: %v", err)
	}

	newProject, err := dao.GetProjectByName(PROJECT_NAME)
	if err != nil {
		log.Fatalf("Error occurred in GetProjectByName: %v", err)
	}
	if newProject == nil {
		log.Fatalf("No project found queried by project name: %v", PROJECT_NAME)
	}
	fmt.Println(newProject.Name)
	//Output:
	//test_project
}

var currentProject *models.Project

func ExampleGetProject() {
	var err error
	currentProject, err = dao.GetProjectByName(PROJECT_NAME)
	if err != nil {
		log.Fatalf("Error occurred in GetProjectByName: %v", err)
	}
	if currentProject == nil {
		log.Fatalf("No project found queried by project name: %v", PROJECT_NAME)
	}
	fmt.Println(currentProject.Name)
	//Output:
	//test_project
}

func getProjectRole(projectId int64) []models.Role {
	o := orm.NewOrm()
	var r []models.Role
	_, err := o.Raw(`select r.role_id, r.name
		from project_role pr 
		 left join role r on pr.role_id = r.role_id
		where project_id = ?`, projectId).QueryRows(&r)
	if err != nil {
		log.Fatalf("Error occurred in querying project_role: %v", err)
	}
	return r
}

func ExampleCheckProjectRoles() {
	r := getProjectRole(currentProject.ProjectId)
	fmt.Println(len(r))

	for _, pr := range r {
		fmt.Println(pr.RoleId, pr.Name)
	}

	//Output: 3
	//2 projectAdmin
	//3 developer
	//4 guest

}

func ExampleGetAccessLog() {
	queryAccessLog := models.AccessLog{
		UserId:    currentUser.UserId,
		ProjectId: currentProject.ProjectId,
	}
	accessLogs, err := dao.GetAccessLogs(queryAccessLog)
	if err != nil {
		log.Fatalf("Error occurred in GetAccessLog: %v", err)
	}
	fmt.Println(len(accessLogs))
	for _, log := range accessLogs {
		fmt.Println(log.Operation, log.RepoName)
	}
	//Output:
	//1
	//create test_project/
}

func ExampleProjectExists() {
	var exists bool
	var err error
	exists, err = dao.ProjectExists(currentProject.ProjectId)
	fmt.Println(exists)
	if err != nil {
		log.Fatalf("Error occurred in ProjectExists: %v", err)
	}
	exists, err = dao.ProjectExists(currentProject.Name)
	fmt.Println(exists)
	if err != nil {
		log.Fatalf("Error occurred in ProjectExists: %v", err)
	}
	//Output:
	//true
	//true

}

func ExampleToggleProjectPublicity() {
	err := dao.ToggleProjectPublicity(currentProject.ProjectId, PUBLICITY_ON)
	if err != nil {
		log.Fatalf("Error occurred in ToggleProjectPublicity: %v", err)
	}

	currentProject, err = dao.GetProjectByName(PROJECT_NAME)
	if err != nil {
		log.Fatalf("Error occurred in GetProjectByName: %v", err)
	}
	fmt.Println(currentProject.Public)

	err = dao.ToggleProjectPublicity(currentProject.ProjectId, PUBLICITY_OFF)
	if err != nil {
		log.Fatalf("Error occurred in ToggleProjectPublicity: %v", err)
	}

	currentProject, err = dao.GetProjectByName(PROJECT_NAME)
	if err != nil {
		log.Fatalf("Error occurred in GetProjectByName: %v", err)
	}

	fmt.Println(currentProject.Public)
	//Output:
	//1
	//0
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

func ExampleGetUserProjectRole() {
	r := getUserProjectRole(currentProject.ProjectId, currentUser.UserId)

	//Get the size of current user project role.
	fmt.Println(len(r))

	//Iterating current user project role info.
	for _, upr := range r {
		fmt.Println(upr.RoleId, upr.Name)
	}

	//Output:
	//1
	//2 projectAdmin
}

func ExampleProjectPermission() {
	roleCode, err := dao.GetPermission(currentUser.Username, currentProject.Name)
	if err != nil {
		log.Fatalf("Error occurred in GetPermission: %v", err)
	}
	fmt.Println(roleCode)
	//Output:
	//MDRWS
}

func ExampleQueryRelevantProjects() {
	projects, err := dao.QueryRelevantProjects(currentUser.UserId)
	if err != nil {
		log.Fatalf("Error occurred in QueryRelevantProjects: %v", err)
	}
	fmt.Println(len(projects))
	for _, p := range projects {
		fmt.Println(p.Name)
	}
	//Output:
	//2
	//library
	//test_project
}

func ExampleAssignUserProjectRole() {
	err := dao.AddUserProjectRole(currentUser.UserId, currentProject.ProjectId, DEVELOPER)
	if err != nil {
		log.Fatalf("Error occurred in AddUserProjectRole: %v", err)
	}

	r := getUserProjectRole(currentProject.ProjectId, currentUser.UserId)

	//Get the size of current user project role info.
	fmt.Println(len(r))

	//Iterating current user project role.
	for _, upr := range r {
		fmt.Println(upr.RoleId, upr.Name)
	}

	//Output:
	//2
	//2 projectAdmin
	//3 developer
}

func ExampleDeleteUserProjectRole() {
	err := dao.DeleteUserProjectRoles(currentUser.UserId, currentProject.ProjectId)
	if err != nil {
		log.Fatalf("Error occurred in DeleteUserProjectRoles: %v", err)
	}

	r := getUserProjectRole(currentProject.ProjectId, currentUser.UserId)

	//Get the size of current user project role.
	fmt.Println(len(r))

	//Iterating current user project role info.
	for _, upr := range r {
		fmt.Println(upr.RoleId, upr.Name)
	}

	//Output:
	//0
}

func ExampleDeleteUser() {
	err := dao.DeleteUser(currentUser.UserId)
	if err != nil {
		log.Fatalf("Error occurred in DeleteUser: %v", err)
	}
	user, err := dao.GetUser(*currentUser)
	if err != nil {
		log.Fatalf("Error occurred in GetUser: %v", err)
	}
	fmt.Println(user)
	//Output:
	//<nil>
}
