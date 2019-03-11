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
	"io"
	"testing"

	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/dao/models"
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

type fakedExecutionManager struct {
	taskID int64
}

func (f *fakedExecutionManager) Create(*models.Execution) (int64, error) {
	return 1, nil
}
func (f *fakedExecutionManager) List(...*models.ExecutionQuery) (int64, []*models.Execution, error) {
	return 0, nil, nil
}
func (f *fakedExecutionManager) Get(int64) (*models.Execution, error) {
	return nil, nil
}
func (f *fakedExecutionManager) Update(*models.Execution, ...string) error {
	return nil
}
func (f *fakedExecutionManager) Remove(int64) error {
	return nil
}
func (f *fakedExecutionManager) RemoveAll(int64) error {
	return nil
}
func (f *fakedExecutionManager) CreateTask(*models.Task) (int64, error) {
	f.taskID++
	id := f.taskID
	return id, nil
}
func (f *fakedExecutionManager) ListTasks(...*models.TaskQuery) (int64, []*models.Task, error) {
	return 0, nil, nil
}
func (f *fakedExecutionManager) GetTask(int64) (*models.Task, error) {
	return nil, nil
}
func (f *fakedExecutionManager) UpdateTask(*models.Task, ...string) error {
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

func (f *fakedAdapter) ListNamespaces(*model.NamespaceQuery) ([]*model.Namespace, error) {
	return nil, nil
}
func (f *fakedAdapter) CreateNamespace(*model.Namespace) error {
	return nil
}
func (f *fakedAdapter) GetNamespace(string) (*model.Namespace, error) {
	return &model.Namespace{}, nil
}
func (f *fakedAdapter) FetchImages(namespace []string, filters []*model.Filter) ([]*model.Resource, error) {
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

func (f *fakedAdapter) ManifestExist(repository, reference string) (exist bool, digest string, err error) {
	return false, "", nil
}
func (f *fakedAdapter) PullManifest(repository, reference string, accepttedMediaTypes []string) (manifest distribution.Manifest, digest string, err error) {
	return nil, "", nil
}
func (f *fakedAdapter) PushManifest(repository, reference, mediaType string, payload []byte) error {
	return nil
}
func (f *fakedAdapter) BlobExist(repository, digest string) (exist bool, err error) {
	return false, nil
}
func (f *fakedAdapter) PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error) {
	return 0, nil, nil
}
func (f *fakedAdapter) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	return nil
}
func (f *fakedAdapter) FetchCharts(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	return []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Name:      "library/harbor",
				Namespace: "library",
				Vtags:     []string{"0.2.0"},
			},
		},
	}, nil
}
func (f *fakedAdapter) ChartExist(name, version string) (bool, error) {
	return false, nil
}
func (f *fakedAdapter) DownloadChart(name, version string) (io.ReadCloser, error) {
	return nil, nil
}
func (f *fakedAdapter) UploadChart(name, version string, chart io.Reader) error {
	return nil
}
func (f *fakedAdapter) DeleteChart(name, version string) error {
	return nil
}

func TestStartReplication(t *testing.T) {
	config.InitWithSettings(nil)
	err := adapter.RegisterFactory(
		&adapter.Info{
			Type: "faked_registry",
			SupportedResourceTypes: []model.ResourceType{
				model.ResourceTypeRepository,
				model.ResourceTypeChart,
			},
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
