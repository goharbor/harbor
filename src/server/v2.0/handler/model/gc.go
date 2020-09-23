package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"time"
)

// ScheduleParam defines the parameter of schedule trigger
type ScheduleParam struct {
	// Daily, Weekly, Custom, Manual, None
	Type string `json:"type"`
	// The cron string of scheduled job
	Cron string `json:"cron"`
}

// GCHistory gc execution history
type GCHistory struct {
	Schedule     *ScheduleParam `json:"schedule"`
	ID           int64          `json:"id"`
	Name         string         `json:"job_name"`
	Kind         string         `json:"job_kind"`
	Parameters   string         `json:"job_parameters"`
	Status       string         `json:"job_status"`
	UUID         string         `json:"-"`
	Deleted      bool           `json:"deleted"`
	CreationTime time.Time      `json:"creation_time"`
	UpdateTime   time.Time      `json:"update_time"`
}

// ToSwagger converts the history to the swagger model
func (h *GCHistory) ToSwagger() *models.GCHistory {
	return &models.GCHistory{
		ID:            h.ID,
		JobName:       h.Name,
		JobKind:       h.Kind,
		JobParameters: h.Parameters,
		Deleted:       h.Deleted,
		JobStatus:     h.Status,
		Schedule: &models.ScheduleObj{
			Cron: h.Schedule.Cron,
			Type: h.Schedule.Type,
		},
		CreationTime: strfmt.DateTime(h.CreationTime),
		UpdateTime:   strfmt.DateTime(h.UpdateTime),
	}
}

// Schedule ...
type Schedule struct {
	*scheduler.Schedule
}

// ToSwagger converts the schedule to the swagger model
// TODO remove the hard code when after issue https://github.com/goharbor/harbor/issues/13047 is resolved.
func (s *Schedule) ToSwagger() *models.GCHistory {
	return &models.GCHistory{
		ID:            0,
		JobName:       "",
		JobKind:       s.CRON,
		JobParameters: s.Param,
		Deleted:       false,
		JobStatus:     "",
		Schedule: &models.ScheduleObj{
			Cron: s.CRON,
			Type: "Custom",
		},
		CreationTime: strfmt.DateTime(s.CreationTime),
		UpdateTime:   strfmt.DateTime(s.UpdateTime),
	}
}

// NewSchedule ...
func NewSchedule(s *scheduler.Schedule) *Schedule {
	return &Schedule{Schedule: s}
}
