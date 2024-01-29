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

	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Robot ...
type Robot struct {
	*robot.Robot
}

// ToSwagger ...
func (r *Robot) ToSwagger() *models.Robot {
	perms := []*models.RobotPermission{}
	for _, p := range r.Permissions {
		temp := &models.RobotPermission{}
		if err := lib.JSONCopy(temp, p); err != nil {
			log.Warningf("failed to do JSONCopy on RobotPermission, error: %v", err)
		}
		perms = append(perms, temp)
	}

	return &models.Robot{
		ID:           r.ID,
		Name:         r.Name,
		Description:  r.Description,
		ExpiresAt:    r.ExpiresAt,
		Duration:     &r.Duration,
		Level:        r.Level,
		Disable:      r.Disabled,
		Editable:     r.Editable,
		CreationTime: strfmt.DateTime(r.CreationTime),
		UpdateTime:   strfmt.DateTime(r.UpdateTime),
		Permissions:  perms,
	}
}

// NewRobot ...
func NewRobot(r *robot.Robot) *Robot {
	return &Robot{
		Robot: r,
	}
}
