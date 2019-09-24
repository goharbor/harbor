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

	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(&Task{})
	orm.RegisterModel(&CheckInData{})
}

// Task model definition
type Task struct {
	ID             int64     `orm:"pk;auto;column(id)"`
	GroupID        int64     `orm:"column(group_id)"`
	JobID          string    `orm:"column(job_id)"`
	Status         string    `orm:"column(status)"`
	StatusCode     int       `orm:"column(status_code)"`
	StatusRevision int64     `orm:"column(status_revision)"`
	Message        string    `orm:"column(message)"`
	Options        string    `orm:"column(options)"`
	StartTime      time.Time `orm:"column(start_time)"`
	EndTime        time.Time `orm:"column(end_time)"`
}

// CheckInData records the check in data of tasks
type CheckInData struct {
	ID           int64     `orm:"pk;auto;column(id)"`
	TaskID       int64     `orm:"column(task_id)"`
	Data         string    `orm:"column(data)"`
	CreationTime time.Time `orm:"column(creation_time)"`
	UpdateTime   time.Time `orm:"column(update_time)"`
}

// StatusCount records the count of the specified status tasks
type StatusCount struct {
	Status string `orm:"column(status)"`
	Count  int64  `orm:"column(count)"`
}
