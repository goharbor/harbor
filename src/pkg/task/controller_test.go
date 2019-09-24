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

package task

import (
	"errors"
	"os"
	"testing"
	"time"

	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/model"
	"github.com/goharbor/harbor/src/pkg/task/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type mockJobserviceClient struct {
	mock.Mock
}

func (m *mockJobserviceClient) SubmitJob(job *models.JobData) (string, error) {
	args := m.Called(job)
	return args.String(0), args.Error(1)
}
func (m *mockJobserviceClient) GetJobLog(uuid string) ([]byte, error) {
	args := m.Called(uuid)
	return []byte(args.String(0)), args.Error(1)
}
func (m *mockJobserviceClient) PostAction(uuid, action string) error {
	args := m.Called(uuid, action)
	return args.Error(0)
}
func (m *mockJobserviceClient) GetExecutions(uuid string) ([]job.Stats, error) {
	return nil, nil
}

type controllerTestingSuite struct {
	suite.Suite
	t                    *testing.T
	assert               *assert.Assertions
	require              *require.Assertions
	mockTaskManager      *test.MockTaskManager
	mockJobserviceClient *mockJobserviceClient
}

func (m *controllerTestingSuite) SetupSuite() {
	m.t = m.T()
	m.assert = assert.New(m.t)
	m.require = require.New(m.t)

	err := os.Setenv("RUN_MODE", "TEST")
	m.require.Nil(err)
}

func (m *controllerTestingSuite) TearDownSuite() {
	err := os.Unsetenv("RUN_MODE")
	m.require.Nil(err)
}

func (m *controllerTestingSuite) SetupTest() {
	m.mockTaskManager = &test.MockTaskManager{}
	Ctl = &controller{
		mgr: m.mockTaskManager,
	}
	m.mockJobserviceClient = &mockJobserviceClient{}
	cjob.GlobalClient = m.mockJobserviceClient
}

// submit nil job
func (m *controllerTestingSuite) TestSubmitWithNilJob() {
	_, err := Ctl.Submit(nil)
	m.require.NotNil(err)
}

// failed to create database record
func (m *controllerTestingSuite) TestSubmitWithFailureOnCreatingTask() {
	m.mockTaskManager.On("Create", mock.Anything).Return(0, errors.New("error"))
	_, err := Ctl.Submit(&model.Job{})
	m.mockTaskManager.AssertCalled(m.t, "Create", mock.Anything)
	m.require.NotNil(err)
}

// failed to submit job to Jobservice
func (m *controllerTestingSuite) TestSubmitWithFailureOnJobservice() {
	m.mockTaskManager.On("Create", mock.Anything).Return(1, nil)
	m.mockTaskManager.On("Update", mock.MatchedBy(func(task *model.Task) bool {
		if task == nil {
			return false
		}
		if task.Status != job.ErrorStatus.String() {
			return false
		}
		if task.StatusCode != job.ErrorStatus.Code() {
			return false
		}
		if len(task.Message) == 0 {
			return false
		}
		if task.EndTime.IsZero() {
			return false
		}
		return true
	})).Return(nil)
	m.mockJobserviceClient.On("SubmitJob", mock.Anything).Return("", errors.New("error"))
	id, err := Ctl.Submit(&model.Job{})
	m.mockTaskManager.AssertCalled(m.t, "Create", mock.Anything)
	m.mockJobserviceClient.AssertCalled(m.t, "SubmitJob", mock.Anything)
	m.mockTaskManager.AssertCalled(m.t, "Update", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}

// success
func (m *controllerTestingSuite) TestSubmitSuccess() {
	m.mockTaskManager.On("Create", mock.Anything).Return(1, nil)
	m.mockJobserviceClient.On("SubmitJob", mock.Anything).Return("uuid-1", nil)
	m.mockTaskManager.On("Update", mock.MatchedBy(func(task *model.Task) bool {
		if task == nil {
			return false
		}
		if len(task.JobID) == 0 {
			return false
		}
		return true
	})).Return(nil)
	id, err := Ctl.Submit(&model.Job{})
	m.mockTaskManager.AssertCalled(m.t, "Create", mock.Anything)
	m.mockJobserviceClient.AssertCalled(m.t, "SubmitJob", mock.Anything)
	m.mockTaskManager.AssertCalled(m.t, "Update", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}

func (m *controllerTestingSuite) TestGet() {
	m.mockTaskManager.On("Get", mock.Anything).Return(&model.Task{
		ID: 1,
	}, nil)
	task, err := Ctl.Get(int64(1))
	m.mockTaskManager.AssertCalled(m.t, "Get", mock.Anything)
	m.require.Nil(err)
	m.require.NotNil(task)
	m.assert.Equal(int64(1), task.ID)
}

func (m *controllerTestingSuite) TestStopWithNonExistTask() {
	m.mockTaskManager.On("Get", int64(1)).Return(nil, nil)
	err := Ctl.Stop(1)
	m.mockTaskManager.AssertCalled(m.t, "Get", mock.Anything)
	m.require.NotNil(err)
}

func (m *controllerTestingSuite) TestStopWithFinalStatusTask() {
	m.mockTaskManager.On("Get", int64(2)).Return(&model.Task{
		ID:     2,
		Status: job.SuccessStatus.String(),
	}, nil)
	err := Ctl.Stop(2)
	m.mockTaskManager.AssertCalled(m.t, "Get", mock.Anything)
	m.require.Nil(err)
}

// delete a non-existing task
func (m *controllerTestingSuite) TestDeleteWithNonExistTask() {
	m.mockTaskManager.On("Get", mock.Anything).Return(nil, nil)
	err := Ctl.Delete(1)
	m.mockTaskManager.AssertCalled(m.t, "Get", mock.Anything)
	m.require.NotNil(err)
}

// delete a running task
func (m *controllerTestingSuite) TestDeleteWithRunningTask() {
	m.mockTaskManager.On("Get", mock.Anything).Return(&model.Task{
		Status: job.RunningStatus.String(),
	}, nil)
	err := Ctl.Delete(1)
	m.mockTaskManager.AssertCalled(m.t, "Get", mock.Anything)
	m.require.NotNil(err)
}

// delete success
func (m *controllerTestingSuite) TestDeleteSuccess() {
	m.mockTaskManager.On("Get", mock.Anything).Return(&model.Task{
		Status: job.SuccessStatus.String(),
	}, nil)
	m.mockTaskManager.On("Delete", mock.Anything).Return(nil)
	err := Ctl.Delete(1)
	m.mockTaskManager.AssertCalled(m.t, "Get", mock.Anything)
	m.mockTaskManager.AssertCalled(m.t, "Delete", mock.Anything)
	m.require.Nil(err)
}

// get task log
func (m *controllerTestingSuite) TestGetLog() {
	m.mockTaskManager.On("Get", mock.Anything).Return(&model.Task{
		Status: job.SuccessStatus.String(),
	}, nil)
	m.mockJobserviceClient.On("GetJobLog", mock.Anything).Return("task log", nil)
	log, err := Ctl.GetLog(1)
	m.mockTaskManager.AssertCalled(m.t, "Get", mock.Anything)
	m.mockJobserviceClient.AssertCalled(m.t, "GetJobLog", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal([]byte("task log"), log)
}

func (m *controllerTestingSuite) TestCalculateTaskGroup() {
	now := time.Now()
	m.mockTaskManager.On("CalculateTaskGroupStatus", mock.Anything).Return(&model.GroupStatus{
		Status:  job.ErrorStatus.String(),
		EndTime: now,
	}, nil)
	status, err := Ctl.CalculateTaskGroupStatus(1)
	m.mockTaskManager.AssertCalled(m.t, "CalculateTaskGroupStatus", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(now, status.EndTime)
	m.assert.Equal(job.ErrorStatus.String(), status.Status)
}

func TestControllerTestingSuite(t *testing.T) {
	suite.Run(t, &controllerTestingSuite{})
}
