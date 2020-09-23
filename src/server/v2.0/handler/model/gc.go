package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/gc"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

const (
	// ScheduleHourly : 'Hourly'
	ScheduleHourly = "Hourly"
	// ScheduleDaily : 'Daily'
	ScheduleDaily = "Daily"
	// ScheduleWeekly : 'Weekly'
	ScheduleWeekly = "Weekly"
	// ScheduleCustom : 'Custom'
	ScheduleCustom = "Custom"
	// ScheduleManual : 'Manual'
	ScheduleManual = "Manual"
	// ScheduleNone : 'None'
	ScheduleNone = "None"
)

// GCHistory ...
type GCHistory struct {
	*gc.History
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
			Cron: h.Schedule.Schedule.Cron,
			Type: h.Schedule.Schedule.Type,
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
