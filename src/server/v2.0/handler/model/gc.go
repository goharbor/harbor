// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/goharbor/harbor/src/common/utils"
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
			Type:              cases.Title(language.English).String(strings.ToLower(h.Schedule.Type)),
			Cron:              h.Schedule.Cron,
			NextScheduledTime: strfmt.DateTime(utils.NextSchedule(h.Schedule.Cron, time.Now())),
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
			Cron:              s.CRON,
			Type:              s.CRONType,
			NextScheduledTime: strfmt.DateTime(utils.NextSchedule(s.CRON, time.Now())),
		},
		CreationTime: strfmt.DateTime(s.CreationTime),
		UpdateTime:   strfmt.DateTime(s.UpdateTime),
	}
}

// NewGCSchedule ...
func NewGCSchedule(s *scheduler.Schedule) *GCSchedule {
	return &GCSchedule{Schedule: s}
}
