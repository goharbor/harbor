// Copyright 2018 Project Harbor Authors
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

package models

import (
	"time"

	"github.com/astaxie/beego/validation"
	common_models "github.com/goharbor/harbor/src/common/models"
	rep_models "github.com/goharbor/harbor/src/replication/models"
)

// ReplicationPolicy defines the data model used in API level
type ReplicationPolicy struct {
	ID                        int64                      `json:"id"`
	Name                      string                     `json:"name"`
	Description               string                     `json:"description"`
	Filters                   []rep_models.Filter        `json:"filters"`
	ReplicateDeletion         bool                       `json:"replicate_deletion"`
	Trigger                   *rep_models.Trigger        `json:"trigger"`
	Projects                  []*common_models.Project   `json:"projects"`
	Targets                   []*common_models.RepTarget `json:"targets"`
	CreationTime              time.Time                  `json:"creation_time"`
	UpdateTime                time.Time                  `json:"update_time"`
	ReplicateExistingImageNow bool                       `json:"replicate_existing_image_now"`
	ErrorJobCount             int64                      `json:"error_job_count"`
}

// Valid ...
func (r *ReplicationPolicy) Valid(v *validation.Validation) {
	if len(r.Name) == 0 {
		v.SetError("name", "can not be empty")
	}

	if len(r.Name) > 256 {
		v.SetError("name", "max length is 256")
	}

	if len(r.Projects) == 0 {
		v.SetError("projects", "can not be empty")
	}

	if len(r.Targets) == 0 {
		v.SetError("targets", "can not be empty")
	}

	for i := range r.Filters {
		r.Filters[i].Valid(v)
	}

	if r.Trigger == nil {
		v.SetError("trigger", "can not be empty")
	} else {
		r.Trigger.Valid(v)
	}
}
