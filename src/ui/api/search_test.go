// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
		assert.Equal(int64(1), result.Projects[0].ProjectID, "Project id should be equal")
		assert.Equal("library", result.Projects[0].Name, "Project name should be library")
		assert.Equal(1, result.Projects[0].Public, "Project public status should be 1 (true)")
	}

	//--------case 2 : Response Code  = 200, sysAdmin and search repo--------//
	httpStatusCode, result, err = apiTest.SearchGet("library", *admin)
	if err != nil {
		t.Error("Error while search project or repository", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal("library", result.Repositories[0].ProjectName, "Project name should be library")
		assert.Equal("library/docker", result.Repositories[0].RepositoryName, "Repository  name should be library/docker")
		assert.Equal(int32(1), result.Repositories[0].ProjectPublic, "Project public status should be 1 (true)")
	}

	//--------case 3 : Response Code  = 200, normal user and search repo--------//
	httpStatusCode, result, err = apiTest.SearchGet("library", *testUser)
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
