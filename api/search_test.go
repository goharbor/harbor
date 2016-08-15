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
	result, err := apiTest.SearchGet("library")
	//fmt.Printf("%+v\n", result)
	if err != nil {
		t.Error("Error while search project or repository", err.Error())
		t.Log(err)
	} else {
		assert.Equal(result.Projects[0].Id, int64(1), "Project id should be equal")
		assert.Equal(result.Projects[0].Name, "library", "Project name should be library")
		assert.Equal(result.Projects[0].Public, int32(1), "Project public status should be 1 (true)")
		//t.Log(result)
	}
	//if result.Response.StatusCode != 200 {
	//	t.Log(result.Response)
	//}

}
