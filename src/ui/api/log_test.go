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

func TestLogGet(t *testing.T) {

	fmt.Println("Testing Log API")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	//prepare for test

	var project apilib.ProjectReq
	project.ProjectName = "my_project"
	project.Public = 1
	now := fmt.Sprintf("%v", time.Now().Unix())
	statusCode, result, err := apiTest.LogGet(*admin, "0", now, "1000")
	if err != nil {
		t.Error("Error while get log information", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), statusCode, "Log get should return 200")

	}
	logNum := len(result)
	fmt.Println("result", result)
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
	now = fmt.Sprintf("%v", time.Now().Unix())
	statusCode, result, err = apiTest.LogGet(*admin, "0", now, "1000")
	if err != nil {
		t.Error("Error while get log information", err.Error())
		t.Log(err)
	} else {
		assert.Equal(logNum+1, len(result), "lines of logs should be equal")
		num, index := getLog(result)
		if num != 1 {
			assert.Equal(1, num, "add my_project log number should be 1")
		} else {
			assert.Equal("my_project/", result[index].RepoName, "RepoName should be equal")
			assert.Equal("N/A", result[index].RepoTag, "RepoTag should be equal")
			assert.Equal("create", result[index].Operation, "Operation should be equal")
		}
	}
	fmt.Println("log ", result)
	//case 2: wrong format of start_time parameter, expect the wrong output
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
	statusCode, result, err = apiTest.LogGet(*admin, "0", now, "s")
	if err != nil {
		t.Error("Error occured while get log information since the format of lines parameter is not right.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), statusCode, "Http status code should be 400")
	}

	//case 5: wrong format of lines parameter, expect the wrong output
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
		//default get 10 logs
		if logNum+1 >= 10 {
			logNum = 10
		} else {
			logNum++
		}
		assert.Equal(logNum, len(result), "lines of logs should be equal")
		num, index := getLog(result)
		if num != 1 {
			assert.Equal(1, num, "add my_project log number should be 1")
		} else {
			assert.Equal("my_project/", result[index].RepoName, "RepoName should be equal")
			assert.Equal("N/A", result[index].RepoTag, "RepoTag should be equal")
			assert.Equal("create", result[index].Operation, "Operation should be equal")
		}
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

func getLog(result []apilib.AccessLog) (int, int) {
	var num, index int
	for i := 0; i < len(result); i++ {
		if result[i].RepoName == "my_project/" {
			num++
			index = i
		}
	}
	return num, index
}
