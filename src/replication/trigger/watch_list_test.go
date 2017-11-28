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

package trigger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/test"
)

func TestMethodsOfWatchList(t *testing.T) {
	dao.DefaultDatabaseWatchItemDAO = &test.FakeWatchItemDAO{}

	var policyID int64 = 1

	// test Add
	item := WatchItem{
		PolicyID:   policyID,
		Namespace:  "library",
		OnDeletion: true,
		OnPush:     false,
	}

	err := DefaultWatchList.Add(item)
	require.Nil(t, err)

	// test Get: non-exist namespace
	items, err := DefaultWatchList.Get("non-exist-namespace", "delete")
	require.Nil(t, err)
	assert.Equal(t, 0, len(items))

	// test Get: non-exist operation
	items, err = DefaultWatchList.Get("library", "non-exist-operation")
	require.Nil(t, err)
	assert.Equal(t, 0, len(items))

	// test Get: valid params
	items, err = DefaultWatchList.Get("library", "delete")
	require.Nil(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, policyID, items[0].PolicyID)

	// test Remove
	err = DefaultWatchList.Remove(policyID)
	require.Nil(t, err)
	items, err = DefaultWatchList.Get("library", "delete")
	require.Nil(t, err)
	assert.Equal(t, 0, len(items))
}
