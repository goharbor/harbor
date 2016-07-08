package HarborAPItest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
)

func TestRepositoryDelete(t *testing.T) {
	fmt.Println("Test for Project Delete (ProjectDelete) API")
	assert := assert.New(t)

	//prepare for test
	adminEr := &HarborAPI.UsrInfo{"admin", "Harbor1234"}
	admin := &HarborAPI.UsrInfo{"admin", "Harbor12345"}
	prjUsr := &HarborAPI.UsrInfo{"unknown", "unknown"}

	fmt.Println("Checking repository status...")
	apiTest := HarborAPI.NewHarborAPI()
	var searchResault HarborAPI.Search
	searchResault, err := apiTest.SearchGet("library")
	//fmt.Printf("%+v\n", resault)

	if err != nil {
		t.Error("Error while search project or repository", err.Error())
		t.Log(err)
	} else {
		//assert.Equal(searchResault.Repositories[0].RepoName, "library/docker", "1st repo name should be")
		if !assert.Equal(searchResault.Repositories[0].RepoName, "library/docker", "1st repo name should be") {
			t.Error("fail to find repo 'library/docker'", err.Error())
			t.Log(err)
		} else {
			fmt.Println("repo 'library/docker' exit")
		}
		//assert.Equal(searchResault.Repositories[1].RepoName, "library/hello-world", "2nd repo name should be")
		if !assert.Equal(searchResault.Repositories[1].RepoName, "library/hello-world", "2nd repo name should be") {
			t.Error("fail to find repo 'library/hello-world'", err.Error())
			t.Log(err)
		} else {
			fmt.Println("repo 'library/hello-world' exit")
		}

		//t.Log(resault)
	}

	//case 1: admin login fail, expect repo delete fail.
	fmt.Println("case 1: admin login fail, expect repo delete fail.")

	resault, err := apiTest.HarborLogin(*adminEr)
	if err != nil {
		t.Error("Error while admin login", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(401), "Admin login status should be 401")
		//t.Log(resault)
	}
	if resault != 401 {
		t.Log(resault)
	} else {
		prjUsr = adminEr
	}

	resault, err = apiTest.RepositoriesDelete(*prjUsr, "library/docker", "")
	if err != nil {
		t.Error("Error while delete repository", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(401), "Case 1: Repository delete status should be 401")
		//t.Log(resault)
	}

	//case 2: admin successful login, expect repository delete success.
	fmt.Println("case 2: admin successful login, expect repository delete success.")
	resault, err = apiTest.HarborLogin(*admin)
	if err != nil {
		t.Error("Error while admin login", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(200), "Admin login status should be 200")
		//t.Log(resault)
	}
	if resault != 200 {
		t.Log(resault)
	} else {
		prjUsr = admin
	}

	resault, err = apiTest.RepositoriesDelete(*prjUsr, "library/docker", "")
	if err != nil {
		t.Error("Error while delete repository", err.Error())
		t.Log(err)
	} else {
		if assert.Equal(resault, int(200), "Case 2: Repository delete status should be 200") {
			fmt.Println("Repository 'library/docker' delete success.")
		}
		//t.Log(resault)
	}

	resault, err = apiTest.RepositoriesDelete(*prjUsr, "library/hello-world", "")
	if err != nil {
		t.Error("Error while delete repository", err.Error())
		t.Log(err)
	} else {
		if assert.Equal(resault, int(200), "Case 2: Repository delete status should be 200") {
			fmt.Println("Repository 'hello-world' delete success.")
		}
		//t.Log(resault)
	}

	//case 3: delete one repo not exit, expect repo delete fail.
	fmt.Println("case 3: delete one repo not exit, expect repo delete fail.")

	resault, err = apiTest.RepositoriesDelete(*prjUsr, "library/hello-world", "")
	if err != nil {
		t.Error("Error while delete repository", err.Error())
		t.Log(err)
	} else {
		if assert.Equal(resault, int(404), "Case 3: Repository delete status should be 404") {
			fmt.Println("Repository 'hello-world' not exit.")
		}
		//t.Log(resault)
	}

	//if resault.Response.StatusCode != 200 {
	//	t.Log(resault.Response)
	//}

}
