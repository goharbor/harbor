//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"strings"
	"time"
)

// ExecHistory execution history
type ExecHistory struct {
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
func (h *ExecHistory) ToSwagger() *models.ExecHistory {
	return &models.ExecHistory{
		ID:            h.ID,
		JobName:       h.Name,
		JobKind:       h.Kind,
		JobParameters: h.Parameters,
		Deleted:       h.Deleted,
		JobStatus:     h.Status,
		Schedule: &models.ScheduleObj{
			// covert MANUAL to Manual because the type of the ScheduleObj
			// must be 'Hourly', 'Daily', 'Weekly', 'Custom', 'Manual' and 'None'
			Type: lib.Title(strings.ToLower(h.Schedule.Type)),
			Cron: h.Schedule.Cron,
		},
		CreationTime: strfmt.DateTime(h.CreationTime),
		UpdateTime:   strfmt.DateTime(h.UpdateTime),
	}
}
