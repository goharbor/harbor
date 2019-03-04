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

	"github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/goharbor/harbor/src/replication/ng/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedPolicyManager struct{}

func (f *fakedPolicyManager) Create(*model.Policy) (int64, error) {
	return 0, nil
}
func (f *fakedPolicyManager) List(...*model.PolicyQuery) (int64, []*model.Policy, error) {
	return 0, nil, nil
}
func (f *fakedPolicyManager) Get(int64) (*model.Policy, error) {
	return &model.Policy{
		ID:             1,
		SrcRegistryID:  1,
		SrcNamespaces:  []string{"library"},
		DestRegistryID: 2,
	}, nil
}
func (f *fakedPolicyManager) Update(*model.Policy) error {
	return nil
}
func (f *fakedPolicyManager) Remove(int64) error {
	return nil
}

type fakedRegistryManager struct{}

func (f *fakedRegistryManager) Add(*model.Registry) (int64, error) {
	return 0, nil
}
func (f *fakedRegistryManager) List(...*model.RegistryQuery) (int64, []*model.Registry, error) {
	return 0, nil, nil
}
func (f *fakedRegistryManager) Get(id int64) (*model.Registry, error) {
	if id == 1 {
		return &model.Registry{
			Type: "faked_registry",
		}, nil
	}
	if id == 2 {
		return &model.Registry{
			Type: "faked_registry",
		}, nil
	}
	return nil, nil
}
func (f *fakedRegistryManager) GetByName(name string) (*model.Registry, error) {
	return nil, nil
}
func (f *fakedRegistryManager) Update(*model.Registry, ...string) error {
	return nil
}
func (f *fakedRegistryManager) Remove(int64) error {
	return nil
}
func (f *fakedRegistryManager) HealthCheck() error {
	return nil
}

type fakedExecutionManager struct{}

func (f *fakedExecutionManager) Create(*model.Execution) (int64, error) {
	return 1, nil
}
func (f *fakedExecutionManager) List(...*model.ExecutionQuery) (int64, []*model.Execution, error) {
	return 0, nil, nil
}
func (f *fakedExecutionManager) Get(int64) (*model.Execution, error) {
	return nil, nil
}
func (f *fakedExecutionManager) Update(*model.Execution, ...string) error {
	return nil
}
func (f *fakedExecutionManager) Remove(int64) error {
	return nil
}
func (f *fakedExecutionManager) RemoveAll(int64) error {
	return nil
}
func (f *fakedExecutionManager) CreateTask(*model.Task) (int64, error) {
	return 1, nil
}
func (f *fakedExecutionManager) ListTasks(...*model.TaskQuery) (int64, []*model.Task, error) {
	return 0, nil, nil
}
func (f *fakedExecutionManager) GetTask(int64) (*model.Task, error) {
	return nil, nil
}
func (f *fakedExecutionManager) UpdateTask(*model.Task, ...string) error {
	return nil
}
func (f *fakedExecutionManager) UpdateTaskStatus(int64, string, ...string) error {
	return nil
}
func (f *fakedExecutionManager) RemoveTask(int64) error {
	return nil
}
func (f *fakedExecutionManager) RemoveAllTasks(int64) error {
	return nil
}
func (f *fakedExecutionManager) GetTaskLog(int64) ([]byte, error) {
	return nil, nil
}

type fakedScheduler struct{}

func (f *fakedScheduler) Preprocess(src []*model.Resource, dst []*model.Resource) ([]*scheduler.ScheduleItem, error) {
	items := []*scheduler.ScheduleItem{}
	for i, res := range src {
		items = append(items, &scheduler.ScheduleItem{
			SrcResource: res,
			DstResource: dst[i],
		})
	}
	return items, nil
}
func (f *fakedScheduler) Schedule(items []*scheduler.ScheduleItem) ([]*scheduler.ScheduleResult, error) {
	results := []*scheduler.ScheduleResult{}
	for _, item := range items {
		results = append(results, &scheduler.ScheduleResult{
			TaskID: item.TaskID,
			Error:  nil,
		})
	}
	return results, nil
}
func (f *fakedScheduler) Stop(id string) error {
	return nil
}

func fakedAdapterFactory(*model.Registry) (adapter.Adapter, error) {
	return &fakedAdapter{}, nil
}

type fakedAdapter struct{}

func (f *fakedAdapter) Info() *adapter.Info {
	return nil
}
func (f *fakedAdapter) ListNamespaces(*model.NamespaceQuery) ([]*model.Namespace, error) {
	return nil, nil
}
func (f *fakedAdapter) CreateNamespace(*model.Namespace) error {
	return nil
}
func (f *fakedAdapter) GetNamespace(string) (*model.Namespace, error) {
	return &model.Namespace{}, nil
}
func (f *fakedAdapter) FetchResources(namespace []string, filters []*model.Filter) ([]*model.Resource, error) {
	return []*model.Resource{
		{
			Type: model.ResourceTypeRepository,
			Metadata: &model.ResourceMetadata{
				Name:      "library/hello-world",
				Namespace: "library",
				Vtags:     []string{"latest"},
			},
			Override: false,
		},
	}, nil
}

func TestStartReplication(t *testing.T) {
	err := adapter.RegisterFactory(
		&adapter.Info{
			Type:                   "faked_registry",
			SupportedResourceTypes: []model.ResourceType{"image"},
		}, fakedAdapterFactory)
	require.Nil(t, err)

	controller, _ := NewController(
		&fakedRegistryManager{},
		&fakedExecutionManager{},
		&fakedScheduler{})

	policy := &model.Policy{
		ID:             1,
		SrcRegistryID:  1,
		DestRegistryID: 2,
		DestNamespace:  "library",
	}
	id, err := controller.StartReplication(policy)
	require.Nil(t, err)
	assert.Equal(t, id, int64(1))
}
