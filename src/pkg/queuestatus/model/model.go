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

import "time"

// JobQueueStatusTable is the name of table in DB that holds the queue status
const JobQueueStatusTable = "job_queue_status"

// JobQueueStatus ...
type JobQueueStatus struct {
	ID         int       `orm:"pk;auto;column(id)" json:"id,omitempty"`
	JobType    string    `orm:"column(job_type)" json:"job_type,omitempty"`
	Paused     bool      `orm:"column(paused)" json:"paused,omitempty"`
	UpdateTime time.Time `orm:"column(update_time);auto_now"`
}

// TableName ...
func (u *JobQueueStatus) TableName() string {
	return JobQueueStatusTable
}
