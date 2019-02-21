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
	"testing"

	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedFlowController struct{}

func (f *fakedFlowController) StartReplication(policy *model.Policy) (int64, error) {
	return 1, nil
}
func (f *fakedFlowController) StopReplication(int64) error {
	return nil
}

type fakedExecutionManager struct{}

func (f *fakedExecutionManager) Create(*model.Execution) (int64, error) {
	return 1, nil
}
func (f *fakedExecutionManager) List(...*model.ExecutionQuery) (int64, []*model.Execution, error) {
	return 1, []*model.Execution{
		{
			ID: 1,
		},
	}, nil
}
func (f *fakedExecutionManager) Get(int64) (*model.Execution, error) {
	return &model.Execution{
		ID: 1,
	}, nil
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
	return 1, []*model.Task{
		{
			ID: 1,
		},
	}, nil
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
	return []byte("message"), nil
}

var ctl = NewController(&fakedFlowController{}, &fakedExecutionManager{})

func TestStartReplication(t *testing.T) {
	id, err := ctl.StartReplication(nil)
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

func TestGetTaskLog(t *testing.T) {
	log, err := ctl.GetTaskLog(1)
	require.Nil(t, err)
	assert.Equal(t, "message", string(log))
}
