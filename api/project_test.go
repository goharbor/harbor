package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
	"strconv"
	"testing"
	"time"
)

var admin, unknownUsr *usrInfo
var addProject apilib.Project

func Init() {
	admin = &usrInfo{"admin", "Harbor12345"}
	unknownUsr = &usrInfo{"unknown", "unknown"}
	addProject.ProjectName = "test_project"
	addProject.Public = 1

}

func TestAddProject(t *testing.T) {

	fmt.Println("\nTesting Add Project(ProjectsPost) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	//prepare for test
	Init()

	//case 1: admin not login, expect project creation fail.

	result, err := apiTest.ProjectsPost(*unknownUsr, addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), result, "Case 1: Project creation status should be 401")
		//t.Log(result)
	}

	//case 2: admin successful login, expect project creation success.
	fmt.Println("case 2: admin successful login, expect project creation success.")

	unknownUsr = admin

	result, err = apiTest.ProjectsPost(*admin, addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), result, "Case 2: Project creation status should be 201")
		//t.Log(result)
	}

	//case 3: duplicate project name, create project fail
	fmt.Println("case 3: duplicate project name, create project fail")

	result, err = apiTest.ProjectsPost(*admin, addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), result, "Case 3: Project creation status should be 409")
		//t.Log(result)
	}
	fmt.Printf("\n")

}

func TestProGet(t *testing.T) {
	fmt.Println("\nTest for Project GET API")
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
		//find add projectID
		addProject.ProjectId = int32(result[0].ProjectId)
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

	fmt.Printf("\n")
}

func TestDeleteProject(t *testing.T) {

	fmt.Println("\nTesting Delete Project(ProjectsPost) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	projectID := strconv.Itoa(int(addProject.ProjectId))
	//--------------------------case 1: Response Code=200---------------------------------//

	httpStatusCode, err := apiTest.ProjectsDelete(*admin, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "Case 1: Project creation status should be 200")
		//t.Log(result)
	}

	fmt.Printf("\n")

}
func TestProHead(t *testing.T) {
	Init()
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
