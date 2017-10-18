// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

	"github.com/astaxie/beego/validation"
)

// RepFilterType ...
type RepFilterType struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Name         string    `orm:"column(name)" json:"value"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (r *RepFilterType) TableName() string {
	return "replication_filter_type"
}

// RepFilter ...
type RepFilter struct {
	ID              int64     `orm:"pk;auto;column(id)" json:"id"`
	RepPolicyID     int64     `orm:"column(replication_policy_id)" json:"policy_id"`
	RepFilterTypeID int64     `orm:"column(replication_filter_type_id)" json:"type_id"`
	Value           string    `orm:"column(value)" json:"value"`
	CreationTime    time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime      time.Time `orm:"column(update_time);auto_now" json:"update_time"`
	Deleted         int       `orm:"column(deleted)" json:"deleted"`
}

// TableName ...
func (r *RepFilter) TableName() string {
	return "replication_filter"
}

// Valid ...
func (r *RepFilter) Valid(v *validation.Validation) {
	if r.RepPolicyID <= 0 {
		v.SetError("policy_id", "invalid")
	}

	if r.RepFilterTypeID <= 0 {
		v.SetError("type_id", "invalid")
	}

	if len(r.Value) == 0 {
		v.SetError("value", "can not be empty")
	}
}
