// Copyright 2018 Project Harbor Authors
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
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRepos(t *testing.T) {

	assert := assert.New(t)
	apiTest := newHarborAPI()
	projectID := "1"
	keyword := "library/hello-world"

	fmt.Println("Testing Repos Get API")
	// -------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	code, repositories, err := apiTest.GetRepos(*admin, projectID, keyword)
	if err != nil {
		t.Errorf("failed to get repositories: %v", err)
	} else {
		assert.Equal(int(200), code, "response code should be 200")
		if repos, ok := repositories.([]repoResp); ok {
			require.Equal(t, int(1), len(repos), "the length of repositories should be 1")
			assert.Equal(repos[0].Name, "library/hello-world", "unexpected repository name")
		} else {
			t.Error("unexpected response")
		}
	}

	// -------------------case 2 : response code = 404------------------------//
	fmt.Println("case 2 : response code = 404:project  not found")
	projectID = "111"
	httpStatusCode, _, err := apiTest.GetRepos(*admin, projectID, keyword)
	if err != nil {
		t.Error("Error whihle get repos by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	// -------------------case 3 : response code = 400------------------------//
	fmt.Println("case 3 : response code = 400,invalid project_id")
	projectID = "ccc"
	httpStatusCode, _, err = apiTest.GetRepos(*admin, projectID, keyword)
	if err != nil {
		t.Error("Error whihle get repos by projectID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}

	fmt.Printf("\n")
}

func TestGetReposTags(t *testing.T) {

	assert := assert.New(t)
	apiTest := newHarborAPI()

	// -------------------case 1 : response code = 404------------------------//
	fmt.Println("case 1 : response code = 404,repo not found")
	repository := "errorRepos"
	code, _, err := apiTest.GetReposTags(*admin, repository)
	if err != nil {
		t.Errorf("failed to get tags of repository %s: %v", repository, err)
	} else {
		assert.Equal(int(404), code, "httpStatusCode should be 404")
	}
	// -------------------case 2 : response code = 200------------------------//
	fmt.Println("case 2 : response code = 200")
	repository = "library/hello-world"
	code, tags, err := apiTest.GetReposTags(*admin, repository)
	if err != nil {
		t.Errorf("failed to get tags of repository %s: %v", repository, err)
	} else {
		assert.Equal(int(200), code, "httpStatusCode should be 200")
		if tg, ok := tags.([]models.TagResp); ok {
			assert.Equal(1, len(tg), fmt.Sprintf("there should be only one tag, but now %v", tg))
			assert.Equal(tg[0].Name, "latest", "the tag should be latest")
		} else {
			t.Error("unexpected response")
		}
	}

	// -------------------case 3 : response code = 404------------------------//
	fmt.Println("case 3 : response code = 404")
	repository = "library/hello-world"
	tag := "not_exist_tag"
	code, result, err := apiTest.GetTag(*admin, repository, tag)
	assert.Nil(err)
	assert.Equal(http.StatusNotFound, code)

	// -------------------case 4 : response code = 200------------------------//
	fmt.Println("case 4 : response code = 200")
	repository = "library/hello-world"
	tag = "latest"
	code, result, err = apiTest.GetTag(*admin, repository, tag)
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal(tag, result.Name)

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
	// -------------------case 1 : response code = 200------------------------//
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
	// -------------------case 2 : response code = 404------------------------//
	fmt.Println("case 2 : response code = 404:tags error,manifest unknown")
	tag = "l"
	httpStatusCode, err = apiTest.GetReposManifests(*admin, repoName, tag)
	if err != nil {
		t.Error("Error whihle get reposManifests by repoName and tag", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	// -------------------case 3 : response code = 404------------------------//
	fmt.Println("case 3 : response code = 404,repo not found")
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

	assert := assert.New(t)
	apiTest := newHarborAPI()

	fmt.Println("Testing ReposTop Get API")
	// -------------------case 1 : response code = 400------------------------//
	fmt.Println("case 1 : response code = 400,invalid count")
	count := "cc"
	code, _, err := apiTest.GetReposTop(*admin, count)
	if err != nil {
		t.Errorf("failed to get the most popular repositories: %v", err)
	} else {
		assert.Equal(int(400), code, "response code should be 400")
	}

	// -------------------case 2 : response code = 200------------------------//
	fmt.Println("case 2 : response code = 200")
	count = "1"
	code, repos, err := apiTest.GetReposTop(*admin, count)
	if err != nil {
		t.Errorf("failed to get the most popular repositories: %v", err)
	} else {
		assert.Equal(int(200), code, "response code should be 200")
		if r, ok := repos.([]*repoResp); ok {
			assert.Equal(int(1), len(r), "the length should be 1")
			assert.Equal(r[0].Name, "library/busybox", "the name of repository should be library/busybox")
		} else {
			t.Error("unexpected response")
		}
	}

	fmt.Printf("\n")
}

func TestPopulateAuthor(t *testing.T) {
	author := "author"
	detail := &models.TagDetail{
		Author: author,
	}
	populateAuthor(detail)
	assert.Equal(t, author, detail.Author)

	detail = &models.TagDetail{}
	populateAuthor(detail)
	assert.Equal(t, "", detail.Author)

	maintainer := "maintainer"
	detail = &models.TagDetail{
		Config: &models.TagCfg{
			Labels: map[string]string{
				"Maintainer": maintainer,
			},
		},
	}
	populateAuthor(detail)
	assert.Equal(t, maintainer, detail.Author)
}

func TestPutOfRepository(t *testing.T) {
	base := "/api/repositories/"
	desc := struct {
		Description string `json:"description"`
	}{
		Description: "description_for_test",
	}

	cases := []*codeCheckingCase{
		// 404
		{
			request: &testingRequest{
				method:   http.MethodPut,
				url:      base + "non_exist_repository",
				bodyJSON: desc,
			},
			code: http.StatusNotFound,
		},
		// 401
		{
			request: &testingRequest{
				method:   http.MethodPut,
				url:      base + "library/hello-world",
				bodyJSON: desc,
			},
			code: http.StatusUnauthorized,
		},
		// 403 non-member
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        base + "library/hello-world",
				bodyJSON:   desc,
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 403 project guest
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        base + "library/hello-world",
				bodyJSON:   desc,
				credential: projGuest,
			},
			code: http.StatusForbidden,
		},
		// 200 project developer
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        base + "library/hello-world",
				bodyJSON:   desc,
				credential: projDeveloper,
			},
			code: http.StatusOK,
		},
		// 200 project admin
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        base + "library/hello-world",
				bodyJSON:   desc,
				credential: projAdmin,
			},
			code: http.StatusOK,
		},
		// 200 system admin
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        base + "library/hello-world",
				bodyJSON:   desc,
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)

	// verify that the description is changed
	repositories := []*repoResp{}
	err := handleAndParse(&testingRequest{
		method: http.MethodGet,
		url:    base,
		queryStruct: struct {
			ProjectID int64 `url:"project_id"`
		}{
			ProjectID: 1,
		},
	}, &repositories)
	require.Nil(t, err)
	var repository *repoResp
	for _, repo := range repositories {
		if repo.Name == "library/hello-world" {
			repository = repo
			break
		}
	}
	require.NotNil(t, repository)
	assert.Equal(t, desc.Description, repository.Description)
}

func TestRetag(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()
	repo := "library/hello-world"

	fmt.Println("Testing Image Retag API")
	// -------------------case 1 : response code = 200------------------------//
	fmt.Println("case 1 : response code = 200")
	retagReq := &apilib.Retag{
		Tag:      "prd",
		SrcImage: "library/hello-world:latest",
		Override: true,
	}
	code, err := apiTest.RetagImage(*admin, repo, retagReq)
	if err != nil {
		t.Errorf("failed to retag: %v", err)
	} else {
		assert.Equal(int(200), code, "response code should be 200")
	}

	// -------------------case 2 : response code = 400------------------------//
	fmt.Println("case 2 : response code = 400: invalid image value provided")
	retagReq = &apilib.Retag{
		Tag:      "prd",
		SrcImage: "hello-world:latest",
		Override: true,
	}
	httpStatusCode, err := apiTest.RetagImage(*admin, repo, retagReq)
	if err != nil {
		t.Errorf("failed to retag: %v", err)
	} else {
		assert.Equal(int(400), httpStatusCode, "httpStatusCode should be 400")
	}

	// -------------------case 3 : response code = 404------------------------//
	fmt.Println("case 3 : response code = 404: source image not exist")
	retagReq = &apilib.Retag{
		Tag:      "prd",
		SrcImage: "release/hello-world:notexist",
		Override: true,
	}
	httpStatusCode, err = apiTest.RetagImage(*admin, repo, retagReq)
	if err != nil {
		t.Errorf("failed to retag: %v", err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	// -------------------case 4 : response code = 404------------------------//
	fmt.Println("case 4 : response code = 404: target project not exist")
	retagReq = &apilib.Retag{
		Tag:      "prd",
		SrcImage: "library/hello-world:latest",
		Override: true,
	}
	httpStatusCode, err = apiTest.RetagImage(*admin, "nonexist/hello-world", retagReq)
	if err != nil {
		t.Errorf("failed to retag: %v", err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	// -------------------case 5 : response code = 401------------------------//
	fmt.Println("case 5 : response code = 401, unathorized")
	retagReq = &apilib.Retag{
		Tag:      "prd",
		SrcImage: "library/hello-world:latest",
		Override: true,
	}
	httpStatusCode, err = apiTest.RetagImage(*unknownUsr, repo, retagReq)
	if err != nil {
		t.Errorf("failed to retag: %v", err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 401")
	}

	// -------------------case 6 : response code = 409------------------------//
	fmt.Println("case 6 : response code = 409, conflict")
	retagReq = &apilib.Retag{
		Tag:      "latest",
		SrcImage: "library/hello-world:latest",
		Override: false,
	}
	httpStatusCode, err = apiTest.RetagImage(*admin, repo, retagReq)
	if err != nil {
		t.Errorf("failed to retag: %v", err)
	} else {
		assert.Equal(int(409), httpStatusCode, "httpStatusCode should be 409")
	}

	// -------------------case 7 : response code = 400------------------------//
	fmt.Println("case 7 : response code = 400")
	retagReq = &apilib.Retag{
		Tag:      ".0.1",
		SrcImage: "library/hello-world:latest",
		Override: true,
	}
	code, err = apiTest.RetagImage(*admin, repo, retagReq)
	if err != nil {
		t.Errorf("failed to retag: %v", err)
	} else {
		assert.Equal(int(400), code, "response code should be 400")
	}

	// -------------------case 8 : response code = 400------------------------//
	fmt.Println("case 8 : response code = 400")
	retagReq = &apilib.Retag{
		Tag:      "v0.1",
		SrcImage: "library/hello-world:latest",
		Override: true,
	}
	code, err = apiTest.RetagImage(*admin, "library/Aaaa", retagReq)
	if err != nil {
		t.Errorf("failed to retag: %v", err)
	} else {
		assert.Equal(int(400), code, "response code should be 400")
	}

	fmt.Printf("\n")
}
