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
)

func init() {
	orm.RegisterModel(new(Registry))
}

// Registry is the model for a registry, which wraps the endpoint URL and credential of a remote registry.
type Registry struct {
	ID             int64     `orm:"pk;auto;column(id)"`
	URL            string    `orm:"column(url)"`
	Name           string    `orm:"column(name)"`
	CredentialType string    `orm:"column(credential_type);default(basic)"`
	AccessKey      string    `orm:"column(access_key)"`
	AccessSecret   string    `orm:"column(access_secret)"`
	Type           string    `orm:"column(type)"`
	Insecure       bool      `orm:"column(insecure)"`
	Description    string    `orm:"column(description)"`
	Status         string    `orm:"column(health)"`
	CreationTime   time.Time `orm:"column(creation_time);auto_now_add"`
	UpdateTime     time.Time `orm:"column(update_time);auto_now"`
}

// TableName is required by by beego orm to map Registry to table registry
func (r *Registry) TableName() string {
	return "registry"
}
