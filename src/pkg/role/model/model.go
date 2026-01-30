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
	"reflect"

	"github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/lib/errors"

	"github.com/goharbor/harbor/src/lib/log"
)

func init() {
	t := reflect.TypeOf(&Role{})

	log.Debug("*** register role ORM model: " + t.PkgPath() + "." + t.Name())
	orm.Debug = true

	orm.RegisterModel(&Role{})
	log.Debug("*** role model has been registered")

}

// Role holds the details of a robot.
type Role struct {
	ID       int64  `orm:"pk;auto;column(role_id)" json:"id"`
	Name     string `orm:"column(name)" json:"name" sort:"default"`
	RoleMask int64  `orm:"column(role_mask)" json:"role_mask"`
	RoleCode string `orm:"column(role_code)" json:"role_code"`
}

// TableName ...
func (r *Role) TableName() string {
	log.Debug("*** table name queried")
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
