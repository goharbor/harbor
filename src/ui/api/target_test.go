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
	httpStatusCode, err = apiTest.AddTargets(*admin, *repTargets)
	if err != nil {
		t.Error("Error whihle add targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), httpStatusCode, "httpStatusCode should be 201")
	}

	//-----------case 2 : response code = 409,name is already used-----------//
	fmt.Println("case 2 : response code = 409,name is already used")
	httpStatusCode, err = apiTest.AddTargets(*admin, *repTargets)
	if err != nil {
		t.Error("Error whihle add targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), httpStatusCode, "httpStatusCode should be 409")
	}

	//-----------case 3 : response code = 409,name is already used-----------//
	fmt.Println("case 3 : response code = 409,endPoint is already used")
	repTargets.Username = "errName"
	httpStatusCode, err = apiTest.AddTargets(*admin, *repTargets)
	if err != nil {
		t.Error("Error whihle add targets", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), httpStatusCode, "httpStatusCode should be 409")
	}

	//--------case 4 : response code = 401,User need to log in first.--------//
	fmt.Println("case 4 : response code = 401,User need to log in first.")
	httpStatusCode, err = apiTest.AddTargets(*unknownUsr, *repTargets)
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

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	id := strconv.Itoa(addTargetID)
	httpStatusCode, err = apiTest.PingTargetsByID(*admin, id)
	if err != nil {
		t.Error("Error whihle ping target", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//--------------case 2 : response code = 404,target not found------------//
	fmt.Println("case 2 : response code = 404,target not found")
	id = "1111"
	httpStatusCode, err = apiTest.PingTargetsByID(*admin, id)
	if err != nil {
		t.Error("Error whihle ping target", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	//------------case 3 : response code = 400,targetID is invalid-----------//
	fmt.Println("case 2 : response code = 400,target not found")
	id = "cc"
	httpStatusCode, err = apiTest.PingTargetsByID(*admin, id)
	if err != nil {
		t.Error("Error whihle ping target", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
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
