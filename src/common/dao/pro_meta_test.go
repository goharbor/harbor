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

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProMetaDaoMethods(t *testing.T) {
	name1 := "key1"
	value1 := "value1"
	name2 := "key2"
	value2 := "value2"
	meta1 := &models.ProjectMetadata{
		ProjectID: 1,
		Name:      name1,
		Value:     value1,
	}
	meta2 := &models.ProjectMetadata{
		ProjectID: 1,
		Name:      name2,
		Value:     value2,
	}
	// test add
	require.Nil(t, AddProjectMetadata(meta1))
	defer func() {
		// clean up
		_, err := GetOrmer().Raw(`delete from project_metadata
			where project_id = 1 and name = ?`, name1).Exec()
		require.Nil(t, err)
	}()
	require.Nil(t, AddProjectMetadata(meta2))
	defer func() {
		// clean up
		_, err := GetOrmer().Raw(`delete from project_metadata
			where project_id = 1 and name = ?`, name2).Exec()
		require.Nil(t, err)
	}()
	// test get
	metas, err := GetProjectMetadata(1, name1, name2)
	require.Nil(t, err)
	assert.Equal(t, 2, len(metas))

	m := map[string]*models.ProjectMetadata{}
	for _, meta := range metas {
		m[meta.Name] = meta
	}
	assert.Equal(t, value1, m[name1].Value)
	assert.Equal(t, value2, m[name2].Value)

	// test list
	metas, err = ListProjectMetadata(name1, value1)
	require.Nil(t, err)
	assert.Equal(t, 1, len(metas))
	assert.Equal(t, int64(1), metas[0].ProjectID)

	// test update
	newValue1 := "new_value1"
	meta1.Value = newValue1
	require.Nil(t, UpdateProjectMetadata(meta1))
	metas, err = GetProjectMetadata(1, name1, name2)
	require.Nil(t, err)
	assert.Equal(t, 2, len(metas))

	m = map[string]*models.ProjectMetadata{}
	for _, meta := range metas {
		m[meta.Name] = meta
	}
	assert.Equal(t, newValue1, m[name1].Value)
	assert.Equal(t, value2, m[name2].Value)

	// test delete
	require.Nil(t, DeleteProjectMetadata(1, name1))
	metas, err = GetProjectMetadata(1, name1, name2)
	require.Nil(t, err)
	assert.Equal(t, 1, len(metas))
	assert.Equal(t, value2, metas[0].Value)
}
