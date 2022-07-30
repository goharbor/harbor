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
	orm.RegisterModel(&Quota{})
	orm.RegisterModel(&QuotaUsage{})
}

// Quota model for quota
type Quota struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Reference    string    `orm:"column(reference)" json:"reference"` // The reference type for quota, eg: project, user
	ReferenceID  string    `orm:"column(reference_id)" json:"reference_id"`
	Hard         string    `orm:"column(hard);type(jsonb)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
	Version      int64     `orm:"column(version)" json:"-"`
}

// TableName returns table name for orm
func (q *Quota) TableName() string {
	return "quota"
}

// QuotaUsage model for quota usage
type QuotaUsage struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Reference    string    `orm:"column(reference)" json:"reference"` // The reference type for quota usage, eg: project, user
	ReferenceID  string    `orm:"column(reference_id)" json:"reference_id"`
	Used         string    `orm:"column(used);type(jsonb)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
	Version      int64     `orm:"column(version)" json:"-"`
}

// TableName returns table name for orm
func (qu *QuotaUsage) TableName() string {
	return "quota_usage"
}
