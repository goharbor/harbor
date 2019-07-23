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

import (
	"time"

	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(
		new(Schedule))
}

// Schedule is a record for a scheduler job
type Schedule struct {
	ID           int64      `orm:"pk;auto;column(id)" json:"id"`
	JobID        string     `orm:"column(job_id)" json:"job_id"`
	Status       string     `orm:"column(status)" json:"status"`
	CreationTime *time.Time `orm:"column(creation_time)" json:"creation_time"`
	UpdateTime   *time.Time `orm:"column(update_time)" json:"update_time"`
}

// ScheduleQuery is query for schedule
type ScheduleQuery struct {
	JobID string
}
