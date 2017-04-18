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
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
)

const (
	addTargetName = "testTargets"
)

var addTargetID int

func TestTargetsPost(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	endPoint := os.Getenv("REGISTRY_URL")
	repTargets := &apilib.RepTargetPost{endPoint, addTargetName, adminName, adminPwd}

	fmt.Println("Testing Targets Post API")

	//-------------------case 1 : response code = 201------------------------//
	fmt.Println("case 1 : response code = 201")
	httpStatusCode, body, err := apiTest.AddTargets(*admin, *repTargets)
	if err != nil {
		t.Error("Error whihle add targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), httpStatusCode, "httpStatusCode should be 201")
		t.Log(body)
	}

	//-----------case 2 : response code = 409,name is already used-----------//
	fmt.Println("case 2 : response code = 409,name is already used")
	httpStatusCode, _, err = apiTest.AddTargets(*admin, *repTargets)
	if err != nil {
		t.Error("Error whihle add targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), httpStatusCode, "httpStatusCode should be 409")
	}

	//-----------case 3 : response code = 409,name is already used-----------//
	fmt.Println("case 3 : response code = 409,endPoint is already used")
	repTargets.Username = "errName"
	httpStatusCode, _, err = apiTest.AddTargets(*admin, *repTargets)
	if err != nil {
		t.Error("Error whihle add targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), httpStatusCode, "httpStatusCode should be 409")
	}

	//--------case 4 : response code = 401,User need to log in first.--------//
	fmt.Println("case 4 : response code = 401,User need to log in first.")
	httpStatusCode, _, err = apiTest.AddTargets(*unknownUsr, *repTargets)
	if err != nil {
		t.Error("Error whihle add targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 401")
	}

	fmt.Printf("\n")

}

func TestTargetsGet(t *testing.T) {
	var httpStatusCode int
	var err error
	var reslut []apilib.RepTarget

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Targets Get API")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	httpStatusCode, reslut, err = apiTest.ListTargets(*admin, addTargetName)
	if err != nil {
		t.Error("Error whihle get targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		addTargetID = int(reslut[0].Id)
	}
}

func TestTargetPing(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Targets Ping Post API")
	//case 1
	body := struct {
		Endpoint string `json:"endpoint"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Endpoint: os.Getenv("REGISTRY_URL"),
		Username: adminName,
		Password: adminPwd,
	}
	httpStatusCode, err = apiTest.PingTarget(*admin, body)
	if err != nil {
		t.Error("Error while ping target", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "")
	}

	//case 2
	body.Endpoint = ""
	httpStatusCode, err = apiTest.PingTarget(*admin, body)
	if err != nil {
		t.Error("Error while ping target", err.Error())
	} else {
		assert.Equal(int(400), httpStatusCode, "")
	}
}

func TestTargetPingByID(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Targets Ping Post API")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	id := addTargetID
	httpStatusCode, err = apiTest.PingTargetByID(*admin, id)
	if err != nil {
		t.Error("Error whihle ping target", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//--------------case 2 : response code = 404,target not found------------//
	fmt.Println("case 2 : response code = 404,target not found")

	id = 1111
	httpStatusCode, err = apiTest.PingTargetByID(*admin, id)
	if err != nil {
		t.Error("Error whihle ping target", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}
}

func TestTargetGetByID(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Targets Get API by Id")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	id := strconv.Itoa(addTargetID)
	httpStatusCode, err = apiTest.GetTargetByID(*admin, id)
	if err != nil {
		t.Error("Error whihle get target by id", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//--------------case 2 : response code = 404,target not found------------//
	fmt.Println("case 2 : response code = 404,target not found")
	id = "1111"
	httpStatusCode, err = apiTest.GetTargetByID(*admin, id)
	if err != nil {
		t.Error("Error whihle get target by id", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

}

func TestTargetsPut(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	endPoint := "1.1.1.1"
	updateRepTargets := &apilib.RepTargetPost{endPoint, addTargetName, adminName, adminPwd}
	id := strconv.Itoa(addTargetID)

	fmt.Println("Testing Target Put API")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	httpStatusCode, err = apiTest.PutTargetByID(*admin, id, *updateRepTargets)
	if err != nil {
		t.Error("Error whihle update target", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//--------------case 2 : response code = 404,target not found------------//
	id = "111"
	fmt.Println("case 2 : response code = 404,target not found")
	httpStatusCode, err = apiTest.PutTargetByID(*admin, id, *updateRepTargets)
	if err != nil {
		t.Error("Error whihle update target", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

}
func TestTargetGetPolicies(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Targets Get API to list policies")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	id := strconv.Itoa(addTargetID)
	httpStatusCode, err = apiTest.GetTargetPoliciesByID(*admin, id)
	if err != nil {
		t.Error("Error whihle get target by id", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//--------------case 2 : response code = 404,target not found------------//
	fmt.Println("case 2 : response code = 404,target not found")
	id = "1111"
	httpStatusCode, err = apiTest.GetTargetPoliciesByID(*admin, id)
	if err != nil {
		t.Error("Error whihle get target by id", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

}

func TestTargetsDelete(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	id := strconv.Itoa(addTargetID)
	fmt.Println("Testing Targets Delete API")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	httpStatusCode, err = apiTest.DeleteTargetsByID(*admin, id)
	if err != nil {
		t.Error("Error whihle delete targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//--------------case 2 : response code = 404,target not found------------//
	fmt.Println("case 2 : response code = 404,target not found")
	id = "1111"
	httpStatusCode, err = apiTest.DeleteTargetsByID(*admin, id)
	if err != nil {
		t.Error("Error whihle delete targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

}
