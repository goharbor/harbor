package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/q"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	defaultHistory = &models.HistoryRecord{
		ID:         1,
		TaskID:     "4de7ffb6-43db-452f-88d3-b4dd6e584a62",
		Image:      "project1/go",
		StartTime:  "2020-02-22 10:30:41",
		FinishTime: "2020-02-22 11:30:41",
		Status:     "success",
		Provider:   "dragonfly",
		Instance:   1,
	}
)

type historySuite struct {
	suite.Suite
}

func (hs *historySuite) SetupTest() {
	t := hs.T()
	_, err := AddHistoryRecord(defaultHistory)
	assert.Nil(t, err)
}

func (hs *historySuite) TearDownTest() {
	t := hs.T()
	err := DeleteHistoryRecord(defaultHistory.ID)
	assert.Nil(t, err)
}

func (hs *historySuite) TestGetHistory() {
	t := hs.T()
	h, err := GetHistoryRecord(defaultHistory.ID)
	assert.Nil(t, err)
	// assert.NotNil(t, h)
	assert.Equal(t, defaultHistory.TaskID, h.TaskID)

	// not existed history
	h, err = GetHistoryRecord(0)
	assert.Nil(t, h)
}

func (hs *historySuite) TestGetHistoryByTaskID() {
	t := hs.T()
	h, err := GetHistoryRecordByTaskID("4de7ffb6-43db-452f-88d3-b4dd6e584a62")
	assert.Nil(t, err)
	assert.Equal(t, defaultHistory.Image, h.Image)
}

func (hs *historySuite) TestUpdateHistory() {
	t := hs.T()
	h, err := GetHistoryRecord(defaultHistory.ID)
	assert.Nil(t, err)
	assert.NotNil(t, h)

	h.Status = "fail"
	err = UpdateHistoryRecord(h)
	assert.Nil(t, err)

	h, err = GetHistoryRecord(defaultHistory.ID)
	assert.Nil(t, err)
	assert.Equal(t, "fail", h.Status)
}

func (hs *historySuite) TestListHistories() {
	t := hs.T()
	// add more history
	testHistory1 := &models.HistoryRecord{
		ID:         2,
		TaskID:     "f8482ee7-b658-43f4-9cc2-8bcd2b4d749e",
		Image:      "project1/java",
		StartTime:  "2020-02-23 06:30:41",
		FinishTime: "2020-02-23 11:30:41",
		Status:     "success",
		Provider:   "kraken",
		Instance:   2,
	}
	_, err := AddHistoryRecord(testHistory1)
	assert.Nil(t, err)

	// without queryLimit should return all histories
	total, histories, err := ListHistoryRecords(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, int(total))
	assert.Equal(t, 2, len(histories))

	// limit 1
	total, histories, err = ListHistoryRecords(&q.Query{PageNumber: 1, PageSize: 1})
	assert.Nil(t, err)
	assert.Equal(t, 2, int(total))
	assert.Equal(t, 1, len(histories))
	assert.Equal(t, defaultHistory.ID, histories[0].ID)

	// keyword search
	total, histories, err = ListHistoryRecords(&q.Query{Keywords: map[string]interface{}{"image": &q.FuzzyMatchValue{Value: "java"}}})
	assert.Nil(t, err)
	assert.Equal(t, 1, int(total))
	assert.Equal(t, 1, len(histories))
	assert.Equal(t, testHistory1.Image, histories[0].Image)
	// clean data
	err = DeleteHistoryRecord(testHistory1.ID)
	assert.Nil(t, err)
}

func TestHistory(t *testing.T) {
	suite.Run(t, &historySuite{})
}
