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
	"github.com/astaxie/beego/validation"
)

// StopJobsReq holds information needed to stop the jobs for a replication rule
type StopJobsReq struct {
	PolicyID int64  `json:"policy_id"`
	Status   string `json:"status"`
}

// Valid ...
func (s *StopJobsReq) Valid(v *validation.Validation) {
	if s.PolicyID <= 0 {
		v.SetError("policy_id", "invalid value")
	}
	if s.Status != "stop" {
		v.SetError("status", "invalid status, valid values: [stop]")
	}
}
