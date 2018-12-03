package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"time"
)

func TestMethodsOfJobLog(t *testing.T) {
	uuid := "uuid_for_unit_test"
	now := time.Now()
	content := "content for unit text"
	jobLog := &models.JobLog{
		UUID:         uuid,
		CreationTime: now,
		Content:      content,
	}

	// create
	_, err := CreateOrUpdateJobLog(jobLog)
	require.Nil(t, err)

	// update
	updateContent := "content for unit text update"
	jobLog.Content = updateContent
	_, err = CreateOrUpdateJobLog(jobLog)
	require.Nil(t, err)

	// get
	log, err := GetJobLog(uuid)
	require.Nil(t, err)
	assert.Equal(t, now.Second(), log.CreationTime.Second())
	assert.Equal(t, updateContent, log.Content)
	assert.Equal(t, jobLog.LogID, log.LogID)

	// delete
	count, err := DeleteJobLogsBefore(time.Now().Add(time.Duration(time.Minute)))
	require.Nil(t, err)
	assert.Equal(t, int64(1), count)
}
