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
	"testing"

	"strconv"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
)

func TestMemGet(t *testing.T) {
	var result []apilib.User
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()
	projectID := "1"

	fmt.Println("Testing Member Get API")
	//-------------------case 1 : response code = 200------------------------//
	httpStatusCode, result, err = apiTest.GetProjectMembersByProID(*admin, projectID)
	if err != nil {
		t.Error("Error whihle get members by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(int(1), result[0].UserId, "User Id should be 1")
		assert.Equal("admin", result[0].Username, "User name should be admin")
	}

	//---------case 2: Response Code=401,User need to log in first.----------//
	fmt.Println("case 2: Response Code=401,User need to log in first.")
	httpStatusCode, result, err = apiTest.GetProjectMembersByProID(*unknownUsr, projectID)
	if err != nil {
		t.Error("Error while get members by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "Case 2: Project creation status should be 401")
	}

	//------------case 3: Response Code=404,Project does not exist-----------//
	fmt.Println("case 3: Response Code=404,Project does not exist")
	projectID = "11"
	httpStatusCode, result, err = apiTest.GetProjectMembersByProID(*admin, projectID)
	if err != nil {
		t.Error("Error while get members by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "Case 3: Project creation status should be 404")
	}

	//------------case 4: Response Code=404, member does not exist-----------//
	fmt.Println("case 4: Response Code=404, member does not exist")
	projectID = "1"
	memberID := "10000"
	httpStatusCode, err = apiTest.GetMemByPIDUID(*admin, projectID, memberID)
	if err != nil {
		t.Fatalf("failed to get member %s of project %s: %v", memberID, projectID, err)
	}

	assert.Equal(int(404), httpStatusCode,
		fmt.Sprintf("response status code should be 404 other than %d", httpStatusCode))

	fmt.Printf("\n")
}

/**
 * Add project role member accompany with  projectID
 * role_id = 1 : ProjectAdmin
 * role_id = 2 : Developer
 * role_id = 3 : Guest
 */

func TestMemPost(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()
	projectID := "1"
	CommonAddUser()
	roles := &apilib.RoleParam{[]int32{1}, TestUserName}
	fmt.Printf("Add User \"%s\" successfully!\n", TestUserName)

	fmt.Println("Testing Member Post API")
	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1: response code = 200")
	httpStatusCode, err = apiTest.AddProjectMember(*admin, projectID, *roles)
	if err != nil {
		t.Error("Error whihle add project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//---------case 2: Response Code=409,User is ready in project.----------//
	fmt.Println("case 2: Response Code=409,User is ready in project.")
	httpStatusCode, err = apiTest.AddProjectMember(*admin, projectID, *roles)
	if err != nil {
		t.Error("Error while add project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), httpStatusCode, "Case 2: httpStatusCode  should be 409")
	}

	//---------case 3: Response Code=404,User does not exist.----------//
	fmt.Println("case 3: Response Code=404,User does not exist.")

	errorRoles := &apilib.RoleParam{[]int32{1}, "T"}
	httpStatusCode, err = apiTest.AddProjectMember(*admin, projectID, *errorRoles)
	if err != nil {
		t.Error("Error while add project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "Case 3: httpStatusCode status should be 404")
	}
	/*
		//---------case 4: Response Code=403,User in session does not have permission to the project..----------//
		fmt.Println("case 4:User in session does not have permission to the project.")

		httpStatusCode, err = apiTest.AddProjectMember(*testUser, projectID, *roles)
		if err != nil {
			t.Error("Error while add project role member", err.Error())
			t.Log(err)
		} else {
			assert.Equal(int(403), httpStatusCode, "Case 3: httpStatusCode status should be 403")
		}

	*/
}

func TestGetMemByPIDUID(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()
	projectID := "1"
	userID := strconv.Itoa(CommonGetUserID())
	fmt.Println("Testing Member Get API by PID and UID")
	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1: response code = 200")
	httpStatusCode, err = apiTest.GetMemByPIDUID(*admin, projectID, userID)
	if err != nil {
		t.Error("Error whihle get project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

}

func TestPutMem(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()
	projectID := "1"
	userID := strconv.Itoa(CommonGetUserID())

	roles := &apilib.RoleParam{[]int32{3}, TestUserName}
	fmt.Println("Testing Member Put API")
	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1: response code = 200")
	httpStatusCode, err = apiTest.PutProjectMember(*admin, projectID, userID, *roles)
	if err != nil {
		t.Error("Error whihle put project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

}

func TestDeleteMemUser(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()
	projectID := "1"

	fmt.Println("Testing Member Delete API")
	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1: response code = 200")

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
