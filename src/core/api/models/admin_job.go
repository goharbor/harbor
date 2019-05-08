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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	common_models "github.com/goharbor/harbor/src/common/models"
	common_utils "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/robfig/cron"
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

// AdminJobReq holds request information for admin job
type AdminJobReq struct {
	AdminJobSchedule
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	ID         int64                  `json:"id"`
	Parameters map[string]interface{} `json:"parameters"`
}

// AdminJobSchedule ...
type AdminJobSchedule struct {
	Schedule *ScheduleParam `json:"schedule"`
}

// ScheduleParam defines the parameter of schedule trigger
type ScheduleParam struct {
	// Daily, Weekly, Custom, Manual, None
	Type string `json:"type"`
	// The cron string of scheduled job
	Cron string `json:"cron"`
}

// AdminJobRep holds the response of query admin job
type AdminJobRep struct {
	AdminJobSchedule
	ID           int64     `json:"id"`
	Name         string    `json:"job_name"`
	Kind         string    `json:"job_kind"`
	Status       string    `json:"job_status"`
	UUID         string    `json:"-"`
	Deleted      bool      `json:"deleted"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// Valid validates the schedule type of a admin job request.
// Only scheduleHourly, ScheduleDaily, ScheduleWeekly, ScheduleCustom, ScheduleManual, ScheduleNone are accepted.
func (ar *AdminJobReq) Valid(v *validation.Validation) {
	if ar.Schedule == nil {
		return
	}
	switch ar.Schedule.Type {
	case ScheduleHourly, ScheduleDaily, ScheduleWeekly, ScheduleCustom:
		if _, err := cron.Parse(ar.Schedule.Cron); err != nil {
			v.SetError("cron", fmt.Sprintf("Invalid schedule trigger parameter cron: %s", ar.Schedule.Cron))
		}
	case ScheduleManual, ScheduleNone:
	default:
		v.SetError("kind", fmt.Sprintf("Invalid schedule kind: %s", ar.Schedule.Type))
	}
}

// ToJob converts request to a job recognized by job service.
func (ar *AdminJobReq) ToJob() *models.JobData {
	metadata := &models.JobMetadata{
		JobKind: ar.JobKind(),
		Cron:    ar.Schedule.Cron,
		// GC job must be unique ...
		IsUnique: true,
	}

	jobData := &models.JobData{
		Name:       ar.Name,
		Parameters: ar.Parameters,
		Metadata:   metadata,
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/adminjob/%d",
			config.InternalCoreURL(), ar.ID),
	}
	return jobData
}

// IsPeriodic ...
func (ar *AdminJobReq) IsPeriodic() bool {
	return ar.JobKind() == job.JobKindPeriodic
}

// JobKind ...
func (ar *AdminJobReq) JobKind() string {
	switch ar.Schedule.Type {
	case ScheduleHourly, ScheduleDaily, ScheduleWeekly, ScheduleCustom:
		return job.JobKindPeriodic
	case ScheduleManual:
		return job.JobKindGeneric
	default:
		return ""
	}
}

// CronString ...
func (ar *AdminJobReq) CronString() string {
	str, err := json.Marshal(ar.Schedule)
	if err != nil {
		log.Debugf("failed to marshal json error, %v", err)
		return ""
	}
	return string(str)
}

// ConvertSchedule converts different kinds of cron string into one standard for UI to show.
// in the latest design, it uses {"type":"Daily","cron":"0 0 0 * * *"} as the cron item.
// As for supporting migration from older version, it needs to convert {"parameter":{"daily_time":0},"type":"daily"}
// and {"type":"Daily","weekday":0,"offtime":57600} into one standard.
func ConvertSchedule(cronStr string) (converted ScheduleParam, err error) {
	if cronStr == "" {
		return ScheduleParam{}, nil
	}

	convertedSchedule := ScheduleParam{}
	convertedSchedule.Type = "custom"

	if strings.Contains(cronStr, "parameter") {
		scheduleModel := common_models.ScanAllPolicy{}
		if err := json.Unmarshal([]byte(cronStr), &scheduleModel); err != nil {
			return ScheduleParam{}, err
		}
		h, m, s := common_utils.ParseOfftime(int64(scheduleModel.Parm["daily_time"].(float64)))
		cron := fmt.Sprintf("%d %d %d * * *", s, m, h)
		convertedSchedule.Cron = cron
		return convertedSchedule, nil
	} else if strings.Contains(cronStr, "offtime") {
		scheduleModel := common_models.ScheduleParam{}
		if err := json.Unmarshal([]byte(cronStr), &scheduleModel); err != nil {
			return ScheduleParam{}, err
		}
		convertedSchedule.Cron = common_models.ParseScheduleParamToCron(&scheduleModel)
		return convertedSchedule, nil
	} else if strings.Contains(cronStr, "cron") {
		scheduleModel := ScheduleParam{}
		if err := json.Unmarshal([]byte(cronStr), &scheduleModel); err != nil {
			return ScheduleParam{}, err
		}
		return scheduleModel, nil
	}

	return ScheduleParam{}, fmt.Errorf("unsupported cron format, %s", cronStr)
}
