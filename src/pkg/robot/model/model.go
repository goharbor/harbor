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
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/lib/errors"
)

func init() {
	orm.RegisterModel(&Robot{})
}

// Robot holds the details of a robot.
type Robot struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Name         string    `orm:"column(name)" json:"name" sort:"default"`
	Description  string    `orm:"column(description)" json:"description"`
	Secret       string    `orm:"column(secret)" json:"secret"`
	Salt         string    `orm:"column(salt)" json:"-"`
	Duration     int64     `orm:"column(duration)" json:"duration"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	ExpiresAt    int64     `orm:"column(expiresat)" json:"expires_at"`
	Disabled     bool      `orm:"column(disabled)" json:"disabled"`
	Visible      bool      `orm:"column(visible)" json:"-"`
	Creator      string    `orm:"column(creator)" json:"creator"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (r *Robot) TableName() string {
	return "robot"
}

// FromJSON parses robot from json data
func (r *Robot) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), r)
}

// ToJSON marshals Robot to JSON data
func (r *Robot) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
