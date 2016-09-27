package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	//	"github.com/vmware/harbor/tests/apitests/apilib"
	//	"strconv"
)

func TestGetRepos(t *testing.T) {
	var httpStatusCode int
	var err error

	assert := assert.New(t)
	apiTest := newHarborAPI()
	projectID := "1"

	fmt.Println("Testing Repos Get API")
	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	httpStatusCode, err = apiTest.GetRepos(*admin, projectID)
	if err != nil {
		t.Error("Error whihle get repos by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
	//-------------------case 2 : response code = 400------------------------//
	fmt.Println("case 2 : response code = 409,invalid project_id")
	projectID = "ccc"
	httpStatusCode, err = apiTest.GetRepos(*admin, projectID)
	if err != nil {
		t.Error("Error whihle get repos by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}
	//-------------------case 3 : response code = 404------------------------//
	fmt.Println("case 3 : response code = 404:project  not found")
	projectID = "111"
	httpStatusCode, err = apiTest.GetRepos(*admin, projectID)
	if err != nil {
		t.Error("Error whihle get repos by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	fmt.Printf("\n")
}

func TestGetReposTags(t *testing.T) {
	var httpStatusCode int
	var err error
	var repoName string

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing ReposTags Get API")
	//-------------------case 1 : response code = 400------------------------//
	fmt.Println("case 1 : response code = 400,repo_name is nil")
	repoName = ""
	httpStatusCode, err = apiTest.GetReposTags(*admin, repoName)
	if err != nil {
		t.Error("Error whihle get reposTags by repoName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}
	//-------------------case 2 : response code = 404------------------------//
	fmt.Println("case 2 : response code = 404,repo not found")
	repoName = "errorRepos"
	httpStatusCode, err = apiTest.GetReposTags(*admin, repoName)
	if err != nil {
		t.Error("Error whihle get reposTags by repoName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}
	//-------------------case 3 : response code = 200------------------------//
	fmt.Println("case 3 : response code = 200")
	repoName = "library/hello-world"
	httpStatusCode, err = apiTest.GetReposTags(*admin, repoName)
	if err != nil {
		t.Error("Error whihle get reposTags by repoName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	fmt.Printf("\n")
}

func TestGetReposManifests(t *testing.T) {
	var httpStatusCode int
	var err error
	var repoName string
	var tag string

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing ReposManifests Get API")
	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	repoName = "library/hello-world"
	tag = "latest"
	httpStatusCode, err = apiTest.GetReposManifests(*admin, repoName, tag)
	if err != nil {
		t.Error("Error whihle get reposManifests by repoName and tag", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
	//-------------------case 2 : response code = 404------------------------//
	fmt.Println("case 2 : response code = 404:tags error,manifest unknown")
	tag = "l"
	httpStatusCode, err = apiTest.GetReposManifests(*admin, repoName, tag)
	if err != nil {
		t.Error("Error whihle get reposManifests by repoName and tag", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	//-------------------case 3 : response code = 400------------------------//
	fmt.Println("case 3 : response code = 400,repo_name or is nil")
	repoName = ""
	httpStatusCode, err = apiTest.GetReposManifests(*admin, repoName, tag)
	if err != nil {
		t.Error("Error whihle get reposManifests by repoName and tag", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}
	//-------------------case 4 : response code = 404------------------------//
	fmt.Println("case 4 : response code = 404,repo not found")
	repoName = "111"
	httpStatusCode, err = apiTest.GetReposManifests(*admin, repoName, tag)
	if err != nil {
		t.Error("Error whihle get reposManifests by repoName and tag", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	fmt.Printf("\n")
}

func TestGetReposTop(t *testing.T) {
	var httpStatusCode int
	var err error
	var count string

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing ReposTop Get API")
	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	count = "1"
	httpStatusCode, err = apiTest.GetReposTop(*admin, count)
	if err != nil {
		t.Error("Error whihle get reposTop to show the most popular public repositories ", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	//-------------------case 2 : response code = 400------------------------//
	fmt.Println("case 2 : response code = 400,invalid count")
	count = "cc"
	httpStatusCode, err = apiTest.GetReposTop(*admin, count)
	if err != nil {
		t.Error("Error whihle get reposTop to show the most popular public repositories ", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}

	fmt.Printf("\n")
}
