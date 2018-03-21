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

// AddResourceLabel add a label to a resource
func AddResourceLabel(rl *models.ResourceLabel) (int64, error) {
	now := time.Now()
	rl.CreationTime = now
	rl.UpdateTime = now
	return GetOrmer().Insert(rl)
}

// GetResourceLabel specified by ID
func GetResourceLabel(rType, rID string, labelID int64) (*models.ResourceLabel, error) {
	rl := &models.ResourceLabel{
		ResourceType: rType,
		ResourceID:   rID,
		LabelID:      labelID,
	}
	if err := GetOrmer().Read(rl, "ResourceType", "ResourceID", "LabelID"); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return rl, nil
}

// GetLabelsOfResource returns the label list of the resource
func GetLabelsOfResource(rType, rID string) ([]*models.Label, error) {
	sql := `select l.id, l.name, l.description, l.color, l.scope, l.project_id, l.creation_time, l.update_time
				from harbor_resource_label rl
				join harbor_label l on rl.label_id=l.id
				where rl.resource_type = ? and rl.resource_id = ?`
	labels := []*models.Label{}
	_, err := GetOrmer().Raw(sql, rType, rID).QueryRows(&labels)
	return labels, err
}

// DeleteResourceLabel ...
func DeleteResourceLabel(id int64) error {
	_, err := GetOrmer().Delete(&models.ResourceLabel{
		ID: id,
	})
	return err
}

// DeleteLabelsOfResource removes all labels of resource specified by rType and rID
func DeleteLabelsOfResource(rType, rID string) error {
	_, err := GetOrmer().QueryTable(&models.ResourceLabel{}).
		Filter("ResourceType", rType).
		Filter("ResourceID", rID).Delete()
	return err
}
