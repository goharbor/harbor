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

package flow

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunOfDeletionFlow(t *testing.T) {
	scheduler := &fakedScheduler{}
	executionMgr := &fakedExecutionManager{}
	policy := &model.Policy{
		SrcRegistry: &model.Registry{
			Type: model.RegistryTypeHarbor,
		},
		DestRegistry: &model.Registry{
			Type: model.RegistryTypeHarbor,
		},
	}
	resources := []*model.Resource{
		{
			Metadata: &model.ResourceMetadata{
				Namespace: &model.Namespace{
					Name: "library",
				},
				Repository: &model.Repository{
					Name: "hello-world",
				},
				Vtags: []string{"latest"},
			},
		},
	}
	flow := NewDeletionFlow(executionMgr, scheduler, 1, policy, resources)
	n, err := flow.Run(nil)
	require.Nil(t, err)
	assert.Equal(t, 1, n)
}
