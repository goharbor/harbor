package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
	"strconv"
	"testing"
	"time"
)

func TestLogGet(t *testing.T) {

	fmt.Println("Testing Log API")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	//prepare for test

	var project apilib.ProjectReq
	project.ProjectName = "my_project"
	project.Public = 1

	//add the project first.
	fmt.Println("add the project first.")
	reply, err := apiTest.ProjectsPost(*admin, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), reply, "Case 2: Project creation status should be 201")
	}
	//case 1: right parameters, expect the right output
	now := fmt.Sprintf("%v", time.Now().Unix())
	statusCode, result, err := apiTest.LogGet(*admin, "0", now, "3")
	if err != nil {
		t.Error("Error while get log information", err.Error())
		t.Log(err)
	} else {
		assert.Equal(1, len(result), "lines of logs should be equal")
		assert.Equal(int32(1), result[0].LogId, "LogId should be equal")
		assert.Equal("my_project/", result[0].RepoName, "RepoName should be equal")
		assert.Equal("N/A", result[0].RepoTag, "RepoTag should be equal")
		assert.Equal("create", result[0].Operation, "Operation should be equal")
	}

	//case 2: wrong format of start_time parameter, expect the wrong output
	now = fmt.Sprintf("%v", time.Now().Unix())
	statusCode, result, err = apiTest.LogGet(*admin, "ss", now, "3")
	if err != nil {
		t.Error("Error occured while get log information since the format of start_time parameter is not right.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), statusCode, "Http status code should be 400")
	}

	//case 3: wrong format of end_time parameter, expect the wrong output
	statusCode, result, err = apiTest.LogGet(*admin, "0", "cc", "3")
	if err != nil {
		t.Error("Error occured while get log information since the format of end_time parameter is not right.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), statusCode, "Http status code should be 400")
	}

	//case 4: wrong format of lines parameter, expect the wrong output
	now = fmt.Sprintf("%v", time.Now().Unix())
	statusCode, result, err = apiTest.LogGet(*admin, "0", now, "s")
	if err != nil {
		t.Error("Error occured while get log information since the format of lines parameter is not right.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), statusCode, "Http status code should be 400")
	}

	//case 5: wrong format of lines parameter, expect the wrong output
	now = fmt.Sprintf("%v", time.Now().Unix())
	statusCode, result, err = apiTest.LogGet(*admin, "0", now, "-5")
	if err != nil {
		t.Error("Error occured while get log information since the format of lines parameter is not right.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), statusCode, "Http status code should be 400")
	}

	//case 6: all parameters are null, expect the right output
	statusCode, result, err = apiTest.LogGet(*admin, "", "", "")
	if err != nil {
		t.Error("Error while get log information", err.Error())
		t.Log(err)
	} else {
		assert.Equal(1, len(result), "lines of logs should be equal")
		assert.Equal(int32(1), result[0].LogId, "LogId should be equal")
		assert.Equal("my_project/", result[0].RepoName, "RepoName should be equal")
		assert.Equal("N/A", result[0].RepoTag, "RepoTag should be equal")
		assert.Equal("create", result[0].Operation, "Operation should be equal")
	}

	//get the project
	var projects []apilib.Project
	var addProjectID int32
	httpStatusCode, projects, err := apiTest.ProjectsGet(project.ProjectName, 1)
	if err != nil {
		t.Error("Error while search project by proName and isPublic", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		addProjectID = projects[0].ProjectId
	}

	//delete the project
	projectID := strconv.Itoa(int(addProjectID))
	httpStatusCode, err = apiTest.ProjectsDelete(*admin, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "Case 1: Project creation status should be 200")
		//t.Log(result)
	}

	fmt.Printf("\n")

}
