// Copyright 2018 Project Harbor Authors
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

package metamgr

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mgr = NewDefaultProjectMetadataManager()

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	os.Exit(m.Run())
}

func TestMetaMgrMethods(t *testing.T) {
	key := "key"
	value := "value"
	newValue := "new_value"

	// test add
	require.Nil(t, mgr.Add(1, map[string]string{
		key: value,
	}))

	defer func() {
		// clean up
		_, err := dao.GetOrmer().Raw(`delete from project_metadata
		where project_id = 1 and name = ?`, key).Exec()
		require.Nil(t, err)
	}()

	// test get
	m, err := mgr.Get(1, key)
	require.Nil(t, err)
	assert.Equal(t, 1, len(m))
	assert.Equal(t, value, m[key])

	// test list
	metas, err := mgr.List(key, value)
	require.Nil(t, err)
	assert.Equal(t, 1, len(metas))
	assert.Equal(t, int64(1), metas[0].ProjectID)

	// test update
	require.Nil(t, mgr.Update(1, map[string]string{
		key: newValue,
	}))
	m, err = mgr.Get(1, key)
	require.Nil(t, err)
	assert.Equal(t, 1, len(m))
	assert.Equal(t, newValue, m[key])

	// test delete
	require.Nil(t, mgr.Delete(1, key))
	m, err = mgr.Get(1, key)
	require.Nil(t, err)
	assert.Equal(t, 0, len(m))
}
