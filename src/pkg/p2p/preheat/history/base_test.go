package history

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateHistoryRecord(t *testing.T) {
	// validate nil
	var nilHistory *models.HistoryRecord
	err := validHistoryRecord(nilHistory)
	assert.NotNil(t, err)

	// lack some field
	invalidHistory := &models.HistoryRecord{
		ID:         0,
		TaskID:     "",
		Image:      "",
		StartTime:  "",
		FinishTime: "",
		Status:     "",
		Provider:   "",
		Instance:   0,
	}
	err = validHistoryRecord(invalidHistory)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "missing [TaskID,Image,StartTime,FinishTime,Status,Provider,Instance]")

	// valid history
	validHistory := &models.HistoryRecord{
		ID:         1,
		TaskID:     "24b4043f-7383-405b-8ed6-39f8e3591641",
		Image:      "library/ubuntu",
		StartTime:  "2020-01-22 10:22:11",
		FinishTime: "2020-01-22 22:10:08",
		Status:     "success",
		Provider:   "kraken",
		Instance:   1,
	}
	err = validHistoryRecord(validHistory)
	assert.Nil(t, err)
}
