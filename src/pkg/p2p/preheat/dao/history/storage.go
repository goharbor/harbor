package history

import (
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
)

// Storage defines the related storing operations for history records.
type Storage interface {
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
