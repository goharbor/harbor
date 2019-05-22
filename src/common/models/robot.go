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
	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"time"
)

// RobotTable is the name of table in DB that holds the robot object
const RobotTable = "robot"

// Robot holds the details of a robot.
type Robot struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Name         string    `orm:"column(name)" json:"name"`
	Description  string    `orm:"column(description)" json:"description"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	ExpiresAt    int64     `orm:"column(expiresat)" json:"expires_at"`
	Disabled     bool      `orm:"column(disabled)" json:"disabled"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// RobotQuery ...
type RobotQuery struct {
	Name           string
	ProjectID      int64
	Disabled       bool
	FuzzyMatchName bool
	Pagination
}

// RobotReq ...
type RobotReq struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Disabled    bool           `json:"disabled"`
	Access      []*rbac.Policy `json:"access"`
}

// Valid ...
func (rq *RobotReq) Valid(v *validation.Validation) {
	if utils.IsIllegalLength(rq.Name, 1, 255) {
		v.SetError("name", "robot name with illegal length")
	}
	if utils.IsContainIllegalChar(rq.Name, []string{",", "~", "#", "$", "%"}) {
		v.SetError("name", "robot name contains illegal characters")
	}
}

// RobotRep ...
type RobotRep struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// TableName ...
func (r *Robot) TableName() string {
	return RobotTable
}
