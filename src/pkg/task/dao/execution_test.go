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

package dao

import (
	"context"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/stretchr/testify/suite"
)

type executionDAOTestSuite struct {
	suite.Suite
	ctx          context.Context
	executionDAO *executionDAO
	taskDao      *taskDAO
	executionID  int64
}

func (e *executionDAOTestSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()
	e.ctx = orm.Context()
	e.taskDao = &taskDAO{}
	e.executionDAO = &executionDAO{
		taskDAO: e.taskDao,
	}
}

func (e *executionDAOTestSuite) SetupTest() {
	id, err := e.executionDAO.Create(e.ctx, &Execution{
		VendorType: "test",
		Trigger:    "test",
		ExtraAttrs: `{"key":"value"}`,
	})
	e.Require().Nil(err)
	e.executionID = id
}

func (e *executionDAOTestSuite) TearDownTest() {
	err := e.executionDAO.Delete(e.ctx, e.executionID)
	e.Nil(err)
}

func (e *executionDAOTestSuite) TestCount() {
	count, err := e.executionDAO.Count(e.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType":     "test",
			"ExtraAttrs.key": "value",
		},
	})
	e.Require().Nil(err)
	e.Equal(int64(1), count)

	count, err = e.executionDAO.Count(e.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType":     "test",
			"ExtraAttrs.key": "incorrect-value",
		},
	})
	e.Require().Nil(err)
	e.Equal(int64(0), count)
}

func (e *executionDAOTestSuite) TestList() {
	executions, err := e.executionDAO.List(e.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType":     "test",
			"ExtraAttrs.key": "value",
		},
	})
	e.Require().Nil(err)
	e.Require().Len(executions, 1)
	e.Equal(e.executionID, executions[0].ID)

	executions, err = e.executionDAO.List(e.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType":     "test",
			"ExtraAttrs.key": "incorrect-value",
		},
	})
	e.Require().Nil(err)
	e.Require().Len(executions, 0)
}

func (e *executionDAOTestSuite) TestGet() {
	// not exist
	_, err := e.executionDAO.Get(e.ctx, 10000)
	e.Require().NotNil(err)
	e.True(errors.IsNotFoundErr(err))

	// exist
	execution, err := e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.NotNil(execution)
}

func (e *executionDAOTestSuite) TestCreate() {
	// happy pass is covered by SetupTest
}

func (e *executionDAOTestSuite) TestUpdate() {
	// not exist
	err := e.executionDAO.Update(e.ctx, &Execution{ID: 10000}, "Status")
	e.Require().NotNil(err)
	e.True(errors.IsNotFoundErr(err))

	// exist
	err = e.executionDAO.Update(e.ctx, &Execution{
		ID:     e.executionID,
		Status: "failed",
	}, "Status")
	e.Require().Nil(err)
	execution, err := e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.Equal("failed", execution.Status)
}

func (e *executionDAOTestSuite) TestDelete() {
	// not exist
	err := e.executionDAO.Delete(e.ctx, 10000)
	e.Require().NotNil(err)
	e.True(errors.IsNotFoundErr(err))

	// happy pass is covered by TearDownTest
}

func (e *executionDAOTestSuite) TestGetMetrics() {
	taskID01, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.SuccessStatus.String(),
		StatusCode:  job.SuccessStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID01)

	taskID02, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.StoppedStatus.String(),
		StatusCode:  job.StoppedStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID02)

	taskID03, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.ErrorStatus.String(),
		StatusCode:  job.ErrorStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID03)

	taskID04, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.PendingStatus.String(),
		StatusCode:  job.PendingStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID04)

	taskID05, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.RunningStatus.String(),
		StatusCode:  job.RunningStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID05)

	taskID06, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.ScheduledStatus.String(),
		StatusCode:  job.ScheduledStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID06)

	metrics, err := e.executionDAO.GetMetrics(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.Equal(int64(6), metrics.TaskCount)
	e.Equal(int64(1), metrics.SuccessTaskCount)
	e.Equal(int64(1), metrics.StoppedTaskCount)
	e.Equal(int64(1), metrics.ErrorTaskCount)
	e.Equal(int64(1), metrics.PendingTaskCount)
	e.Equal(int64(1), metrics.RunningTaskCount)
	e.Equal(int64(1), metrics.ScheduledTaskCount)
}

func (e *executionDAOTestSuite) TestRefreshStatus() {
	// contains tasks with status: success
	taskID01, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.SuccessStatus.String(),
		StatusCode:  job.SuccessStatus.Code(),
		ExtraAttrs:  "{}",
		EndTime:     time.Now(),
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID01)

	statusChanged, currentStatus, err := e.executionDAO.RefreshStatus(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.True(statusChanged)
	e.Equal(job.SuccessStatus.String(), currentStatus)
	execution, err := e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.Equal(job.SuccessStatus.String(), execution.Status)
	e.NotEmpty(execution.EndTime)

	// contains tasks with status: stopped
	taskID02, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.StoppedStatus.String(),
		StatusCode:  job.StoppedStatus.Code(),
		ExtraAttrs:  "{}",
		EndTime:     time.Now(),
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID02)

	statusChanged, currentStatus, err = e.executionDAO.RefreshStatus(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.True(statusChanged)
	e.Equal(job.StoppedStatus.String(), currentStatus)
	execution, err = e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.Equal(job.StoppedStatus.String(), execution.Status)
	e.NotEmpty(execution.EndTime)

	// contains tasks with status: error
	taskID03, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.ErrorStatus.String(),
		StatusCode:  job.ErrorStatus.Code(),
		ExtraAttrs:  "{}",
		EndTime:     time.Now(),
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID03)

	statusChanged, currentStatus, err = e.executionDAO.RefreshStatus(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.True(statusChanged)
	e.Equal(job.ErrorStatus.String(), currentStatus)
	execution, err = e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.Equal(job.ErrorStatus.String(), execution.Status)
	e.NotEmpty(execution.EndTime)

	// contains tasks with status: pending, running, scheduled
	taskID04, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.PendingStatus.String(),
		StatusCode:  job.PendingStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID04)

	taskID05, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.RunningStatus.String(),
		StatusCode:  job.RunningStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID05)

	taskID06, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.ScheduledStatus.String(),
		StatusCode:  job.ScheduledStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID06)

	statusChanged, currentStatus, err = e.executionDAO.RefreshStatus(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.True(statusChanged)
	e.Equal(job.RunningStatus.String(), currentStatus)
	execution, err = e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.Equal(job.RunningStatus.String(), execution.Status)
	e.Empty(execution.EndTime)

	// add another running task, the status shouldn't be changed
	taskID07, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: e.executionID,
		Status:      job.RunningStatus.String(),
		StatusCode:  job.RunningStatus.Code(),
		ExtraAttrs:  "{}",
	})
	e.Require().Nil(err)
	defer e.taskDao.Delete(e.ctx, taskID07)

	statusChanged, currentStatus, err = e.executionDAO.RefreshStatus(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.False(statusChanged)
	e.Equal(job.RunningStatus.String(), currentStatus)
	execution, err = e.executionDAO.Get(e.ctx, e.executionID)
	e.Require().Nil(err)
	e.Equal(job.RunningStatus.String(), execution.Status)
	e.Empty(execution.EndTime)
}

func TestExecutionDAOSuite(t *testing.T) {
	suite.Run(t, &executionDAOTestSuite{})
}

func Test_buildInClauseSqlForExtraAttrs(t *testing.T) {
	type args struct {
		keys []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"extra_attrs.", args{[]string{}}, ""},
		{"extra_attrs.id", args{[]string{"id"}}, "select id from execution where extra_attrs->>?=?"},
		{"extra_attrs.artifact.digest", args{[]string{"artifact", "digest"}}, "select id from execution where extra_attrs->?->>?=?"},
		{"extra_attrs.a.b.c", args{[]string{"a", "b", "c"}}, "select id from execution where extra_attrs->?->?->>?=?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildInClauseSqlForExtraAttrs(tt.args.keys); got != tt.want {
				t.Errorf("buildInClauseSqlForExtraAttrs() = %v, want %v", got, tt.want)
			}
		})
	}
}
