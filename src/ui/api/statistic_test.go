package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/vmware/harbor/tests/apitests/apilib"
)

func TestStatisticGet(t *testing.T) {

	fmt.Println("Testing Statistic API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	//prepare for test

	var priMyProjectCount, priMyRepoCount int32
	var priPublicProjectCount, priPublicRepoCount int32
	var priTotalProjectCount, priTotalRepoCount int32

	//case 1: case 1: user not login, expect fail to get status info.
	fmt.Println("case 1: user not login, expect fail to get status info.")
	httpStatusCode, result, err := apiTest.StatisticGet(*unknownUsr)
	if err != nil {
		t.Error("Error get statistic info.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(httpStatusCode, int(401), "Case 1: Get status info without login. (401)")
	}

	//case 2: admin successful login, expect get status info successful.
	fmt.Println("case 2: admin successful login, expect get status info successful.")
	httpStatusCode, result, err = apiTest.StatisticGet(*admin)
	if err != nil {
		t.Error("Error get statistic info.", err.Error())
		t.Log(err)
	} else {
		assert.Equal(httpStatusCode, int(200), "Case 2: Get status info with admin login. (200)")
		//fmt.Println("pri status data %+v", result)
		priMyProjectCount = result.MyProjectCount
		priMyRepoCount = result.MyRepoCount
		priPublicProjectCount = result.PublicProjectCount
		priPublicRepoCount = result.PublicRepoCount
		priTotalProjectCount = result.TotalProjectCount
		priTotalRepoCount = result.TotalRepoCount
	}

	//case 3: status info increased after add more project and repo.
	fmt.Println("case 3: status info increased after add more project and repo.")

	CommonAddProject()
	CommonAddRepository()

	httpStatusCode, result, err = apiTest.StatisticGet(*admin)
	//fmt.Println("new status data %+v", result)

	if err != nil {
		t.Error("Error while get statistic information", err.Error())
		t.Log(err)
	} else {
		assert.Equal(priMyProjectCount+1, result.MyProjectCount, "MyProjectCount should be +1")
		assert.Equal(priMyRepoCount+1, result.MyRepoCount, "MyRepoCount should be +1")
		assert.Equal(priPublicProjectCount, result.PublicProjectCount, "PublicProjectCount should be equal")
		assert.Equal(priPublicRepoCount+1, result.PublicRepoCount, "PublicRepoCount should be +1")
		assert.Equal(priTotalProjectCount+1, result.TotalProjectCount, "TotalProCount should be +1")
		assert.Equal(priTotalRepoCount+1, result.TotalRepoCount, "TotalRepoCount should be +1")

	}

	//delete the project and repo
	CommonDelProject()
	CommonDelRepository()
}
