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

package models

import (
	"time"
)

// JobLogTable is the name of the table that record the job execution result.
const JobLogTable = "job_log"

// JobLog holds information about logs which are used to record the result of execution of a job.
type JobLog struct {
	LogID        int       `orm:"pk;auto;column(log_id)" json:"log_id"`
	UUID         string    `orm:"column(job_uuid)" json:"uuid"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	Content      string    `orm:"column(content)" json:"content"`
}

// TableName is required by by beego orm to map JobLog to table job_log
func (a *JobLog) TableName() string {
	return JobLogTable
}
