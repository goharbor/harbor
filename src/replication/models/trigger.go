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

	"github.com/astaxie/beego/validation"
	"github.com/vmware/harbor/src/replication"
)

//Trigger is replication launching approach definition
type Trigger struct {
	Kind          string         `json:"kind"`           // the type of the trigger
	ScheduleParam *ScheduleParam `json:"schedule_param"` // optional, only used when kind is 'schedule'
}

// Valid ...
func (t *Trigger) Valid(v *validation.Validation) {
	if !(t.Kind == replication.TriggerKindImmediate ||
		t.Kind == replication.TriggerKindManual ||
		t.Kind == replication.TriggerKindSchedule) {
		v.SetError("kind", fmt.Sprintf("invalid trigger kind: %s", t.Kind))
	}

	if t.Kind == replication.TriggerKindSchedule {
		if t.ScheduleParam == nil {
			v.SetError("schedule_param", "empty schedule_param")
		} else {
			t.ScheduleParam.Valid(v)
		}
	}
}

// ScheduleParam defines the parameters used by schedule trigger
type ScheduleParam struct {
	Type    string `json:"type"`    //daily or weekly
	Weekday int8   `json:"weekday"` //Optional, only used when type is 'weekly'
	Offtime int64  `json:"offtime"` //The time offset with the UTC 00:00 in seconds
}

// Valid ...
func (s *ScheduleParam) Valid(v *validation.Validation) {
	if !(s.Type == replication.TriggerScheduleDaily ||
		s.Type == replication.TriggerScheduleWeekly) {
		v.SetError("type", fmt.Sprintf("invalid schedule trigger parameter type: %s", s.Type))
	}

	if s.Type == replication.TriggerScheduleWeekly {
		if s.Weekday < 1 || s.Weekday > 7 {
			v.SetError("weekday", fmt.Sprintf("invalid schedule trigger parameter weekday: %d", s.Weekday))
		}
	}

	if s.Offtime < 0 || s.Offtime > 3600*24 {
		v.SetError("offtime", fmt.Sprintf("invalid schedule trigger parameter offtime: %d", s.Offtime))
	}
}

// Equal ...
func (s *ScheduleParam) Equal(param *ScheduleParam) bool {
	if param == nil {
		return false
	}

	return s.Type == param.Type && s.Weekday == param.Weekday && s.Offtime == param.Offtime
}
