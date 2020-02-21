package history

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/history/mocks"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type historyManagerSuite struct {
	suite.Suite
	manager *mocks.Manager
}

func (hm *historyManagerSuite) SetupTest() {
	hm.manager = new(mocks.Manager)
}

func (hm *historyManagerSuite) TestAppendHistory() {
	hm.manager.On("AppendHistory", mock.Anything).Return(nil)
	err := hm.manager.AppendHistory(&models.HistoryRecord{})
	assert.Nil(hm.T(), err)
}

func (hm *historyManagerSuite) TestUpdateStatus() {
	hm.manager.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := hm.manager.UpdateStatus("", models.TrackStatus(""), "", "")
	assert.Nil(hm.T(), err)
}

func (hm *historyManagerSuite) TestLoadHistories() {
	histories := []*models.HistoryRecord{
		{TaskID: "6c1074e7-d9f7-4e9f-99f7-4bcb38db7374", Image: "library/go"},
		{TaskID: "9a57e251-ee1f-431e-b365-16b738e76ad2", Image: "library/java"},
	}
	hm.manager.On("LoadHistories", mock.Anything).Return(histories, nil)
	hs, err := hm.manager.LoadHistories(nil)
	assert.Nil(hm.T(), err)
	assert.Len(hm.T(), hs, 2)
	assert.Equal(hm.T(), histories, hs)
}

func TestHistoryManager(t *testing.T) {
	suite.Run(t, &historyManagerSuite{})
}
