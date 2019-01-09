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

// RobotTable is the name of table in DB that holds the robot object
const RobotTable = "robot"

// Robot holds the details of a robot.
type Robot struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Name         string    `orm:"column(name)" json:"name"`
	Token        string    `orm:"column(token)" json:"token"`
	Description  string    `orm:"column(description)" json:"description"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	Disabled     bool      `orm:"column(disabled)" json:"disabled"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// RobotQuery ...
type RobotQuery struct {
	Name      string
	ProjectID int64
	Disabled  bool
	Pagination
}

// TableName ...
func (u *Robot) TableName() string {
	return RobotTable
}
