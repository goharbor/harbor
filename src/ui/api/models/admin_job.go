// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"fmt"
	"strings"

	"github.com/astaxie/beego/validation"
	"github.com/vmware/harbor/src/common/job"
	"github.com/vmware/harbor/src/common/job/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/ui/config"
)

const (
	//KnowJobGC ...
	KnowJobGC = "gc"
	//ScheduleDaily : 'Daily'
	ScheduleDaily = "Daily"
	//ScheduleWeekly : 'Weekly'
	ScheduleWeekly = "Weekly"
)

// AdminJobReq holds request information for admin job
type AdminJobReq struct {
	//GC or others
	Name string `json:"name"`
	Kind string `json:"kind"`
	//Optional, only used when kind is 'periodic'
	Schedule   *ScheduleParam `json:"schedule"`
	Parameters Parameters     `json:"parameters"`
	Status     string         `json:"status"`
	ID         int64          `json:"id"`
}

//ScheduleParam defines the parameter of schedule trigger
type ScheduleParam struct {
	//Daily or weekly
	Type string `json:"type"`
	//Optional, only used when type is 'weekly'
	Weekday int8 `json:"Weekday"`
	//The time offset with the UTC 00:00 in seconds
	Offtime int64 `json:"Offtime"`
}

//Parameters ...
type Parameters map[string]interface{}

// Valid validates the job type for admin job, so far only GC is supported
func (ajr *AdminJobReq) Valid(v *validation.Validation) {
	jobNameString := strings.ToLower(ajr.Name)

	switch jobNameString {
	case KnowJobGC:
	default:
		v.SetError("type", fmt.Sprintf("Must be one of [%s]", KnowJobGC))
	}

	switch ajr.Kind {
	case job.JobKindGeneric, job.JobKindPeriodic:
	default:
		v.SetError("kind", fmt.Sprintf("Invalid job kind: %s, Must be one of [%s, %s]", ajr.Kind, job.JobKindGeneric, job.JobKindPeriodic))
	}

	if ajr.Kind == job.JobKindPeriodic {
		switch ajr.Schedule.Type {
		case ScheduleDaily, ScheduleWeekly:
		default:
			v.SetError("type", fmt.Sprintf("Invalid job schedule type: %s, Must be one of [%s, %s]", ajr.Schedule.Type, ScheduleDaily, ScheduleWeekly))
		}

		if ajr.Schedule.Offtime < 0 || ajr.Schedule.Offtime > 3600*24 {
			v.SetError("offtime", fmt.Sprintf("Invalid schedule trigger parameter offtime: %d", ajr.Schedule.Offtime))
		}
	}
}

// ToJob converts request to a job reconiged by job service.
func (ajr *AdminJobReq) ToJob() (*models.JobData, error) {
	var jobData *models.JobData
	var err error
	switch strings.ToLower(ajr.Name) {
	case KnowJobGC:
		jobData, err = NewGCJobCreateor().Create(ajr)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Invalid job type %s Must be one of [gc]", ajr.Name)
	}

	return jobData, nil
}

// JobCreateor ...
type JobCreateor interface {
	Create(req *AdminJobReq) (*models.JobData, error)
}

// GCJobCreateor ...
type GCJobCreateor struct{}

// NewGCJobCreateor returns an instance of secretAuthenticator
func NewGCJobCreateor() JobCreateor {
	return &GCJobCreateor{}
}

// Create ...
func (gcjc *GCJobCreateor) Create(req *AdminJobReq) (*models.JobData, error) {
	var isUnique bool
	if req.Kind == job.JobKindGeneric {
		isUnique = true
	}
	metadata := &models.JobMetadata{
		JobKind:  req.Kind,
		IsUnique: isUnique,
	}

	if req.Kind == job.JobKindPeriodic {
		switch req.Schedule.Type {
		case ScheduleDaily:
			h, m, s := utils.ParseOfftime(req.Schedule.Offtime)
			metadata.Cron = fmt.Sprintf("%d %d %d * * *", s, m, h)
		case ScheduleWeekly:
			h, m, s := utils.ParseOfftime(req.Schedule.Offtime)
			metadata.Cron = fmt.Sprintf("%d %d %d * * %d", s, m, h, req.Schedule.Weekday%7)
		default:
			return nil, fmt.Errorf("unsupported schedual trigger type: %s", req.Schedule.Type)
		}
	}

	jobData := &models.JobData{
		Name:     job.ImageGC,
		Metadata: metadata,
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/adminjob",
			config.InternalUIURL()),
	}
	return jobData, nil
}
