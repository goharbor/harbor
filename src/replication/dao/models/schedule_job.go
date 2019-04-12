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

// TODO rename the package name to model

package models

import "time"

// ScheduleJob is the persistent model for the schedule job which is
// used as a scheduler
type ScheduleJob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	JobID        string    `orm:"column(job_id)" json:"job_id"`
	Status       string    `orm:"column(status)" json:"status"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName is required by by beego orm to map the object to the database table
func (s *ScheduleJob) TableName() string {
	return "replication_schedule_job"
}

// ScheduleJobQuery is the query used to list schedule jobs
type ScheduleJobQuery struct {
	PolicyID int64
}
