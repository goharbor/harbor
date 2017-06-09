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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
)

func TestLogGet(t *testing.T) {
	fmt.Println("Testing Log API")
	apiTest := newHarborAPI()
	assert := assert.New(t)

	CommonAddUser()

	statusCode, result, err := apiTest.LogGet(*testUser)
	assert.Nil(err)
	assert.Equal(200, statusCode)

	logNum := len(result)

	fmt.Println("add the project first.")
	project := apilib.ProjectReq{
		ProjectName: "project_for_test_log",
		Public:      1,
	}

	reply, err := apiTest.ProjectsPost(*testUser, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), reply, "Case 2: Project creation status should be 201")
	}
	//case 1: right parameters, expect the right output
	statusCode, result, err = apiTest.LogGet(*testUser)
	if err != nil {
		t.Error("Error while get log information", err.Error())
		t.Log(err)
	} else {
		assert.Equal(logNum+1, len(result), "lines of logs should be equal")
		num, index := getLog(result)
		if num != 1 {
			assert.Equal(1, num, "add my_project log number should be 1")
		} else {
			assert.Equal("project_for_test_log/", result[index].RepoName)
			assert.Equal("N/A", result[index].RepoTag, "RepoTag should be equal")
			assert.Equal("create", result[index].Operation, "Operation should be equal")
		}
	}
	fmt.Println("log ", result)

	//get the project
	var projects []apilib.Project
	var addProjectID int32
	httpStatusCode, projects, err := apiTest.ProjectsGet(
		&apilib.ProjectQuery{
			Name:   project.ProjectName,
			Owner:  testUser.Name,
			Public: true,
		})
	if err != nil {
		t.Error("Error while search project by proName and isPublic", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		addProjectID = projects[0].ProjectId
	}
	t.Logf("%%%%%%%%%%%%% %v", projects)

	//delete the project
	projectID := strconv.Itoa(int(addProjectID))
	httpStatusCode, err = apiTest.ProjectsDelete(*testUser, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "Case 1: Project creation status should be 200")
		//t.Log(result)
	}
	CommonDelUser()
	fmt.Printf("\n")
}

func getLog(result []apilib.AccessLog) (int, int) {
	var num, index int
	for i := 0; i < len(result); i++ {
		if result[i].RepoName == "project_for_test_log/" {
			num++
			index = i
		}
	}
	return num, index
}
