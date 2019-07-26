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
	"encoding/json"
	"time"
)

// QuotaUsed a map for the quota used
type QuotaUsed map[string]int64

func (u QuotaUsed) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

// Copy returns copied quota used
func (u QuotaUsed) Copy() QuotaUsed {
	used := QuotaUsed{}
	for key, value := range u {
		used[key] = value
	}

	return used
}

// QuotaUsage model for quota usage
type QuotaUsage struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Reference    string    `orm:"column(reference)" json:"reference"` // The reference type for quota usage, eg: project, user
	ReferenceID  string    `orm:"column(reference_id)" json:"reference_id"`
	Used         string    `orm:"column(used);type(jsonb)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName returns table name for orm
func (qu *QuotaUsage) TableName() string {
	return "quota_usage"
}

// GetUsed returns quota used
func (qu *QuotaUsage) GetUsed() (QuotaUsed, error) {
	var used QuotaUsed
	if err := json.Unmarshal([]byte(qu.Used), &used); err != nil {
		return nil, err
	}

	return used, nil
}

// SetUsed set quota used
func (qu *QuotaUsage) SetUsed(used QuotaUsed) {
	qu.Used = used.String()
}

// QuotaUsageQuery query parameters for quota
type QuotaUsageQuery struct {
	Reference    string
	ReferenceID  string
	ReferenceIDs []string
	Pagination
	Sorting
}
