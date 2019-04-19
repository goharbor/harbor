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
)

// WatchItem ...
type WatchItem struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	Namespace    string    `orm:"column(namespace)" json:"namespace"`
	OnDeletion   bool      `orm:"column(on_deletion)" json:"on_deletion"`
	OnPush       bool      `orm:"column(on_push)" json:"on_push"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

//TableName ...
func (w *WatchItem) TableName() string {
	return "replication_immediate_trigger"
}
