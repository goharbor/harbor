package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
	"strconv"
	"testing"
)

const (
	addPolicyName = "testPolicy"
)

var addPolicyID int

func TestPoliciesPost(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	//add target
	CommonAddTarget()
	targetID := int64(CommonGetTarget())
	repPolicy := &apilib.RepPolicyPost{int64(1), targetID, addPolicyName}

	fmt.Println("Testing Policies Post API")

	//-------------------case 1 : response code = 201------------------------//
	fmt.Println("case 1 : response code = 201")
	httpStatusCode, err = apiTest.AddPolicy(*admin, *repPolicy)
	if err != nil {
		t.Error("Error while add policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), httpStatusCode, "httpStatusCode should be 201")
	}

	//-------------------case 2 : response code = 409------------------------//
	fmt.Println("case 1 : response code = 409:policy already exists")
	httpStatusCode, err = apiTest.AddPolicy(*admin, *repPolicy)
	if err != nil {
		t.Error("Error while add policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), httpStatusCode, "httpStatusCode should be 409")
	}

	//-------------------case 3 : response code = 401------------------------//
	fmt.Println("case 3 : response code = 401: User need to log in first.")
	httpStatusCode, err = apiTest.AddPolicy(*unknownUsr, *repPolicy)
	if err != nil {
		t.Error("Error while add policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 401")
	}

	//-------------------case 4 : response code = 400------------------------//
	fmt.Println("case 4 : response code = 400:project_id invalid.")

	repPolicy = &apilib.RepPolicyPost{TargetId: targetID, Name: addPolicyName}
	httpStatusCode, err = apiTest.AddPolicy(*admin, *repPolicy)
	if err != nil {
		t.Error("Error while add policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}

	//-------------------case 5 : response code = 400------------------------//
	fmt.Println("case 5 : response code = 400:project_id does not exist.")

	repPolicy.ProjectId = int64(1111)
	httpStatusCode, err = apiTest.AddPolicy(*admin, *repPolicy)
	if err != nil {
		t.Error("Error while add policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}

	//-------------------case 6 : response code = 400------------------------//
	fmt.Println("case 6 : response code = 400:target_id invalid.")

	repPolicy = &apilib.RepPolicyPost{ProjectId: int64(1), Name: addPolicyName}
	httpStatusCode, err = apiTest.AddPolicy(*admin, *repPolicy)
	if err != nil {
		t.Error("Error while add policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}

	//-------------------case 7 : response code = 400------------------------//
	fmt.Println("case 6 : response code = 400:target_id does not exist.")

	repPolicy.TargetId = int64(1111)
	httpStatusCode, err = apiTest.AddPolicy(*admin, *repPolicy)
	if err != nil {
		t.Error("Error while add policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}

}

func TestPoliciesList(t *testing.T) {
	var httpStatusCode int
	var err error
	var reslut []apilib.RepPolicy

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Policies Get/List API")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	projectID := "1"
	httpStatusCode, reslut, err = apiTest.ListPolicies(*admin, addPolicyName, projectID)
	if err != nil {
		t.Error("Error while get policies", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		addPolicyID = int(reslut[0].Id)
	}

	//-------------------case 2 : response code = 400------------------------//
	fmt.Println("case 2 : response code = 400:invalid projectID")
	projectID = "cc"
	httpStatusCode, reslut, err = apiTest.ListPolicies(*admin, addPolicyName, projectID)
	if err != nil {
		t.Error("Error while get policies", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}

}

func TestPolicyGet(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Policy Get API by PolicyID")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")

	policyID := strconv.Itoa(addPolicyID)
	httpStatusCode, err = apiTest.GetPolicyByID(*admin, policyID)
	if err != nil {
		t.Error("Error while get policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
}

func TestPolicyUpdateInfo(t *testing.T) {
	var httpStatusCode int
	var err error

	targetID := int64(CommonGetTarget())
	policyInfo := &apilib.RepPolicyUpdate{TargetId: targetID, Name: "testNewName"}

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Policy PUT API to update policyInfo")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")

	policyID := strconv.Itoa(addPolicyID)
	httpStatusCode, err = apiTest.PutPolicyInfoByID(*admin, policyID, *policyInfo)
	if err != nil {
		t.Error("Error while update policyInfo", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
}

func TestPolicyUpdateEnablement(t *testing.T) {
	var httpStatusCode int
	var err error

	enablement := &apilib.RepPolicyEnablementReq{int32(0)}

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Policy PUT API to update policy enablement")

	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")

	policyID := strconv.Itoa(addPolicyID)
	httpStatusCode, err = apiTest.PutPolicyEnableByID(*admin, policyID, *enablement)
	if err != nil {
		t.Error("Error while put policy enablement", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
	//-------------------case 2 : response code = 404------------------------//
	fmt.Println("case 2 : response code = 404,Not Found")

	policyID = "111"
	httpStatusCode, err = apiTest.PutPolicyEnableByID(*admin, policyID, *enablement)
	if err != nil {
		t.Error("Error while put policy enablement", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

}

func TestPolicyDelete(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing Policy Delete API")

	//-------------------case 1 : response code = 412------------------------//
	fmt.Println("case 1 : response code = 412:policy is enabled, can not be deleted")

	CommonPolicyEabled(addPolicyID, 1)
	policyID := strconv.Itoa(addPolicyID)

	httpStatusCode, err = apiTest.DeletePolicyByID(*admin, policyID)
	if err != nil {
		t.Error("Error while delete policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(412), httpStatusCode, "httpStatusCode should be 412")
	}

	//-------------------case 2 : response code = 200------------------------//
	fmt.Println("case 2 : response code = 200")

	CommonPolicyEabled(addPolicyID, 0)
	policyID = strconv.Itoa(addPolicyID)

	httpStatusCode, err = apiTest.DeletePolicyByID(*admin, policyID)
	if err != nil {
		t.Error("Error while delete policy", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	CommonDelTarget()
}
