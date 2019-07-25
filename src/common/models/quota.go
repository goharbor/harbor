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

	"github.com/goharbor/harbor/src/pkg/types"
)

// QuotaHard a map for the quota hard
type QuotaHard map[string]int64

func (h QuotaHard) String() string {
	bytes, _ := json.Marshal(h)
	return string(bytes)
}

// Copy returns copied quota hard
func (h QuotaHard) Copy() QuotaHard {
	hard := QuotaHard{}
	for key, value := range h {
		hard[key] = value
	}

	return hard
}

// Quota model for quota
type Quota struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Reference    string    `orm:"column(reference)" json:"reference"` // The reference type for quota, eg: project, user
	ReferenceID  string    `orm:"column(reference_id)" json:"reference_id"`
	Hard         string    `orm:"column(hard);type(jsonb)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName returns table name for orm
func (q *Quota) TableName() string {
	return "quota"
}

// GetHard returns quota hard
func (q *Quota) GetHard() (QuotaHard, error) {
	var hard QuotaHard
	if err := json.Unmarshal([]byte(q.Hard), &hard); err != nil {
		return nil, err
	}

	return hard, nil
}

// SetHard set new quota hard
func (q *Quota) SetHard(hard QuotaHard) {
	q.Hard = hard.String()
}

// QuotaQuery query parameters for quota
type QuotaQuery struct {
	ID           int64
	Reference    string
	ReferenceID  string
	ReferenceIDs []string
	Pagination
	Sorting
}

// QuotaUpdateRequest the request for quota update
type QuotaUpdateRequest struct {
	Hard types.ResourceList `json:"hard"`
}
