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

package manager

import (
	"errors"
	"os"
	"testing"
	"time"

	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/goharbor/harbor/src/pkg/task/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type mockTaskDao struct {
	mock.Mock
}

func (m *mockTaskDao) Create(task *dao.Task) (int64, error) {
	args := m.Called(task)
	return int64(args.Int(0)), args.Error(1)
}
func (m *mockTaskDao) Get(id int64) (*dao.Task, error) {
	args := m.Called(id)
	var task *dao.Task
	if args.Get(0) != nil {
		task = args.Get(0).(*dao.Task)
	}
	return task, args.Error(1)
}
func (m *mockTaskDao) Update(task *dao.Task, cols ...string) error {
	args := m.Called(task)
	return args.Error(0)
}
func (m *mockTaskDao) UpdateStatus(id int64, status string, statusCode int,
	statusRevision int64, endTime time.Time) error {
	args := m.Called(id, status, statusCode, statusRevision, endTime)
	return args.Error(0)
}
func (m *mockTaskDao) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *mockTaskDao) CreateCheckInData(data *dao.CheckInData) (int64, error) {
	args := m.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (m *mockTaskDao) UpdateCheckInData(data *dao.CheckInData, cols ...string) error {
	args := m.Called()
	return args.Error(0)
}
func (m *mockTaskDao) ListCheckInData(taskID int64) ([]*dao.CheckInData, error) {
	args := m.Called()
	return args.Get(0).([]*dao.CheckInData), args.Error(1)
}
func (m *mockTaskDao) DeleteCheckInData(id int64) error {
	args := m.Called()
	return args.Error(0)
}
func (m *mockTaskDao) GetGroupStatus(groupID int64) ([]*dao.StatusCount, error) {
	args := m.Called()
	var sgs []*dao.StatusCount
	if args.Get(0) != nil {
		sgs = args.Get(0).([]*dao.StatusCount)
	}
	return sgs, args.Error(1)
}
func (m *mockTaskDao) GetMaxEndTime(groupID int64) (time.Time, error) {
	args := m.Called()
	return args.Get(0).(time.Time), args.Error(1)
}

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

type managerTestingSuite struct {
	suite.Suite
	t                    *testing.T
	assert               *assert.Assertions
	require              *require.Assertions
	mockTaskDao          *mockTaskDao
	mockJobserviceClient *mockJobserviceClient
	mgr                  Manager
}

func (m *managerTestingSuite) SetupSuite() {
	m.t = m.T()
	m.assert = assert.New(m.t)
	m.require = require.New(m.t)

	err := os.Setenv("RUN_MODE", "TEST")
	m.require.Nil(err)
}

func (m *managerTestingSuite) TearDownSuite() {
	err := os.Unsetenv("RUN_MODE")
	m.require.Nil(err)
}

func (m *managerTestingSuite) SetupTest() {
	m.mockTaskDao = &mockTaskDao{}
	m.mgr = &manager{
		dao: m.mockTaskDao,
	}
	m.mockJobserviceClient = &mockJobserviceClient{}
	cjob.GlobalClient = m.mockJobserviceClient
}

func (m *managerTestingSuite) TestCreate() {
	m.mockTaskDao.On("Create", mock.Anything).Return(1, nil)
	_, err := m.mgr.Create(&model.Task{})
	m.mockTaskDao.AssertCalled(m.t, "Create", mock.Anything)
	m.require.Nil(err)
}

// try to get a non-exist task
func (m *managerTestingSuite) TestGetNonExistTask() {
	m.mockTaskDao.On("Get", mock.Anything).Return(nil, errors.New("error"))
	task, err := m.mgr.Get(1)
	m.mockTaskDao.AssertCalled(m.t, "Get", mock.Anything)
	m.require.NotNil(err)
	m.assert.Nil(task)
}

// get the task with check in data
func (m *managerTestingSuite) TestGetWithCheckInData() {
	m.mockTaskDao.On("Get", mock.Anything).Return(&dao.Task{
		ID:    1,
		JobID: "uuid-1",
	}, nil)
	m.mockTaskDao.On("ListCheckInData", mock.Anything).Return([]*dao.CheckInData{
		{
			ID:   1,
			Data: "data",
		},
	}, nil)
	task, err := m.mgr.Get(1)
	m.mockTaskDao.AssertCalled(m.t, "Get", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "ListCheckInData", mock.Anything)
	m.require.Nil(err)
	m.require.NotNil(task)
	m.assert.Equal(int64(1), task.ID)
	m.assert.Equal("uuid-1", task.JobID)
	m.require.Equal(1, len(task.CheckInData))
	m.assert.Equal("data", task.CheckInData[0])
}

func (m *managerTestingSuite) TestUpdate() {
	m.mockTaskDao.On("Update", mock.Anything).Return(nil)
	err := m.mgr.Update(&model.Task{})
	m.mockTaskDao.AssertCalled(m.t, "Update", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestUpdateStatusWithRunning() {
	m.mockTaskDao.On("UpdateStatus", int64(1), job.RunningStatus.String(),
		job.RunningStatus.Code(), int64(1), time.Time{}).Return(nil)
	err := m.mgr.UpdateStatus(1, job.RunningStatus.String(),
		job.RunningStatus.Code(), 1)
	m.mockTaskDao.AssertCalled(m.t, "UpdateStatus", int64(1),
		job.RunningStatus.String(), job.RunningStatus.Code(), int64(1), time.Time{})
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestUpdateStatusWithError() {
	m.mockTaskDao.On("UpdateStatus", int64(1), job.ErrorStatus.String(),
		job.ErrorStatus.Code(), int64(1), mock.Anything).Return(nil)
	err := m.mgr.UpdateStatus(1, job.ErrorStatus.String(),
		job.ErrorStatus.Code(), 1)
	m.mockTaskDao.AssertNotCalled(m.t, "UpdateStatus", int64(1),
		job.ErrorStatus.String(), job.ErrorStatus.Code(), int64(1), time.Time{})
	m.mockTaskDao.AssertCalled(m.t, "UpdateStatus", int64(1),
		job.ErrorStatus.String(), job.ErrorStatus.Code(), int64(1), mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestDeleteWithCheckInData() {
	m.mockTaskDao.On("ListCheckInData", mock.Anything).Return([]*dao.CheckInData{
		{
			ID: 1,
		},
	}, nil)
	m.mockTaskDao.On("DeleteCheckInData", mock.Anything).Return(nil)
	m.mockTaskDao.On("Delete", mock.Anything).Return(nil)
	err := m.mgr.Delete(1)
	m.mockTaskDao.AssertCalled(m.t, "ListCheckInData", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "DeleteCheckInData", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "Delete", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestDeleteWithoutCheckInData() {
	m.mockTaskDao.On("ListCheckInData", mock.Anything).Return([]*dao.CheckInData{}, nil)
	m.mockTaskDao.On("DeleteCheckInData", mock.Anything).Return(nil)
	m.mockTaskDao.On("Delete", mock.Anything).Return(nil)
	err := m.mgr.Delete(1)
	m.mockTaskDao.AssertCalled(m.t, "ListCheckInData", mock.Anything)
	m.mockTaskDao.AssertNotCalled(m.t, "DeleteCheckInData", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "Delete", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestAppendCheckInDataToNonExistTask() {
	m.mockTaskDao.On("Get", mock.Anything).Return(nil, nil)
	err := m.mgr.AppendCheckInData(1, "data")
	m.mockTaskDao.AssertCalled(m.t, "Get", mock.Anything)
	m.require.NotNil(err)
}

func (m *managerTestingSuite) TestAppendCheckInDataWithOptionAppend() {
	m.mockTaskDao.On("Get", mock.Anything).Return(&dao.Task{
		Options: `{"append_check_in_data":true}`,
	}, nil)
	m.mockTaskDao.On("CreateCheckInData", mock.Anything).Return(1, nil)
	m.mockTaskDao.On("ListCheckInData", mock.Anything).Return([]*dao.CheckInData{}, nil)
	m.mockTaskDao.On("UpdateCheckInData", mock.Anything).Return(nil)
	err := m.mgr.AppendCheckInData(1, "data")
	m.mockTaskDao.AssertCalled(m.t, "Get", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "CreateCheckInData", mock.Anything)
	m.mockTaskDao.AssertNotCalled(m.t, "UpdateCheckInData", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestAppendCheckInDataWithOptionOverride() {
	m.mockTaskDao.On("Get", mock.Anything).Return(&dao.Task{}, nil)
	m.mockTaskDao.On("CreateCheckInData", mock.Anything).Return(1, nil)
	m.mockTaskDao.On("ListCheckInData", mock.Anything).Return([]*dao.CheckInData{
		{
			ID: 1,
		},
	}, nil)
	m.mockTaskDao.On("UpdateCheckInData", mock.Anything).Return(nil)
	err := m.mgr.AppendCheckInData(1, "data")
	m.mockTaskDao.AssertCalled(m.t, "Get", mock.Anything)
	m.mockTaskDao.AssertNotCalled(m.t, "CreateCheckInData", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "UpdateCheckInData", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestCalculateTaskGroupWithNoTasks() {
	m.mockTaskDao.On("GetGroupStatus", mock.Anything).Return(nil, nil)
	status, err := m.mgr.CalculateTaskGroupStatus(1)
	m.mockTaskDao.AssertCalled(m.t, "GetGroupStatus", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), status.ID)
	m.assert.Equal(int64(0), status.Total)
	m.assert.True(status.EndTime.IsZero())
	m.assert.Equal(job.SuccessStatus.String(), status.Status)
}

func (m *managerTestingSuite) TestCalculateTaskGroupStatusRunning() {
	m.mockTaskDao.On("GetGroupStatus", mock.Anything).Return([]*dao.StatusCount{
		{
			Status: job.PendingStatus.String(),
			Count:  1,
		},
		{
			Status: job.ScheduledStatus.String(),
			Count:  1,
		},
		{
			Status: job.RunningStatus.String(),
			Count:  1,
		},
		{
			Status: job.ErrorStatus.String(),
			Count:  1,
		},
		{
			Status: job.StoppedStatus.String(),
			Count:  1,
		},
		{
			Status: job.SuccessStatus.String(),
			Count:  1,
		},
	}, nil)
	now := time.Now()
	m.mockTaskDao.On("GetMaxEndTime", mock.Anything).Return(now, nil)
	status, err := m.mgr.CalculateTaskGroupStatus(1)
	m.mockTaskDao.AssertCalled(m.t, "GetGroupStatus", mock.Anything)
	m.mockTaskDao.AssertNotCalled(m.t, "GetMaxEndTime", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(job.RunningStatus.String(), status.Status)
	m.assert.Equal(int64(6), status.Total)
	m.assert.Equal(int64(3), status.Running)
	m.assert.Equal(int64(1), status.Stopped)
	m.assert.Equal(int64(1), status.Error)
	m.assert.Equal(int64(1), status.Success)
	m.assert.True(status.EndTime.IsZero())
}

func (m *managerTestingSuite) TestCalculateTaskGroupStatusStopped() {
	m.mockTaskDao.On("GetGroupStatus", mock.Anything).Return([]*dao.StatusCount{
		{
			Status: job.ErrorStatus.String(),
			Count:  1,
		}, {
			Status: job.StoppedStatus.String(),
			Count:  1,
		},
		{
			Status: job.SuccessStatus.String(),
			Count:  1,
		},
	}, nil)
	now := time.Now()
	m.mockTaskDao.On("GetMaxEndTime", mock.Anything).Return(now, nil)
	status, err := m.mgr.CalculateTaskGroupStatus(1)
	m.mockTaskDao.AssertCalled(m.t, "GetGroupStatus", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "GetMaxEndTime", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(job.StoppedStatus.String(), status.Status)
	m.assert.Equal(int64(3), status.Total)
	m.assert.Equal(int64(0), status.Running)
	m.assert.Equal(int64(1), status.Stopped)
	m.assert.Equal(int64(1), status.Error)
	m.assert.Equal(int64(1), status.Success)
	m.assert.Equal(now, status.EndTime)
}

func (m *managerTestingSuite) TestCalculateTaskGroupStatusError() {
	m.mockTaskDao.On("GetGroupStatus", mock.Anything).Return([]*dao.StatusCount{
		{
			Status: job.ErrorStatus.String(),
			Count:  1,
		},
		{
			Status: job.SuccessStatus.String(),
			Count:  1,
		},
	}, nil)
	now := time.Now()
	m.mockTaskDao.On("GetMaxEndTime", mock.Anything).Return(now, nil)
	status, err := m.mgr.CalculateTaskGroupStatus(1)
	m.mockTaskDao.AssertCalled(m.t, "GetGroupStatus", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "GetMaxEndTime", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(job.ErrorStatus.String(), status.Status)
	m.assert.Equal(int64(2), status.Total)
	m.assert.Equal(int64(0), status.Running)
	m.assert.Equal(int64(0), status.Stopped)
	m.assert.Equal(int64(1), status.Error)
	m.assert.Equal(int64(1), status.Success)
	m.assert.Equal(now, status.EndTime)
}

func (m *managerTestingSuite) TestCalculateTaskGroupStatusSuccess() {
	m.mockTaskDao.On("GetGroupStatus", mock.Anything).Return([]*dao.StatusCount{
		{
			Status: job.SuccessStatus.String(),
			Count:  1,
		},
	}, nil)
	now := time.Now()
	m.mockTaskDao.On("GetMaxEndTime", mock.Anything).Return(now, nil)
	status, err := m.mgr.CalculateTaskGroupStatus(1)
	m.mockTaskDao.AssertCalled(m.t, "GetGroupStatus", mock.Anything)
	m.mockTaskDao.AssertCalled(m.t, "GetMaxEndTime", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(job.SuccessStatus.String(), status.Status)
	m.assert.Equal(int64(1), status.Total)
	m.assert.Equal(int64(0), status.Running)
	m.assert.Equal(int64(0), status.Stopped)
	m.assert.Equal(int64(0), status.Error)
	m.assert.Equal(int64(1), status.Success)
	m.assert.Equal(now, status.EndTime)
}

func TestManagerTestingSuite(t *testing.T) {
	suite.Run(t, &managerTestingSuite{})
}
