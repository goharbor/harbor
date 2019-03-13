package models

import (
	"time"
)

// Execution defines the data model used in API level
type Execution struct {
	ID          int64     `json:"id"`
	Status      string    `json:"status"`
	TriggerMode string    `json:"trigger_mode"`
	Duration    int       `json:"duration"`
	SuccessRate string    `json:"success_rate"`
	StartTime   time.Time `json:"start_time"`
}
