package api

import (
	"fmt"
	"testing"

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
	fmt.Printf("\n")

}
