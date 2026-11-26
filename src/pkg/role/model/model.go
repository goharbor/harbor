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
	orm.RegisterModel(&Role{})
}

// Role holds the details of a role.
type Role struct {
	ID          int64     `orm:"pk;auto;column(role_id)" json:"id"`
	Name        string    `orm:"column(name)" json:"name" sort:"default"`
	RoleMask    int64     `orm:"column(role_mask)" json:"role_mask"`
	RoleCode    string    `orm:"column(role_code)" json:"role_code"`
	IsBuiltin   bool      `orm:"column(is_builtin)" json:"is_builtin"`
	Description string    `orm:"column(description)" json:"description"`
	Modified    bool      `orm:"column(modified)" json:"modified"`
	CreatedBy   string    `orm:"column(created_by)" json:"created_by"`
	CreatedAt   time.Time `orm:"column(created_at);auto_now_add;type(datetime)" json:"created_at"`
	ModifiedBy  string    `orm:"column(modified_by)" json:"modified_by"`
	ModifiedAt  time.Time `orm:"column(modified_at);auto_now;type(datetime)" json:"modified_at"`
}

// TableName ...
func (r *Role) TableName() string {
	return "role"
}

// FromJSON parses role from json data
func (r *Role) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), r)
}

// ToJSON marshals role to JSON data
func (r *Role) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
