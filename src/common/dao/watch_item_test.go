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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/models"
)

func TestMethodsOfWatchItem(t *testing.T) {
	targetID, err := AddRepTarget(models.RepTarget{
		Name: "test_target_for_watch_item",
		URL:  "http://127.0.0.1",
	})
	require.Nil(t, err)
	defer DeleteRepTarget(targetID)

	policyID, err := AddRepPolicy(models.RepPolicy{
		Name:      "test_policy_for_watch_item",
		ProjectID: 1,
		TargetID:  targetID,
	})
	require.Nil(t, err)
	defer DeleteRepPolicy(policyID)

	item := &models.WatchItem{
		PolicyID:   policyID,
		Namespace:  "library",
		OnPush:     false,
		OnDeletion: true,
	}

	// test Add
	id, err := DefaultDatabaseWatchItemDAO.Add(item)
	require.Nil(t, err)

	// test Get: operation-push
	items, err := DefaultDatabaseWatchItemDAO.Get("library", "push")
	require.Nil(t, err)
	assert.Equal(t, 0, len(items))

	// test Get: operation-delete
	items, err = DefaultDatabaseWatchItemDAO.Get("library", "delete")
	require.Nil(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, id, items[0].ID)
	assert.Equal(t, "library", items[0].Namespace)
	assert.True(t, items[0].OnDeletion)

	// test DeleteByPolicyID
	err = DefaultDatabaseWatchItemDAO.DeleteByPolicyID(policyID)
	require.Nil(t, err)
	items, err = DefaultDatabaseWatchItemDAO.Get("library", "delete")
	require.Nil(t, err)
	assert.Equal(t, 0, len(items))
}
