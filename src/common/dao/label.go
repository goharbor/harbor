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
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
)

// AddLabel creates a label
func AddLabel(label *models.Label) (int64, error) {
	now := time.Now()
	label.CreationTime = now
	label.UpdateTime = now
	return GetOrmer().Insert(label)
}

// GetLabel specified by ID
func GetLabel(id int64) (*models.Label, error) {
	label := &models.Label{
		ID: id,
	}
	if err := GetOrmer().Read(label); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return label, nil
}

// GetTotalOfLabels returns the total count of labels
func GetTotalOfLabels(query *models.LabelQuery) (int64, error) {
	qs := getLabelQuerySetter(query)
	return qs.Count()
}

// ListLabels list labels according to the query conditions
func ListLabels(query *models.LabelQuery) ([]*models.Label, error) {
	qs := getLabelQuerySetter(query)
	if query.Size > 0 {
		qs = qs.Limit(query.Size)
		if query.Page > 0 {
			qs = qs.Offset((query.Page - 1) * query.Size)
		}
	}
	qs = qs.OrderBy("Name")

	labels := []*models.Label{}
	_, err := qs.All(&labels)
	return labels, err
}

func getLabelQuerySetter(query *models.LabelQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.Label{})
	if len(query.Name) > 0 {
		if query.FuzzyMatchName {
			qs = qs.Filter("Name__icontains", query.Name)
		} else {
			qs = qs.Filter("Name", query.Name)
		}
	}
	if len(query.Level) > 0 {
		qs = qs.Filter("Level", query.Level)
	}
	if len(query.Scope) > 0 {
		qs = qs.Filter("Scope", query.Scope)
	}
	if query.ProjectID != 0 {
		qs = qs.Filter("ProjectID", query.ProjectID)
	}
	qs = qs.Filter("Deleted", false)
	return qs
}

// UpdateLabel ...
func UpdateLabel(label *models.Label) error {
	label.UpdateTime = time.Now()
	_, err := GetOrmer().Update(label)
	return err
}

// DeleteLabel ...
func DeleteLabel(id int64) error {
	label, err := GetLabel(id)
	if err != nil {
		return err
	}
	label.Name = fmt.Sprintf("%s#%d", label.Name, label.ID)
	label.UpdateTime = time.Now()
	label.Deleted = true
	_, err = GetOrmer().Update(label, "Name", "UpdateTime", "Deleted")
	return err
}
