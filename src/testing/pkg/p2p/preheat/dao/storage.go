package dao

import (
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/history"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/instance"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
	"github.com/stretchr/testify/mock"
)

// FakeInstanceStore ...
type FakeInstanceStore struct {
	mock.Mock
}

var _ instance.Manager = (*FakeInstanceStore)(nil)

// Save ...
func (i *FakeInstanceStore) Save(inst *models.Metadata) (int64, error) {
	args := i.Called(inst)
	return 0, args.Error(1)
}

// Delete ...
func (i *FakeInstanceStore) Delete(id int64) error {
	args := i.Called(id)
	return args.Error(0)
}

// Update ...
func (i *FakeInstanceStore) Update(inst *models.Metadata) error {
	args := i.Called(inst)
	return args.Error(0)
}

// Get ...
func (i *FakeInstanceStore) Get(id int64) (*models.Metadata, error) {
	args := i.Called(id)
	var metadata *models.Metadata
	if args.Get(0) != nil {
		metadata = args.Get(0).(*models.Metadata)
	}
	return metadata, args.Error(1)
}

// List ...
func (i *FakeInstanceStore) List(param *models.QueryParam) ([]*models.Metadata, error) {
	args := i.Called(param)
	var metadatas []*models.Metadata
	if args.Get(0) != nil {
		metadatas = args.Get(0).([]*models.Metadata)
	}

	return metadatas, args.Error(1)
}

// FakeHistoryStore ...
type FakeHistoryStore struct {
	mock.Mock
}

var _ history.Manager = (*FakeHistoryStore)(nil)

// AppendHistory ...
func (h *FakeHistoryStore) AppendHistory(record *models.HistoryRecord) error {
	args := h.Called(record)
	return args.Error(0)
}

// UpdateStatus ...
func (h *FakeHistoryStore) UpdateStatus(taskID string, status models.TrackStatus, startTime, endTime string) error {
	args := h.Called(taskID, status, startTime, endTime)
	return args.Error(0)
}

// LoadHistories ...
func (h *FakeHistoryStore) LoadHistories(params *models.QueryParam) ([]*models.HistoryRecord, error) {
	args := h.Called(params)
	var records []*models.HistoryRecord
	if args.Get(0) != nil {
		records = args.Get(0).([]*models.HistoryRecord)
	}

	return records, args.Error(1)
}
