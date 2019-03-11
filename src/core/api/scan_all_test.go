package api

import (
	"testing"

	"github.com/goharbor/harbor/tests/apitests/apilib"
	"github.com/stretchr/testify/assert"
)

var adminJob002 apilib.AdminJobReq

func TestScanAllPost(t *testing.T) {

	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: add a new admin job
	code, err := apiTest.AddScanAll(*admin, adminJob002)
	if err != nil {
		t.Error("Error occurred while add a admin job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Add adminjob status should be 200")
	}
}

func TestScanAllGet(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	code, _, err := apiTest.ScanAllScheduleGet(*admin)
	if err != nil {
		t.Error("Error occurred while get a admin job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get adminjob status should be 200")
	}
}
