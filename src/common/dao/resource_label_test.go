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
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/label/dao"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/label/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMethodsOfResourceLabel(t *testing.T) {
	labelDao := dao.New()
	labelID, err := labelDao.Create(orm.Context(), &model.Label{
		Name:  "test_label",
		Level: common.LabelLevelUser,
		Scope: common.LabelScopeGlobal,
	})
	require.Nil(t, err)
	defer labelDao.Delete(orm.Context(), labelID)

	var resourceID int64 = 1
	resourceType := common.ResourceTypeRepository

	// add
	rl := &models.ResourceLabel{
		LabelID:      labelID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}

	id, err := AddResourceLabel(rl)
	require.Nil(t, err)

	// get
	r, err := GetResourceLabel(resourceType, resourceID, labelID)
	require.Nil(t, err)
	assert.Equal(t, id, r.ID)

	// get by resource
	labels, err := GetLabelsOfResource(resourceType, resourceID)
	require.Nil(t, err)
	require.Equal(t, 1, len(labels))
	assert.Equal(t, id, r.ID)

	// list
	rls, err := ListResourceLabels(&models.ResourceLabelQuery{
		LabelID:      labelID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(rls))
	assert.Equal(t, id, rls[0].ID)

	// delete
	err = DeleteResourceLabel(id)
	require.Nil(t, err)
	labels, err = GetLabelsOfResource(resourceType, resourceID)
	require.Nil(t, err)
	require.Equal(t, 0, len(labels))

	// delete by resource
	id, err = AddResourceLabel(rl)
	require.Nil(t, err)
	err = DeleteLabelsOfResource(resourceType, resourceID)
	require.Nil(t, err)
	labels, err = GetLabelsOfResource(resourceType, resourceID)
	require.Nil(t, err)
	require.Equal(t, 0, len(labels))

	// delete by label ID
	id, err = AddResourceLabel(rl)
	require.Nil(t, err)
	err = DeleteResourceLabelByLabel(labelID)
	require.Nil(t, err)
	rls, err = ListResourceLabels(&models.ResourceLabelQuery{
		LabelID: labelID,
	})
	require.Nil(t, err)
	require.Equal(t, 0, len(rls))
}
