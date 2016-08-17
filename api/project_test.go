package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
)

func TestAddProject(t *testing.T) {

	fmt.Println("Testing Add Project(ProjectsPost) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	//prepare for test

	admin := &usrInfo{"admin", "Harbor12345"}

	prjUsr := &usrInfo{"unknown", "unknown"}

	var project apilib.Project
	project.ProjectName = "test_project"
	project.Public = true

	//case 1: admin not login, expect project creation fail.

	result, err := apiTest.ProjectsPost(*prjUsr, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(result, int(401), "Case 1: Project creation status should be 401")
		//t.Log(result)
	}

	//case 2: admin successful login, expect project creation success.
	fmt.Println("case 2: admin successful login, expect project creation success.")

	prjUsr = admin

	result, err = apiTest.ProjectsPost(*prjUsr, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(result, int(201), "Case 2: Project creation status should be 201")
		//t.Log(result)
	}

	//case 3: duplicate project name, create project fail
	fmt.Println("case 3: duplicate project name, create project fail")

	result, err = apiTest.ProjectsPost(*prjUsr, project)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(result, int(409), "Case 3: Project creation status should be 409")
		//t.Log(result)
	}

}
