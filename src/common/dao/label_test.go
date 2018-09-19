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
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMethodsOfLabel(t *testing.T) {
	labelName := "test"
	label := &models.Label{
		Name:      labelName,
		Level:     common.LabelLevelUser,
		Scope:     common.LabelScopeProject,
		ProjectID: 1,
	}

	// add
	id, err := AddLabel(label)
	require.Nil(t, err)
	label.ID = id

	// add a label which has the same name to another project
	projectID, err := AddProject(models.Project{
		OwnerID: 1,
		Name:    "project_for_label_test",
	})
	require.Nil(t, err)
	defer GetOrmer().QueryTable(&models.Project{}).
		Filter("project_id", projectID).Delete()

	id2, err := AddLabel(&models.Label{
		Name:      labelName,
		Level:     common.LabelLevelUser,
		Scope:     common.LabelScopeProject,
		ProjectID: projectID,
	})
	require.Nil(t, err)
	defer DeleteLabel(id2)

	// get
	l, err := GetLabel(id)
	require.Nil(t, err)
	assert.Equal(t, label.ID, l.ID)
	assert.Equal(t, label.Name, l.Name)
	assert.Equal(t, label.Scope, l.Scope)
	assert.Equal(t, label.ProjectID, l.ProjectID)

	// get total count
	total, err := GetTotalOfLabels(&models.LabelQuery{
		Scope:     common.LabelScopeProject,
		ProjectID: 1,
	})
	require.Nil(t, err)
	assert.Equal(t, int64(1), total)

	// list: exact match
	labels, err := ListLabels(&models.LabelQuery{
		Scope:     common.LabelScopeProject,
		ProjectID: 1,
		Name:      label.Name,
	})
	require.Nil(t, err)
	assert.Equal(t, 1, len(labels))

	// list: fuzzy match
	labels, err = ListLabels(&models.LabelQuery{
		Scope:          common.LabelScopeProject,
		ProjectID:      1,
		Name:           label.Name[:1],
		FuzzyMatchName: true,
	})
	require.Nil(t, err)
	assert.Equal(t, 1, len(labels))

	// list: not exist
	labels, err = ListLabels(&models.LabelQuery{
		Scope:     common.LabelScopeProject,
		ProjectID: 1,
		Name:      label.Name[:1],
	})
	require.Nil(t, err)
	assert.Equal(t, 0, len(labels))

	// update
	newName := "dev"
	label.Name = newName
	err = UpdateLabel(label)
	require.Nil(t, err)

	l, err = GetLabel(id)
	require.Nil(t, err)
	assert.Equal(t, newName, l.Name)

	// delete
	err = DeleteLabel(id)
	require.Nil(t, err)

	l, err = GetLabel(id)
	require.Nil(t, err)
	assert.True(t, l.Deleted)
}
