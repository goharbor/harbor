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

	"github.com/beego/beego/orm"
)

func init() {
	orm.RegisterModel(new(Policy))
}

// Policy is the model for replication policy
type Policy struct {
	ID                        int64     `orm:"pk;auto;column(id)"`
	Name                      string    `orm:"column(name)"`
	Description               string    `orm:"column(description)"`
	Creator                   string    `orm:"column(creator)"`
	SrcRegistryID             int64     `orm:"column(src_registry_id)"`
	DestRegistryID            int64     `orm:"column(dest_registry_id)"`
	DestNamespace             string    `orm:"column(dest_namespace)"`
	DestNamespaceReplaceCount int8      `orm:"column(dest_namespace_replace_count)"`
	Override                  bool      `orm:"column(override)"`
	Enabled                   bool      `orm:"column(enabled)"`
	Trigger                   string    `orm:"column(trigger)"`
	Filters                   string    `orm:"column(filters)"`
	ReplicateDeletion         bool      `orm:"column(replicate_deletion)"`
	CreationTime              time.Time `orm:"column(creation_time);auto_now_add" sort:"default:desc"`
	UpdateTime                time.Time `orm:"column(update_time);auto_now"`
	Speed                     int32     `orm:"column(speed_kb)"`
}

// TableName set table name for ORM
func (p *Policy) TableName() string {
	return "replication_policy"
}
