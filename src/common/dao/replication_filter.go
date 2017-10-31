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

package dao

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/src/common/models"
)

// GetRepFilterType returns the replication filter type specified by ID
func GetRepFilterType(id int64) (*models.RepFilterType, error) {
	t := &models.RepFilterType{
		ID: id,
	}
	if err := GetOrmer().Read(t); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

// GetRepFilterTypes returns all replication filter types
func GetRepFilterTypes() ([]*models.RepFilterType, error) {
	types := []*models.RepFilterType{}
	_, err := GetOrmer().QueryTable(&models.RepFilterType{}).All(&types)
	return types, err
}

// GetRepFilterByID returns the replication filter specified by ID
func GetRepFilterByID(id int64) (*models.RepFilter, error) {
	filters := []*models.RepFilter{}
	_, err := GetOrmer().QueryTable(&models.RepFilter{}).
		Filter("id", id).
		Filter("deleted", 0).
		All(&filters)
	if err != nil {
		return nil, err
	}

	if filters == nil || len(filters) == 0 {
		return nil, nil
	}

	return filters[0], nil
}

// GetRepFiltersByPolicyID returns all replication filters defined
// for one replication policy
func GetRepFiltersByPolicyID(id int64) ([]*models.RepFilter, error) {
	filters := []*models.RepFilter{}
	_, err := GetOrmer().QueryTable(&models.RepFilter{}).
		Filter("replication_policy_id", id).
		Filter("deleted", 0).
		All(&filters)
	return filters, err
}

// AddRepFilter adds a new replication filter
func AddRepFilter(filter *models.RepFilter) (int64, error) {
	now := time.Now()
	filter.CreationTime = now
	filter.UpdateTime = now
	filter.Deleted = 0
	return GetOrmer().Insert(filter)
}

// UpdateRepFilter ...
func UpdateRepFilter(filter *models.RepFilter) error {
	filter.UpdateTime = time.Now()
	_, err := GetOrmer().Update(filter, "replication_policy_id",
		"replication_filter_type_id", "value", "update_time")
	return err
}

// DeleteRepFilter ...
func DeleteRepFilter(id int64) error {
	filter := &models.RepFilter{
		ID:         id,
		UpdateTime: time.Now(),
		Deleted:    1,
	}
	_, err := GetOrmer().Update(filter, "update_time", "deleted")
	return err
}
