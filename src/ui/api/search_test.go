package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
)

func TestSearch(t *testing.T) {
	fmt.Println("Testing Search(SearchGet) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()
	var result apilib.Search

	//-------------case 1 : Response Code  = 200, Not sysAdmin --------------//
	httpStatusCode, result, err := apiTest.SearchGet("library")
	if err != nil {
		t.Error("Error while search project or repository", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(int64(1), result.Projects[0].Id, "Project id should be equal")
		assert.Equal("library", result.Projects[0].Name, "Project name should be library")
		assert.Equal(int32(1), result.Projects[0].Public, "Project public status should be 1 (true)")
	}

	//--------case 2 : Response Code  = 200, sysAdmin and search repo--------//
	httpStatusCode, result, err = apiTest.SearchGet("docker", *admin)
	if err != nil {
		t.Error("Error while search project or repository", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal("library", result.Repositories[0].ProjectName, "Project name should be library")
		assert.Equal("library/docker", result.Repositories[0].RepositoryName, "Repository  name should be library/docker")
		assert.Equal(int32(1), result.Repositories[0].ProjectPublic, "Project public status should be 1 (true)")
	}

}
