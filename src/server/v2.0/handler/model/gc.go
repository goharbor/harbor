package model

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/server/v2.0/models"
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
			// covert MANUAL to Manual because the type of the ScheduleObj
			// must be 'Hourly', 'Daily', 'Weekly', 'Custom', 'Manual' and 'None'
			Type: strings.Title(strings.ToLower(h.Schedule.Type)),
			Cron: h.Schedule.Cron,
		},
		CreationTime: strfmt.DateTime(h.CreationTime),
		UpdateTime:   strfmt.DateTime(h.UpdateTime),
	}
}

// GCSchedule ...
type GCSchedule struct {
	*scheduler.Schedule
}

// ToSwagger converts the schedule to the swagger model
func (s *GCSchedule) ToSwagger() *models.GCHistory {
	if s.Schedule == nil {
		return nil
	}

	e, err := json.Marshal(s.ExtraAttrs)
	if err != nil {
		log.Error(err)
	}

	return &models.GCHistory{
		ID:            s.ID,
		JobName:       "",
		JobKind:       s.CRON,
		JobParameters: string(e),
		Deleted:       false,
		JobStatus:     s.Status,
		Schedule: &models.ScheduleObj{
			Cron: s.CRON,
			Type: s.CRONType,
		},
		CreationTime: strfmt.DateTime(s.CreationTime),
		UpdateTime:   strfmt.DateTime(s.UpdateTime),
	}
}

// NewGCSchedule ...
func NewGCSchedule(s *scheduler.Schedule) *GCSchedule {
	return &GCSchedule{Schedule: s}
}
