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

package core

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
	"github.com/goharbor/harbor/src/replication/source"
	"github.com/goharbor/harbor/src/replication/target"
	"github.com/goharbor/harbor/src/replication/trigger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	GlobalController = &DefaultController{
		policyManager:  &test.FakePolicyManager{},
		targetManager:  target.NewDefaultManager(),
		sourcer:        source.NewSourcer(),
		triggerManager: trigger.NewManager(0),
	}
	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	assert.Nil(t, GlobalController.Init())
}

func TestCreatePolicy(t *testing.T) {
	_, err := GlobalController.CreatePolicy(models.ReplicationPolicy{
		Trigger: &models.Trigger{
			Kind: replication.TriggerKindManual,
		},
	})
	assert.Nil(t, err)
}

func TestUpdatePolicy(t *testing.T) {
	assert.Nil(t, GlobalController.UpdatePolicy(models.ReplicationPolicy{
		ID: 2,
		Trigger: &models.Trigger{
			Kind: replication.TriggerKindManual,
		},
	}))
}

func TestRemovePolicy(t *testing.T) {
	assert.Nil(t, GlobalController.RemovePolicy(1))
}

func TestGetPolicy(t *testing.T) {
	_, err := GlobalController.GetPolicy(1)
	assert.Nil(t, err)
}

func TestGetPolicies(t *testing.T) {
	_, err := GlobalController.GetPolicies(models.QueryParameter{})
	assert.Nil(t, err)
}

func TestReplicate(t *testing.T) {
	// TODO
}

func TestGetCandidates(t *testing.T) {
	policy := &models.ReplicationPolicy{
		ID: 1,
		Filters: []models.Filter{
			{
				Kind:  replication.FilterItemKindTag,
				Value: "*",
			},
		},
		Trigger: &models.Trigger{
			Kind: replication.TriggerKindImmediate,
		},
	}

	sourcer := source.NewSourcer()

	candidates := []models.FilterItem{
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/hello-world:release-1.0",
		},
		{
			Kind:  replication.FilterItemKindTag,
			Value: "library/hello-world:latest",
		},
	}
	metadata := map[string]interface{}{
		"candidates": candidates,
	}
	result := getCandidates(policy, sourcer, metadata)
	assert.Equal(t, 2, len(result))

	policy.Filters = []models.Filter{
		{
			Kind:  replication.FilterItemKindTag,
			Value: "release-*",
		},
	}
	result = getCandidates(policy, sourcer, metadata)
	assert.Equal(t, 1, len(result))

	// test label filter
	test.InitDatabaseFromEnv()
	policy.Filters = []models.Filter{
		{
			Kind:  replication.FilterItemKindLabel,
			Value: int64(1),
		},
	}
	result = getCandidates(policy, sourcer, metadata)
	assert.Equal(t, 0, len(result))
}

func TestBuildFilterChain(t *testing.T) {
	policy := &models.ReplicationPolicy{
		ID: 1,
		Filters: []models.Filter{
			{
				Kind:  replication.FilterItemKindRepository,
				Value: "*",
			},

			{
				Kind:  replication.FilterItemKindTag,
				Value: "*",
			},

			{
				Kind:  replication.FilterItemKindLabel,
				Value: int64(1),
			},
		},
	}

	sourcer := source.NewSourcer()

	chain := buildFilterChain(policy, sourcer)
	assert.Equal(t, 3, len(chain.Filters()))
}

func TestGetOpUUID(t *testing.T) {
	uuid, err := getOpUUID()
	assert.Nil(t, err)
	assert.NotEmpty(t, uuid)

	uuid, err = getOpUUID(map[string]interface{}{
		"name": "test",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, uuid)

	uuid, err = getOpUUID(map[string]interface{}{
		"op_uuid": 0,
	})
	assert.NotNil(t, err)

	uuid, err = getOpUUID(map[string]interface{}{
		"op_uuid": "0",
	})
	assert.Nil(t, err)
	assert.Equal(t, uuid, "0")
}
