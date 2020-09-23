package gc

import (
	"time"
)

// Schedule ...
type Schedule struct {
	Schedule *ScheduleParam `json:"schedule"`
}

// ScheduleParam defines the parameter of schedule trigger
type ScheduleParam struct {
	// Daily, Weekly, Custom, Manual, None
	Type string `json:"type"`
	// The cron string of scheduled job
	Cron string `json:"cron"`
}

// History gc execution history
type History struct {
	Schedule
	ID           int64     `json:"id"`
	Name         string    `json:"job_name"`
	Kind         string    `json:"job_kind"`
	Parameters   string    `json:"job_parameters"`
	Status       string    `json:"job_status"`
	UUID         string    `json:"-"`
	Deleted      bool      `json:"deleted"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}
