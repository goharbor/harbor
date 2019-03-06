package api

import (
	"testing"

	common_models "github.com/goharbor/harbor/src/common/models"
	api_modes "github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/tests/apitests/apilib"
	"github.com/stretchr/testify/assert"
)

var adminJob001 apilib.GCReq
var adminJob001schdeule apilib.ScheduleParam

func TestAdminJobPost(t *testing.T) {

	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: add a new admin job
	code, err := apiTest.AddGC(*admin, adminJob001)
	if err != nil {
		t.Error("Error occurred while add a admin job", err.Error())
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
		t.Error("Error occurred while get a admin job", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get adminjob status should be 200")
	}
}

func TestConvertToGCRep(t *testing.T) {
	cases := []struct {
		input    *common_models.AdminJob
		expected api_modes.GCRep
	}{
		{
			input:    nil,
			expected: api_modes.GCRep{},
		},
		{
			input: &common_models.AdminJob{
				ID:      1,
				Name:    "IMAGE_GC",
				Kind:    "Generic",
				Cron:    "{\"Type\":\"Daily\",\"Cron\":\"20 3 0 * * *\"}",
				Status:  "pending",
				Deleted: false,
			},
			expected: api_modes.GCRep{
				ID:   1,
				Name: "IMAGE_GC",
				Kind: "Generic",
				Schedule: &api_modes.ScheduleParam{
					Type: "Daily",
					Cron: "20 3 0 * * *",
				},
				Status:  "pending",
				Deleted: false,
			},
		},
	}

	for _, c := range cases {
		actual, _ := convertToGCRep(c.input)
		assert.EqualValues(t, c.expected, actual)
	}
}
