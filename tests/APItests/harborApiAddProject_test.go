package HarborApi

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddProject(t *testing.T) {

	assert := assert.New(t)

	apiTest := NewHarborApi()

	//prepare test
	adminEr := &UsrInfo{"admin", "Harbor1234"}
	admin := &UsrInfo{"admin", "Harbor12345"}

	var prjUsr = new(UsrInfo)
	prjUsr.Name = "unknown"
	prjUsr.Passwd = "unknown"

	var project Project
	project.ProjectName = "testProject"
	project.Public = true

	//case 1: admin login fail, expect project creation fail.
	fmt.Println("case 1: admin login fail, expect project creation fail.")
	resault, err := apiTest.HarborLogin(*adminEr)
	if err != nil {
		t.Errorf("Error while admin login", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(401), "Admin login status should be 401")
		//t.Log(resault)
	}

	resault, err = apiTest.ProjectsPost(*prjUsr, project)
	if err != nil {
		t.Errorf("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(401), "Case 1: Project creation status should be 401")
		//t.Log(resault)
	}

	//case 2: admin successful login, expect project creation success.
	fmt.Println("case 2: admin successful login, expect project creation success.")
	resault, err = apiTest.HarborLogin(*admin)
	if err != nil {
		t.Errorf("Error while admin login", err.Error())
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
		t.Errorf("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(200), "Case 2: Project creation status should be 200")
		//t.Log(resault)
	}

	//case 3: duplicate project name, create project fail
	fmt.Println("case 3: duplicate project name, create project fail")
	resault, err = apiTest.ProjectsPost(*prjUsr, project)
	if err != nil {
		t.Errorf("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(resault, int(409), "Case 3: Project creation status should be 409")
		//t.Log(resault)
	}

	//resault1, err := apiTest.HarborLogout()
	//if err != nil {
	//        t.Errorf("Error while admin logout", err.Error())
	//        t.Log(err)
	//} else {
	//        assert.Equal(resault1, int(200), "Admin logout status")
	//        //t.Log(resault)
	//}
	//if resault1 != 200 {
	//        t.Log(resault)
	//}

}
