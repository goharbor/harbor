package HarborAPItest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
        "github.com/vmware/harbor/tests/apitests/apilib"
)

func TestSearch(t *testing.T) {
        fmt.Println("Test for Search (SearchGet) API")
	assert := assert.New(t)

	apiTest := HarborAPI.NewHarborAPI()
	var resault HarborAPI.Search
        resault, err := apiTest.SearchGet("library")
	//fmt.Printf("%+v\n", resault)
	if err != nil {
		t.Error("Error while search project or repository", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault.Projects[0].ProjectId, int32(1), "Project id should be equal")
		assert.Equal(resault.Projects[0].ProjectName, "library", "Project name should be library")
		assert.Equal(resault.Projects[0].Public, int32(1), "Project public status should be 1 (true)")
		//t.Log(resault)
	}
	//if resault.Response.StatusCode != 200 {
	//	t.Log(resault.Response)
	//}

}
