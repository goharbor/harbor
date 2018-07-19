package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
)

var adminJob001 apilib.AdminJobReq

func TestAdminJobPost(t *testing.T) {

	fmt.Println("Testing Add Admin Job")

	assert := assert.New(t)
	apiTest := newHarborAPI()

	//case 1: add a new admin job
	adminJob001.Name = "gc"
	adminJob001.Kind = "Generic"
	code, err := apiTest.AddAdminJob(*admin, adminJob001)
	if err != nil {
		t.Error("Error occured while add a admin job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Add adminjob status should be 200")
	}

}

func TestAdminJobGet(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	code, jobs, err := apiTest.AdminJobGet(*admin)
	if err != nil {
		t.Error("Error occured while get a admin job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get adminjob status should be 200")
		assert.Equal(1, len(jobs), "Get adminjob record should be 1 ")
		assert.Equal(jobs[0].Name, "gc", "Get adminjob one should be gc job")
	}
}
