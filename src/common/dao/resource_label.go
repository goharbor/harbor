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

	"github.com/goharbor/harbor/src/common/models"

	"github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/pkg/label/model"
)

// AddResourceLabel add a label to a resource
func AddResourceLabel(rl *models.ResourceLabel) (int64, error) {
	now := time.Now()
	rl.CreationTime = now
	rl.UpdateTime = now
	return GetOrmer().Insert(rl)
}

// GetResourceLabel specified by resource ID or name
// Get the ResourceLabel by ResourceID if rIDOrName is int
// Get the ResourceLabel by ResourceName if rIDOrName is string
func GetResourceLabel(rType string, rIDOrName interface{}, labelID int64) (*models.ResourceLabel, error) {
	rl := &models.ResourceLabel{
		ResourceType: rType,
		LabelID:      labelID,
	}

	var err error
	id, ok := rIDOrName.(int64)
	if ok {
		rl.ResourceID = id
		err = GetOrmer().Read(rl, "ResourceType", "ResourceID", "LabelID")
	} else {
		rl.ResourceName = rIDOrName.(string)
		err = GetOrmer().Read(rl, "ResourceType", "ResourceName", "LabelID")
	}

	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return rl, nil
}

// GetLabelsOfResource returns the label list of the resource
// Get the labels by ResourceID if rIDOrName is int, or get the labels by ResourceName
func GetLabelsOfResource(rType string, rIDOrName interface{}) ([]*model.Label, error) {
	sql := `select l.id, l.name, l.description, l.color, l.scope, l.project_id, l.creation_time, l.update_time
				from harbor_resource_label rl
				join harbor_label l on rl.label_id=l.id
				where rl.resource_type = ? and`
	if _, ok := rIDOrName.(int64); ok {
		sql += ` rl.resource_id = ?`
	} else {
		sql += ` rl.resource_name = ?`
	}

	labels := []*model.Label{}
	_, err := GetOrmer().Raw(sql, rType, rIDOrName).QueryRows(&labels)
	return labels, err
}

// DeleteResourceLabel ...
func DeleteResourceLabel(id int64) error {
	_, err := GetOrmer().Delete(&models.ResourceLabel{
		ID: id,
	})
	return err
}

// DeleteLabelsOfResource removes all labels of the resource
func DeleteLabelsOfResource(rType string, rIDOrName interface{}) error {
	qs := GetOrmer().QueryTable(&models.ResourceLabel{}).
		Filter("ResourceType", rType)
	if _, ok := rIDOrName.(int64); ok {
		qs = qs.Filter("ResourceID", rIDOrName)
	} else {
		qs = qs.Filter("ResourceName", rIDOrName)
	}
	_, err := qs.Delete()
	return err
}

// ListResourceLabels lists ResourceLabel according to the query conditions
func ListResourceLabels(query ...*models.ResourceLabelQuery) ([]*models.ResourceLabel, error) {
	qs := GetOrmer().QueryTable(&models.ResourceLabel{})
	if len(query) > 0 {
		q := query[0]
		if q.LabelID > 0 {
			qs = qs.Filter("LabelID", q.LabelID)
		}
		if len(q.ResourceType) > 0 {
			qs = qs.Filter("ResourceType", q.ResourceType)
		}
		if q.ResourceID > 0 {
			qs = qs.Filter("ResourceID", q.ResourceID)
		}
		if len(q.ResourceName) > 0 {
			qs = qs.Filter("ResourceName", q.ResourceName)
		}
	}

	rls := []*models.ResourceLabel{}
	_, err := qs.All(&rls)
	return rls, err
}

// DeleteResourceLabelByLabel delete the mapping relationship by label ID
func DeleteResourceLabelByLabel(id int64) error {
	_, err := GetOrmer().QueryTable(&models.ResourceLabel{}).Filter("LabelID", id).Delete()
	return err
}
