package history

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao"
	daomodels "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
)

// Manager defines the related storing operations for history records.
type Manager interface {
	// Append new preheating history record
	// If succeed, a nil error should be returned.
	AppendHistory(record *models.HistoryRecord) error

	// Update the status of history
	UpdateStatus(taskID string, status models.TrackStatus, startTime, endTime string) error

	// Load history records on top of the query parameters
	// If succeed, a record list will be returned.
	// Otherwise, a non nil error will be set.
	LoadHistories(params *models.QueryParam) ([]*models.HistoryRecord, error)
}

// DefaultManager implement the Manager interface
type DefaultManager struct{}

// NewDefaultManager returns an instance of DefaultManger
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// Ensure *DefaultManager has implemented Manager interface.
var _ Manager = (*DefaultManager)(nil)

var (
	errNilHistoryModel = errors.New("nil history model")
)

// AppendHistory implements @Manager.AppendHistory.
func (dm *DefaultManager) AppendHistory(record *models.HistoryRecord) error {
	if err := validHistoryRecord(record); err != nil {
		return err
	}

	hr, err := convertToDaoModel(record)
	if err != nil {
		return err
	}

	_, err = dao.AddHistoryRecord(hr)
	return err
}

func convertToDaoModel(record *models.HistoryRecord) (*daomodels.HistoryRecord, error) {
	if record == nil {
		return nil, errNilHistoryModel
	}

	hr := &daomodels.HistoryRecord{
		ID:         record.ID,
		TaskID:     record.TaskID,
		Image:      record.Image,
		StartTime:  record.StartTime,
		FinishTime: record.FinishTime,
		Status:     record.Status,
		Provider:   record.Provider,
		Instance:   record.Instance,
	}
	return hr, nil
}

func convertFromDaoModel(record *daomodels.HistoryRecord) (*models.HistoryRecord, error) {
	if record == nil {
		return nil, errNilHistoryModel
	}

	hr := &models.HistoryRecord{
		ID:         record.ID,
		TaskID:     record.TaskID,
		Image:      record.Image,
		StartTime:  record.StartTime,
		FinishTime: record.FinishTime,
		Status:     record.Status,
		Provider:   record.Provider,
		Instance:   record.Instance,
	}
	return hr, nil
}

func convertQueryParams(params *models.QueryParam) *dao.ListHistoryQuery {
	if params != nil {
		return &dao.ListHistoryQuery{
			Page:     params.Page,
			PageSize: params.PageSize,
			Keyword:  params.Keyword,
		}
	}

	return nil
}

// UpdateStatus implements @Manager.UpdateStatus
func (dm *DefaultManager) UpdateStatus(taskID string, status models.TrackStatus, startTime, endTime string) error {
	if len(taskID) == 0 {
		return errors.New("empty task ID of history record")
	}

	if !status.Valid() {
		return fmt.Errorf("invalid status %s", status)
	}

	hr, err := dao.GetHistoryRecordByTaskID(taskID)
	if err != nil {
		return err
	}

	hr.Status = status.String()
	if len(startTime) > 0 {
		hr.StartTime = startTime
	}
	if len(endTime) > 0 {
		hr.FinishTime = endTime
	}

	return dao.UpdateHistoryRecord(hr)
}

// LoadHistories implements @Manager.LoadHistories
func (dm *DefaultManager) LoadHistories(params *models.QueryParam) ([]*models.HistoryRecord, error) {
	hrs, err := dao.ListHistoryRecords(convertQueryParams(params))
	if err != nil {
		return nil, err
	}

	var results []*models.HistoryRecord
	for _, hr := range hrs {
		if h, err := convertFromDaoModel(hr); err == nil {
			results = append(results, h)
		}
	}

	return results, nil
}
