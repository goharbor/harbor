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
	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/redis"
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
	// initializes cache for testing
	err := cache.Initialize(cache.Redis, "redis://localhost:6379/0")
	e.NoError(err)
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

func (e *executionDAOTestSuite) TestAsyncRefreshStatus() {
	err := e.executionDAO.AsyncRefreshStatus(e.ctx, e.executionID, "GC")
	e.NoError(err)
	defer cache.Default().Delete(e.ctx, buildExecStatusOutdateKey(e.executionID, "GC"))
	e.True(cache.Default().Contains(e.ctx, buildExecStatusOutdateKey(e.executionID, "GC")))
}

func (e *executionDAOTestSuite) TestScanAndRefreshOutdateStatus() {
	// create execution1 with 1 running task
	id1, err := e.executionDAO.Create(e.ctx, &Execution{
		VendorType: "test1",
		Trigger:    "test",
		ExtraAttrs: `{"key":"value"}`,
	})
	e.NoError(err)
	defer e.executionDAO.Delete(e.ctx, id1)

	tid1, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: id1,
		Status:      job.RunningStatus.String(),
		StatusCode:  job.RunningStatus.Code(),
		ExtraAttrs:  `{}`,
	})
	e.NoError(err)
	defer e.taskDao.Delete(e.ctx, tid1)

	// create execution1 with 1 error task
	id2, err := e.executionDAO.Create(e.ctx, &Execution{
		VendorType: "test2",
		Trigger:    "test",
		ExtraAttrs: `{"key":"value"}`,
	})
	e.NoError(err)
	defer e.executionDAO.Delete(e.ctx, id2)

	tid2, err := e.taskDao.Create(e.ctx, &Task{
		ExecutionID: id2,
		Status:      job.ErrorStatus.String(),
		StatusCode:  job.ErrorStatus.Code(),
		ExtraAttrs:  `{}`,
	})
	e.NoError(err)
	defer e.taskDao.Delete(e.ctx, tid2)

	// async refresh the status
	err = e.executionDAO.AsyncRefreshStatus(e.ctx, id1, "GC")
	e.NoError(err)
	err = e.executionDAO.AsyncRefreshStatus(e.ctx, id2, "GC")
	e.NoError(err)
	// test scan out and refresh
	scanAndRefreshOutdateStatus(e.ctx)
	exec1, err := e.executionDAO.Get(e.ctx, id1)
	e.NoError(err)
	e.Equal(job.RunningStatus.String(), exec1.Status)
	exec2, err := e.executionDAO.Get(e.ctx, id2)
	e.NoError(err)
	e.Equal(job.ErrorStatus.String(), exec2.Status)
}

func TestExecutionDAOSuite(t *testing.T) {
	suite.Run(t, &executionDAOTestSuite{})
}

func Test_buildInClauseSQLForExtraAttrs(t *testing.T) {
	tests := []struct {
		name string
		args []jsonbStru
		want string
	}{
		{"extra_attrs.", []jsonbStru{}, ""},
		{"extra_attrs.", []jsonbStru{{}}, ""},
		{"extra_attrs.id", []jsonbStru{{
			keyPrefix: "extra_attrs.",
			key:       "extra_attrs.id",
			value:     "1",
		}}, "select id from execution where extra_attrs->>?=?"},
		{"extra_attrs.artifact.digest", []jsonbStru{{
			keyPrefix: "extra_attrs.",
			key:       "extra_attrs.artifact.digest",
			value:     "sha256:1234",
		}}, "select id from execution where extra_attrs->?->>?=?"},
		{"extra_attrs.a.b.c", []jsonbStru{{
			keyPrefix: "extra_attrs.",
			key:       "extra_attrs.a.b.c",
			value:     "test_value_1",
		}, {
			keyPrefix: "extra_attrs.",
			key:       "extra_attrs.d.e",
			value:     "test_value_2",
		}}, "select id from execution where extra_attrs->?->?->>?=? and extra_attrs->?->>?=?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := buildInClauseSQLForExtraAttrs(tt.args); got != tt.want {
				t.Errorf("buildInClauseSQLForExtraAttrs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractExecIDVendorFromKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name       string
		args       args
		wantID     int64
		wantVendor string
		wantErr    bool
	}{
		{"invalid format", args{"invalid:foo:bar"}, 0, "", true},
		{"invalid execution id", args{"execution:id:12abc:vendor:GC:status_outdate"}, 0, "", true},
		{"invalid vendor type", args{"execution:id:100:vendor:foo:status_outdate"}, 0, "", true},
		{"valid 1", args{"execution:id:100:vendor:GARBAGE_COLLECTION:status_outdate"}, 100, "GARBAGE_COLLECTION", false},
		{"valid 2", args{"execution:id:100:vendor:P2P_PREHEAT:status_outdate"}, 100, "P2P_PREHEAT", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := extractExecIDVendorFromKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractExecIDVendorFromKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantID {
				t.Errorf("extractExecIDVendorFromKey() got = %v, want %v", got, tt.wantID)
			}
			if got1 != tt.wantVendor {
				t.Errorf("extractExecIDVendorFromKey() got1 = %v, want %v", got1, tt.wantVendor)
			}
		})
	}
}
