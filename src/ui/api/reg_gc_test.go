package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
)

var adminJob001 apilib.GCReq
var adminJob001schdeule apilib.ScheduleParam

func TestAdminJobPost(t *testing.T) {

	assert := assert.New(t)
	apiTest := newHarborAPI()

	//case 1: add a new admin job
	code, err := apiTest.AddGC(*admin, adminJob001)
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

	code, _, err := apiTest.GCScheduleGet(*admin)
	if err != nil {
		t.Error("Error occured while get a admin job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get adminjob status should be 200")
	}
}
