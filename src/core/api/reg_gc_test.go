package api

import (
	"testing"

	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/stretchr/testify/assert"
)

var adminJob001 apilib.AdminJobReq

func TestGCPost(t *testing.T) {

	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: add a new admin job
	code, err := apiTest.AddGC(*admin, adminJob001)
	if err != nil {
		t.Error("Error occurred while add a admin job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(201, code, "Add adminjob status should be 201")
	}
}

func TestGCGet(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	code, _, err := apiTest.GCScheduleGet(*admin)
	if err != nil {
		t.Error("Error occurred while get a admin job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get adminjob status should be 200")
	}
}
