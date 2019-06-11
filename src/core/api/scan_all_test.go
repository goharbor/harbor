package api

import (
	"testing"

	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/stretchr/testify/assert"
)

var adminJob002 apilib.AdminJobReq

func TestScanAllPost(t *testing.T) {

	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: add a new scan all job
	code, err := apiTest.AddScanAll(*admin, adminJob002)
	if err != nil {
		t.Error("Error occurred while add a scan all job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Add scan all status should be 200")
	}
}

func TestScanAllGet(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	code, _, err := apiTest.ScanAllScheduleGet(*admin)
	if err != nil {
		t.Error("Error occurred while get a scan all job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get scan all status should be 200")
	}
}
