package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRepos(t *testing.T) {

	assert := assert.New(t)
	apiTest := newHarborAPI()
	projectID := "1"
	keyword := "hello-world"
	detail := "true"

	fmt.Println("Testing Repos Get API")
	//-------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	code, repositories, err := apiTest.GetRepos(*admin, projectID, keyword, detail)
	if err != nil {
		t.Errorf("failed to get repositories: %v", err)
	} else {
		assert.Equal(int(200), code, "response code should be 200")
		if repos, ok := repositories.([]repoResp); ok {
			assert.Equal(int(1), len(repos), "the length of repositories should be 1")
			assert.Equal(repos[0].Name, "library/hello-world", "unexpected repository name")
		} else {
			t.Error("the response should return more info as detail is true")
		}
	}

	//-------------------case 2 : response code = 400------------------------//
	fmt.Println("case 2 : response code = 400,invalid project_id")
	projectID = "ccc"
	httpStatusCode, _, err := apiTest.GetRepos(*admin, projectID, keyword, detail)
	if err != nil {
		t.Error("Error whihle get repos by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}
	//-------------------case 3 : response code = 404------------------------//
	fmt.Println("case 3 : response code = 404:project  not found")
	projectID = "111"
	httpStatusCode, _, err = apiTest.GetRepos(*admin, projectID, keyword, detail)
	if err != nil {
		t.Error("Error whihle get repos by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	//-------------------case 4 : response code = 200------------------------//
	fmt.Println("case 4 : response code = 200")
	projectID = "1"
	detail = "false"
	code, repositories, err = apiTest.GetRepos(*admin, projectID, keyword, detail)
	if err != nil {
		t.Errorf("failed to get repositories: %v", err)
	} else {
		assert.Equal(int(200), code, "response code should be 200")
		if repos, ok := repositories.([]string); ok {
			assert.Equal(int(1), len(repos), "the length of repositories should be 1")
			assert.Equal(repos[0], "library/hello-world", "unexpected repository name")
		} else {
			t.Error("the response should not return detail info as detail is false")
		}
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
