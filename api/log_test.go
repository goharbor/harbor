package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
	"testing"
	"time"
)

func TestLogGet(t *testing.T) {

	fmt.Println("Testing Log API")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	//prepare for test

	admin := &usrInfo{"admin", "Harbor12345"}
	var project apilib.Project
	project.ProjectName = "my_project"
	project.Public = true

	//add the project first.
	fmt.Println("add the project first.")
	reply, err := apiTest.ProjectsPost(*admin, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(reply, int(201), "Case 2: Project creation status should be 201")
	}
	//case 1: right parameters, expect the right output
	now := fmt.Sprintf("%v", time.Now().Unix())
	result, err := apiTest.LogGet(*admin, "0", now, "3")
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
	result, err = apiTest.LogGet(*admin, "ss", now, "3")
	if err != nil {
		t.Error("Error occured while get log information since the format of start_time parameter is not right.", err.Error())
		t.Log(err)
	} else {
		t.Error("Supposed to be an error.")
	}
	//case 3: wrong format of end_time parameter, expect the wrong output
	result, err = apiTest.LogGet(*admin, "0", "ss", "3")
	if err != nil {
		t.Error("Error occured while get log information since the format of end_time parameter is not right.", err.Error())
		t.Log(err)
	} else {
		t.Error("Supposed to be an error.")
	}
	//case 4: wrong format of lines parameter, expect the wrong output
	now = fmt.Sprintf("%v", time.Now().Unix())
	result, err = apiTest.LogGet(*admin, "0", now, "s")
	if err != nil {
		t.Error("Error occured while get log information since the format of lines parameter is not right.", err.Error())
		t.Log(err)
	} else {
		t.Error("Supposed to be an error.")
	}
	//case 5: wrong format of lines parameter, expect the wrong output
	now = fmt.Sprintf("%v", time.Now().Unix())
	result, err = apiTest.LogGet(*admin, "0", now, "-5")
	if err != nil {
		t.Error("Error occured while get log information since the format of lines parameter is not right.", err.Error())
		t.Log(err)
	} else {
		t.Error("Supposed to be an error.")
	}
	//case 6: all parameters are null, expect the right output
	result, err = apiTest.LogGet(*admin, "", "", "")
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

}
