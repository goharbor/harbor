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

package policy

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/replication/models"
)

func TestConvertToPersistModel(t *testing.T) {
	var id, projectID, targetID int64 = 1, 1, 1
	name := "policy01"
	replicateDeletion := true
	trigger := &models.Trigger{
		Kind: "trigger_kind",
	}
	filters := []models.Filter{
		models.Filter{
			Kind:    "filter_kind",
			Pattern: "filter_pattern",
		},
	}
	policy := models.ReplicationPolicy{
		ID:                id,
		Name:              name,
		ReplicateDeletion: replicateDeletion,
		ProjectIDs:        []int64{projectID},
		TargetIDs:         []int64{targetID},
		Trigger:           trigger,
		Filters:           filters,
	}

	ply, err := convertToPersistModel(policy)
	require.Nil(t, err)
	assert.Equal(t, id, ply.ID)
	assert.Equal(t, name, ply.Name)
	assert.Equal(t, replicateDeletion, ply.ReplicateDeletion)
	assert.Equal(t, projectID, ply.ProjectID)
	assert.Equal(t, targetID, ply.TargetID)
	tg, _ := json.Marshal(trigger)
	assert.Equal(t, string(tg), ply.Trigger)
	ft, _ := json.Marshal(filters)
	assert.Equal(t, string(ft), ply.Filters)
}
