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
package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
	"strconv"
	"testing"
	"time"
)

var addProject *apilib.ProjectReq
var addPID int

func InitAddPro() {
	addProject = &apilib.ProjectReq{"add_project", 1}
}

func TestAddProject(t *testing.T) {

	fmt.Println("\nTesting Add Project(ProjectsPost) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	//prepare for test
	InitAddPro()

	//case 1: admin not login, expect project creation fail.

	result, err := apiTest.ProjectsPost(*unknownUsr, *addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), result, "Case 1: Project creation status should be 401")
		//t.Log(result)
	}

	//case 2: admin successful login, expect project creation success.
	fmt.Println("case 2: admin successful login, expect project creation success.")

	result, err = apiTest.ProjectsPost(*admin, *addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), result, "Case 2: Project creation status should be 201")
		//t.Log(result)
	}

	//case 3: duplicate project name, create project fail
	fmt.Println("case 3: duplicate project name, create project fail")

	result, err = apiTest.ProjectsPost(*admin, *addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), result, "Case 3: Project creation status should be 409")
		//t.Log(result)
	}

	//case 4: reponse code = 400 : Project name is illegal in length
	fmt.Println("case 4 : reponse code = 400 : Project name is illegal in length ")

	result, err = apiTest.ProjectsPost(*admin, apilib.ProjectReq{"t", 1})
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), result, "case 4 : reponse code = 400 : Project name is illegal in length ")
	}

	fmt.Printf("\n")

}

//Get project by proName
func TestProGetByName(t *testing.T) {
	fmt.Println("\nTest for Project GET API by project name")
	assert := assert.New(t)

	apiTest := newHarborAPI()
	var result []apilib.Project

	//----------------------------case 1 : Response Code=200----------------------------//
	fmt.Println("case 1: respose code:200")
	httpStatusCode, result, err := apiTest.ProjectsGet(addProject.ProjectName, 1)
	if err != nil {
		t.Error("Error while search project by proName and isPublic", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(addProject.ProjectName, result[0].ProjectName, "Project name is wrong")
		assert.Equal(int32(1), result[0].Public, "Public is wrong")
		//find add projectID
		addPID = int(result[0].ProjectId)
	}
	//----------------------------case 2 : Response Code=401:is_public=0----------------------------//
	fmt.Println("case 2: respose code:401,isPublic = 0")
	httpStatusCode, result, err = apiTest.ProjectsGet("library", 0)
	if err != nil {
		t.Error("Error while search project by proName and isPublic", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 200")
	}

	//-------------------case 3 :  check admin project role------------------------//
	httpStatusCode, result, err = apiTest.ProjectsGet(addProject.ProjectName, 0, *admin)
	if err != nil {
		t.Error("Error while search project by proName and isPublic", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(addProject.ProjectName, result[0].ProjectName, "Project name is wrong")
		assert.Equal(int32(1), result[0].Public, "Public is wrong")
		assert.Equal(int32(1), result[0].CurrentUserRoleId, "User project role is wrong")
	}

	//-------------------case 4 : add project member and check his role ------------------------//
	CommonAddUser()
	roles := &apilib.RoleParam{[]int32{2}, TestUserName}
	projectID := strconv.Itoa(addPID)
	httpStatusCode, err = apiTest.AddProjectMember(*admin, projectID, *roles)
	if err != nil {
		t.Error("Error whihle add project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
	httpStatusCode, result, err = apiTest.ProjectsGet(addProject.ProjectName, 0, *testUser)
	if err != nil {
		t.Error("Error while search project by proName and isPublic", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(addProject.ProjectName, result[0].ProjectName, "Project name is wrong")
		assert.Equal(int32(1), result[0].Public, "Public is wrong")
		assert.Equal(int32(2), result[0].CurrentUserRoleId, "User project role is wrong")
	}
	id := strconv.Itoa(CommonGetUserID())
	httpStatusCode, err = apiTest.DeleteProjectMember(*admin, projectID, id)
	if err != nil {
		t.Error("Error whihle add project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
	CommonDelUser()
}

//Get project by proID
func TestProGetByID(t *testing.T) {
	fmt.Println("\nTest for Project GET API by project id")
	assert := assert.New(t)

	apiTest := newHarborAPI()
	var result apilib.Project
	projectID := strconv.Itoa(addPID)

	//----------------------------case 1 : Response Code=200----------------------------//
	fmt.Println("case 1: respose code:200")
	httpStatusCode, result, err := apiTest.ProjectsGetByPID(projectID)
	if err != nil {
		t.Error("Error while search project by proID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(addProject.ProjectName, result.ProjectName, "ProjectName is wrong")
		assert.Equal(int32(1), result.Public, "Public is wrong")
	}
	fmt.Printf("\n")
}
func TestDeleteProject(t *testing.T) {

	fmt.Println("\nTesting Delete Project(ProjectsPost) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	projectID := strconv.Itoa(addPID)

	//--------------------------case 1: Response Code=401,User need to log in first.-----------------------//
	fmt.Println("case 1: Response Code=401,User need to log in first.")
	httpStatusCode, err := apiTest.ProjectsDelete(*unknownUsr, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "Case 1: Project creation status should be 401")
	}

	//--------------------------case 2: Response Code=200---------------------------------//
	fmt.Println("case2: respose code:200")
	httpStatusCode, err = apiTest.ProjectsDelete(*admin, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "Case 2: Project creation status should be 200")
	}

	//--------------------------case 3: Response Code=404,Project does not exist---------------------------------//
	fmt.Println("case 3: Response Code=404,Project does not exist")
	projectID = "11"
	httpStatusCode, err = apiTest.ProjectsDelete(*admin, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "Case 3: Project creation status should be 404")
	}

	//--------------------------case 4: Response Code=400,Invalid project id.---------------------------------//
	fmt.Println("case 4: Response Code=400,Invalid project id.")
	projectID = "cc"
	httpStatusCode, err = apiTest.ProjectsDelete(*admin, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "Case 4: Project creation status should be 400")
	}
	fmt.Printf("\n")

}
func TestProHead(t *testing.T) {
	fmt.Println("\nTest for Project HEAD API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	//----------------------------case 1 : Response Code=200----------------------------//
	fmt.Println("case 1: respose code:200")
	httpStatusCode, err := apiTest.ProjectsHead(*admin, "library")
	if err != nil {
		t.Error("Error while search project by proName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//----------------------------case 2 : Response Code=404:Project name does not exist.----------------------------//
	fmt.Println("case 2: respose code:404,Project name does not exist.")
	httpStatusCode, err = apiTest.ProjectsHead(*admin, "libra")
	if err != nil {
		t.Error("Error while search project by proName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}
	//----------------------------case 3 : Response Code=401:User need to log in first..----------------------------//
	fmt.Println("case 3: respose code:401,User need to log in first..")
	httpStatusCode, err = apiTest.ProjectsHead(*unknownUsr, "libra")
	if err != nil {
		t.Error("Error while search project by proName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 401")
	}

	fmt.Printf("\n")

}

func TestToggleProjectPublicity(t *testing.T) {
	fmt.Println("\nTest for Project PUT API: Update properties for a selected project")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	//-------------------case1: Response Code=200------------------------------//
	fmt.Println("case 1: respose code:200")
	httpStatusCode, err := apiTest.ToggleProjectPublicity(*admin, "1", 1)
	if err != nil {
		t.Error("Error while search project by proId", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
	//-------------------case2: Response Code=401 User need to log in first. ------------------------------//
	fmt.Println("case 2: respose code:401, User need to log in first.")
	httpStatusCode, err = apiTest.ToggleProjectPublicity(*unknownUsr, "1", 1)
	if err != nil {
		t.Error("Error while search project by proId", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 401")
	}
	//-------------------case3: Response Code=400 Invalid project id------------------------------//
	fmt.Println("case 3: respose code:400, Invalid project id")
	httpStatusCode, err = apiTest.ToggleProjectPublicity(*admin, "cc", 1)
	if err != nil {
		t.Error("Error while search project by proId", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}
	//-------------------case4: Response Code=404 Not found the project------------------------------//
	fmt.Println("case 4: respose code:404, Not found the project")
	httpStatusCode, err = apiTest.ToggleProjectPublicity(*admin, "0", 1)
	if err != nil {
		t.Error("Error while search project by proId", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	fmt.Printf("\n")
}
func TestProjectLogsFilter(t *testing.T) {
	fmt.Println("\nTest for search access logs filtered by operations and date time ranges..")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	endTimestamp := time.Now().Unix()
	startTimestamp := endTimestamp - 3600
	accessLog := &apilib.AccessLogFilter{
		Username:       "admin",
		Keywords:       "",
		BeginTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	}

	//-------------------case1: Response Code=200------------------------------//
	fmt.Println("case 1: respose code:200")
	projectID := "1"
	httpStatusCode, _, err := apiTest.ProjectLogsFilter(*admin, projectID, *accessLog)
	if err != nil {
		t.Error("Error while search access logs")
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
	//-------------------case2: Response Code=401:User need to log in first.------------------------------//
	fmt.Println("case 2: respose code:401:User need to log in first.")
	projectID = "1"
	httpStatusCode, _, err = apiTest.ProjectLogsFilter(*unknownUsr, projectID, *accessLog)
	if err != nil {
		t.Error("Error while search access logs")
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 401")
	}
	//-------------------case3: Response Code=404:Project does not exist.-------------------------//
	fmt.Println("case 3: respose code:404:Illegal format of provided ID value.")
	projectID = "11111"
	httpStatusCode, _, err = apiTest.ProjectLogsFilter(*admin, projectID, *accessLog)
	if err != nil {
		t.Error("Error while search access logs")
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}
	fmt.Printf("\n")
}
