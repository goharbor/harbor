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

package operation

import (
	"errors"
	"io"
	"testing"

	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/operation/flow"
	"github.com/goharbor/harbor/src/replication/operation/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedExecutionManager struct{}

func (f *fakedExecutionManager) Create(*models.Execution) (int64, error) {
	return 1, nil
}
func (f *fakedExecutionManager) List(...*models.ExecutionQuery) (int64, []*models.Execution, error) {
	return 1, []*models.Execution{
		{
			ID: 1,
		},
	}, nil
}
func (f *fakedExecutionManager) Get(int64) (*models.Execution, error) {
	return &models.Execution{
		ID: 1,
	}, nil
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
	return 1, nil
}
func (f *fakedExecutionManager) ListTasks(...*models.TaskQuery) (int64, []*models.Task, error) {
	return 1, []*models.Task{
		{
			ID: 1,
		},
	}, nil
}
func (f *fakedExecutionManager) GetTask(int64) (*models.Task, error) {
	return &models.Task{
		ID: 1,
	}, nil
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
	return []byte("message"), nil
}

type fakedScheduler struct{}

func (f *fakedScheduler) Preprocess(src []*model.Resource, dst []*model.Resource) ([]*scheduler.ScheduleItem, error) {
	items := make([]*scheduler.ScheduleItem, 0)
	for i, res := range src {
		items = append(items, &scheduler.ScheduleItem{
			SrcResource: res,
			DstResource: dst[i],
		})
	}
	return items, nil
}
func (f *fakedScheduler) Schedule(items []*scheduler.ScheduleItem) ([]*scheduler.ScheduleResult, error) {
	results := make([]*scheduler.ScheduleResult, 0)
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

func (f *fakedAdapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeHarbor,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeImage,
			model.ResourceTypeChart,
		},
		SupportedTriggers: []model.TriggerType{model.TriggerTypeManual},
	}, nil
}

func (f *fakedAdapter) PrepareForPush([]*model.Resource) error {
	return nil
}
func (f *fakedAdapter) HealthCheck() (model.HealthStatus, error) {
	return model.Healthy, nil
}
func (f *fakedAdapter) FetchImages(namespace []string, filters []*model.Filter) ([]*model.Resource, error) {
	return []*model.Resource{
		{
			Type: model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
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
func (f *fakedAdapter) DeleteManifest(repository, digest string) error {
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
				Repository: &model.Repository{
					Name: "library/harbor",
				},
				Vtags: []string{"0.2.0"},
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

var ctl = &controller{
	executionMgr: &fakedExecutionManager{},
	scheduler:    &fakedScheduler{},
	flowCtl:      flow.NewController(),
}

func TestStartReplication(t *testing.T) {
	err := adapter.RegisterFactory(model.RegistryTypeHarbor, fakedAdapterFactory)
	require.Nil(t, err)
	config.Config = &config.Configuration{}
	policy := &model.Policy{
		SrcRegistry: &model.Registry{
			Type: model.RegistryTypeHarbor,
		},
		DestRegistry: &model.Registry{
			Type: model.RegistryTypeHarbor,
		},
	}
	resource := &model.Resource{
		Type: model.ResourceTypeImage,
		Metadata: &model.ResourceMetadata{
			Repository: &model.Repository{
				Name: "library/hello-world",
			},
			Vtags: []string{"1.0", "2.0"},
		},
	}
	// policy is disabled
	_, err = ctl.StartReplication(policy, resource, model.TriggerTypeEventBased)
	require.NotNil(t, err)

	policy.Enabled = true
	// the resource contains Vtags whose length isn't 1
	_, err = ctl.StartReplication(policy, resource, model.TriggerTypeEventBased)
	require.NotNil(t, err)

	// replicate resource deletion
	resource.Metadata.Vtags = []string{"1.0"}
	resource.Deleted = true
	id, err := ctl.StartReplication(policy, resource, model.TriggerTypeEventBased)
	require.Nil(t, err)
	assert.Equal(t, int64(1), id)

	// replicate resource copy
	resource.Deleted = false
	id, err = ctl.StartReplication(policy, resource, model.TriggerTypeEventBased)
	require.Nil(t, err)
	assert.Equal(t, int64(1), id)

	// nil resource
	id, err = ctl.StartReplication(policy, nil, model.TriggerTypeEventBased)
	require.Nil(t, err)
	assert.Equal(t, int64(1), id)
}

func TestStopReplication(t *testing.T) {
	err := ctl.StopReplication(1)
	require.Nil(t, err)
}

func TestListExecutions(t *testing.T) {
	n, executions, err := ctl.ListExecutions()
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)
	assert.Equal(t, int64(1), executions[0].ID)
}

func TestGetExecution(t *testing.T) {
	execution, err := ctl.GetExecution(1)
	require.Nil(t, err)
	assert.Equal(t, int64(1), execution.ID)
}

func TestListTasks(t *testing.T) {
	n, tasks, err := ctl.ListTasks()
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)
	assert.Equal(t, int64(1), tasks[0].ID)
}

func TestGetTask(t *testing.T) {
	task, err := ctl.GetTask(1)
	require.Nil(t, err)
	assert.Equal(t, int64(1), task.ID)
}

func TestUpdateTaskStatus(t *testing.T) {
	err := ctl.UpdateTaskStatus(1, "running")
	require.Nil(t, err)
}

func TestGetTaskLog(t *testing.T) {
	log, err := ctl.GetTaskLog(1)
	require.Nil(t, err)
	assert.Equal(t, "message", string(log))
}

func TestIsTaskRunning(t *testing.T) {
	cases := []struct {
		task      *models.Task
		isRunning bool
	}{
		{
			task:      nil,
			isRunning: false,
		},
		{
			task: &models.Task{
				Status: models.TaskStatusSucceed,
			},
			isRunning: false,
		},
		{
			task: &models.Task{
				Status: models.TaskStatusFailed,
			},
			isRunning: false,
		},
		{
			task: &models.Task{
				Status: models.TaskStatusStopped,
			},
			isRunning: false,
		},
		{
			task: &models.Task{
				Status: models.TaskStatusInProgress,
			},
			isRunning: true,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.isRunning, isTaskRunning(c.task))
	}
}

func TestIsNotRunningJobError(t *testing.T) {
	cases := []struct {
		err                  error
		isNotRunningJobError bool
	}{
		{
			err:                  nil,
			isNotRunningJobError: false,
		},
		{
			err:                  errors.New("not the error"),
			isNotRunningJobError: false,
		},
		{
			err:                  errors.New(`[ERROR] [handler.go:253]: Serve http request 'POST /api/v1/jobs/734a11140d939ef700889725' error: 500 {"code":10008,"message":"Stop job failed with error","details":"job '734a11140d939ef700889725' is not a running job"}`),
			isNotRunningJobError: true,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.isNotRunningJobError, isNotRunningJobError(c.err))
	}
}
