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
	"testing"

	"github.com/stretchr/testify/assert"
	// "github.com/goharbor/harbor/src/testing/apitests/apilib"
)

func TestStatisticGet(t *testing.T) {

	fmt.Println("Testing Statistic API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	// prepare for test

	var privateProjectCount, privateRepoCount int32
	var priPublicProjectCount, priPublicRepoCount int32
	var priTotalProjectCount, priTotalRepoCount int32

	// case 1: case 1: user not login, expect fail to get status info.
	fmt.Println("case 1: user not login, expect fail to get status info.")
	httpStatusCode, result, err := apiTest.StatisticGet(*unknownUsr)
	if err != nil {
		t.Error("Error get statistic info.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(httpStatusCode, int(401), "Case 1: Get status info without login. (401)")
	}

	// case 2: admin successful login, expect get status info successful.
	fmt.Println("case 2: admin successful login, expect get status info successful.")
	httpStatusCode, result, err = apiTest.StatisticGet(*admin)
	if err != nil {
		t.Error("Error get statistic info.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(httpStatusCode, int(200), "Case 2: Get status info with admin login. (200)")
		// fmt.Println("pri status data %+v", result)
		privateProjectCount = result.PrivateProjectCount
		privateRepoCount = result.PrivateRepoCount
		priPublicProjectCount = result.PublicProjectCount
		priPublicRepoCount = result.PublicRepoCount
		priTotalProjectCount = result.TotalProjectCount
		priTotalRepoCount = result.TotalRepoCount
	}

	// case 3: status info increased after add more project and repo.
	fmt.Println("case 3: status info increased after add more project and repo.")

	CommonAddProject()
	CommonAddRepository()

	httpStatusCode, result, err = apiTest.StatisticGet(*admin)
	// fmt.Println("new status data %+v", result)

	if err != nil {
		t.Error("Error while get statistic information", err.Error())
		t.Log(err)
	} else {
		assert.Equal(privateProjectCount+1, result.PrivateProjectCount, "PrivateProjectCount should be +1")
		assert.Equal(privateRepoCount, result.PrivateRepoCount)
		assert.Equal(priPublicProjectCount, result.PublicProjectCount, "PublicProjectCount should be equal")
		assert.Equal(priPublicRepoCount+1, result.PublicRepoCount, "PublicRepoCount should be +1")
		assert.Equal(priTotalProjectCount+1, result.TotalProjectCount, "TotalProCount should be +1")
		assert.Equal(priTotalRepoCount+1, result.TotalRepoCount, "TotalRepoCount should be +1")

	}

	// delete the project and repo
	CommonDelProject()
	CommonDelRepository()
}
