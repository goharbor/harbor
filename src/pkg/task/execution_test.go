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
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/goharbor/harbor/src/testing/lib/orm"
)

type executionManagerTestSuite struct {
	suite.Suite
	execMgr    *executionManager
	taskMgr    *mockTaskManager
	execDAO    *mockExecutionDAO
	taskDAO    *mockTaskDAO
	ormCreator *orm.Creator
}

func (e *executionManagerTestSuite) SetupTest() {
	e.taskMgr = &mockTaskManager{}
	e.execDAO = &mockExecutionDAO{}
	e.taskDAO = &mockTaskDAO{}
	e.ormCreator = &orm.Creator{}
	e.execMgr = &executionManager{
		executionDAO: e.execDAO,
		taskMgr:      e.taskMgr,
		taskDAO:      e.taskDAO,
	}
}

func (e *executionManagerTestSuite) TestCount() {
	e.execDAO.On("Count", mock.Anything, mock.Anything).Return(int64(10), nil)
	total, err := e.execMgr.Count(nil, &q.Query{})
	e.Require().Nil(err)
	e.Equal(int64(10), total)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestCreate() {
	e.execDAO.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := e.execMgr.Create(nil, "vendor", 0, ExecutionTriggerManual,
		map[string]any{"k": "v"})
	e.Require().Nil(err)
	e.Equal(int64(1), id)
	// sleep to make sure the function in the goroutine run
	time.Sleep(1 * time.Second)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestUpdateExtraAttrs() {
	e.execDAO.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := e.execMgr.UpdateExtraAttrs(nil, 1, map[string]any{"key": "value"})
	e.Require().Nil(err)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestMarkDone() {
	e.execDAO.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := e.execMgr.MarkDone(nil, 1, "success")
	e.Require().Nil(err)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestMarkError() {
	e.execDAO.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := e.execMgr.MarkError(nil, 1, "error")
	e.Require().Nil(err)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestStop() {
	// the execution contains no tasks and the status is final
	e.execDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Execution{
		ID:     1,
		Status: job.SuccessStatus.String(),
	}, nil)
	e.taskDAO.On("List", mock.Anything, mock.Anything).Return(nil, nil)
	err := e.execMgr.Stop(nil, 1)
	e.Require().Nil(err)
	e.taskDAO.AssertExpectations(e.T())
	e.execDAO.AssertExpectations(e.T())

	// reset the mocks
	e.SetupTest()

	// the execution contains no tasks and the status isn't final
	e.execDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Execution{
		ID:     1,
		Status: job.RunningStatus.String(),
	}, nil)
	e.taskDAO.On("List", mock.Anything, mock.Anything).Return(nil, nil)
	e.execDAO.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = e.execMgr.Stop(nil, 1)
	e.Require().Nil(err)
	e.taskDAO.AssertExpectations(e.T())
	e.execDAO.AssertExpectations(e.T())

	// reset the mocks
	e.SetupTest()

	// the execution contains tasks
	e.execDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Execution{
		ID:     1,
		Status: job.RunningStatus.String(),
	}, nil)
	e.taskDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID:          1,
			ExecutionID: 1,
		},
	}, nil)
	e.taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	err = e.execMgr.Stop(nil, 1)
	e.Require().Nil(err)
	e.taskDAO.AssertExpectations(e.T())
	e.execDAO.AssertExpectations(e.T())
	e.taskMgr.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestStopAndWait() {
	// timeout
	e.execDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Execution{
		ID:     1,
		Status: job.RunningStatus.String(),
	}, nil)
	e.taskDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID:          1,
			ExecutionID: 1,
		},
	}, nil)
	e.taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	err := e.execMgr.StopAndWait(nil, 1, 1*time.Second)
	e.Require().NotNil(err)
	e.taskDAO.AssertExpectations(e.T())
	e.execDAO.AssertExpectations(e.T())
	e.taskMgr.AssertExpectations(e.T())

	// reset mocks
	e.SetupTest()

	// pass
	e.execDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Execution{
		ID:     1,
		Status: job.StoppedStatus.String(),
	}, nil)
	e.taskDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID:          1,
			ExecutionID: 1,
		},
	}, nil)
	e.taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	err = e.execMgr.StopAndWait(nil, 1, 1*time.Second)
	e.Require().Nil(err)
	e.taskDAO.AssertExpectations(e.T())
	e.execDAO.AssertExpectations(e.T())
	e.taskMgr.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestDelete() {
	// try to delete the execution which contains running tasks
	e.taskDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.RunningStatus.String(),
		},
	}, nil)
	err := e.execMgr.Delete(nil, 1)
	e.Require().NotNil(err)
	e.True(errors.IsErr(err, errors.PreconditionCode))
	e.taskDAO.AssertExpectations(e.T())

	// reset the mock
	e.SetupTest()

	e.taskDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.SuccessStatus.String(),
		},
	}, nil)
	e.taskDAO.On("Delete", mock.Anything, mock.Anything).Return(nil)
	e.execDAO.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err = e.execMgr.Delete(nil, 1)
	e.Require().Nil(err)
	e.taskDAO.AssertExpectations(e.T())
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestGet() {
	e.execDAO.On("Get", mock.Anything, mock.Anything).Return(&dao.Execution{
		ID:     1,
		Status: job.SuccessStatus.String(),
	}, nil)
	e.execDAO.On("GetMetrics", mock.Anything, mock.Anything).Return(&dao.Metrics{
		TaskCount:        1,
		SuccessTaskCount: 1,
	}, nil)
	exec, err := e.execMgr.Get(nil, 1)
	e.Require().Nil(err)
	e.Equal(int64(1), exec.ID)
	e.Equal(job.SuccessStatus.String(), exec.Status)
	e.Equal(int64(1), exec.Metrics.TaskCount)
	e.Equal(int64(1), exec.Metrics.SuccessTaskCount)
	e.execDAO.AssertExpectations(e.T())
}

func (e *executionManagerTestSuite) TestList() {
	e.execDAO.On("List", mock.Anything, mock.Anything).Return([]*dao.Execution{
		{
			ID:     1,
			Status: job.SuccessStatus.String(),
		},
	}, nil)
	e.execDAO.On("GetMetrics", mock.Anything, mock.Anything).Return(&dao.Metrics{
		TaskCount:        1,
		SuccessTaskCount: 1,
	}, nil)
	execs, err := e.execMgr.List(nil, nil)
	e.Require().Nil(err)
	e.Require().Len(execs, 1)
	e.Equal(int64(1), execs[0].ID)
	e.Equal(job.SuccessStatus.String(), execs[0].Status)
	e.Equal(int64(1), execs[0].Metrics.TaskCount)
	e.Equal(int64(1), execs[0].Metrics.SuccessTaskCount)
	e.execDAO.AssertExpectations(e.T())
}

func TestExecutionManagerSuite(t *testing.T) {
	suite.Run(t, &executionManagerTestSuite{})
}
