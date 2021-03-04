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
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Schedule model
type Schedule struct {
	*scheduler.Schedule
}

// ToSwagger converts the schedule to the swagger model
func (s *Schedule) ToSwagger() *models.Schedule {
	if s.Schedule == nil {
		return nil
	}

	return &models.Schedule{
		ID:     s.ID,
		Status: s.Status,
		Schedule: &models.ScheduleObj{
			Cron: s.CRON,
			Type: s.CRONType,
		},
		Parameters:   s.ExtraAttrs,
		CreationTime: strfmt.DateTime(s.CreationTime),
		UpdateTime:   strfmt.DateTime(s.UpdateTime),
	}
}

// NewSchedule new schedule instance
func NewSchedule(schedule *scheduler.Schedule) *Schedule {
	return &Schedule{Schedule: schedule}
}
