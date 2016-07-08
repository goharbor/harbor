package HarborAPItest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
        "github.com/vmware/harbor/tests/apitests/apilib"
)

func TestAddProject(t *testing.T) {
        
        fmt.Println("Test for Project Add (ProjectsPost) API")
	assert := assert.New(t)

	apiTest := HarborAPI.NewHarborAPI()

	//prepare for test
	adminEr := &HarborAPI.UsrInfo{"admin", "Harbor1234"}
	admin := &HarborAPI.UsrInfo{"admin", "Harbor12345"}

	prjUsr := &HarborAPI.UsrInfo{"unknown", "unknown"}

	var project HarborAPI.Project
	project.ProjectName = "testproject"
	project.Public = true

	//case 1: admin login fail, expect project creation fail.
	fmt.Println("case 1: admin login fail, expect project creation fail.")
	resault, err := apiTest.HarborLogin(*adminEr)
	if err != nil {
		t.Error("Error while admin login", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(401), "Admin login status should be 401")
		//t.Log(resault)
	}

	resault, err = apiTest.ProjectsPost(*prjUsr, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(401), "Case 1: Project creation status should be 401")
		//t.Log(resault)
	}

	//case 2: admin successful login, expect project creation success.
	fmt.Println("case 2: admin successful login, expect project creation success.")
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

	resault, err = apiTest.ProjectsPost(*prjUsr, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(201), "Case 2: Project creation status should be 201")
		//t.Log(resault)
	}

	//case 3: duplicate project name, create project fail
	fmt.Println("case 3: duplicate project name, create project fail")
	resault, err = apiTest.ProjectsPost(*prjUsr, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(409), "Case 3: Project creation status should be 409")
		//t.Log(resault)
	}

	//resault1, err := apiTest.HarborLogout()
	//if err != nil {
	//        t.Error("Error while admin logout", err.Error())
	//        t.Log(err)
	//} else {
	//        assert.Equal(resault1, int(200), "Admin logout status")
	//        //t.Log(resault)
	//}
	//if resault1 != 200 {
	//        t.Log(resault)
	//}

}
