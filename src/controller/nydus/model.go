package nydus

import "time"

// Task model for Nydus
type Task struct {
	ID            int64
	ExecutionID   int64
	Status        string
	StatusMessage string
	RunCount      int32
	Repository    string
	Tag           string
	JobID         string
	CreationTime  time.Time
	StartTime     time.Time
	UpdateTime    time.Time
	EndTime       time.Time
}
