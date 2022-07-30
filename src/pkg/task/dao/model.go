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

package dao

import (
	"time"

	"github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	orm.RegisterModel(&Execution{})
	orm.RegisterModel(&Task{})
}

// Execution database model
type Execution struct {
	ID         int64  `orm:"pk;auto;column(id)"`
	VendorType string `orm:"column(vendor_type)"`
	VendorID   int64  `orm:"column(vendor_id)"`
	// In most of cases, the status should be calculated from the referenced tasks.
	// When the execution contains no task or failed to create tasks, the status should
	// be set manually
	Status        string    `orm:"column(status)"`
	StatusMessage string    `orm:"column(status_message)"`
	Trigger       string    `orm:"column(trigger)"`
	ExtraAttrs    string    `orm:"column(extra_attrs)"` // json string
	StartTime     time.Time `orm:"column(start_time)" sort:"default:desc"`
	UpdateTime    time.Time `orm:"column(update_time)"`
	EndTime       time.Time `orm:"column(end_time)"`
	Revision      int64     `orm:"column(revision)"`
}

// Metrics is the task metrics for one execution
type Metrics struct {
	TaskCount          int64 `json:"task_count"`
	SuccessTaskCount   int64 `json:"success_task_count"`
	ErrorTaskCount     int64 `json:"error_task_count"`
	PendingTaskCount   int64 `json:"pending_task_count"`
	RunningTaskCount   int64 `json:"running_task_count"`
	ScheduledTaskCount int64 `json:"scheduled_task_count"`
	StoppedTaskCount   int64 `json:"stopped_task_count"`
}

// Task database model
type Task struct {
	ID             int64     `orm:"pk;auto;column(id)"`
	VendorType     string    `orm:"column(vendor_type)"`
	ExecutionID    int64     `orm:"column(execution_id)"`
	JobID          string    `orm:"column(job_id)"`
	Status         string    `orm:"column(status)"`
	StatusCode     int       `orm:"column(status_code)"`
	StatusRevision int64     `orm:"column(status_revision)"`
	StatusMessage  string    `orm:"column(status_message)"`
	RunCount       int32     `orm:"column(run_count)"`
	ExtraAttrs     string    `orm:"column(extra_attrs)"` // json string
	CreationTime   time.Time `orm:"column(creation_time)"`
	StartTime      time.Time `orm:"column(start_time)"`
	UpdateTime     time.Time `orm:"column(update_time)"`
	EndTime        time.Time `orm:"column(end_time)"`
}

// GetDefaultSorts specifies the default sorts
func (t *Task) GetDefaultSorts() []*q.Sort {
	// sort by ID to fix https://github.com/goharbor/harbor/issues/14433
	return []*q.Sort{
		{
			Key:  "StartTime",
			DESC: true,
		},
		{
			Key:  "ID",
			DESC: true,
		},
	}
}

// StatusCount model
type StatusCount struct {
	Status string `orm:"column(status)"`
	Count  int64  `orm:"column(count)"`
}
