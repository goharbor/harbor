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

package replication

import (
	"context"
	"fmt"
	"testing"
	"time"

	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	testingreg "github.com/goharbor/harbor/src/testing/pkg/reg"
	testingrep "github.com/goharbor/harbor/src/testing/pkg/replication"
	testingscheduler "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
)

type replicationTestSuite struct {
	suite.Suite
	ctl        *controller
	repMgr     *testingrep.Manager
	regMgr     *testingreg.Manager
	execMgr    *testingTask.ExecutionManager
	taskMgr    *testingTask.Manager
	scheduler  *testingscheduler.Scheduler
	flowCtl    *flowController
	ormCreator *orm.Creator
}

func (r *replicationTestSuite) SetupTest() {
	r.repMgr = &testingrep.Manager{}
	r.regMgr = &testingreg.Manager{}
	r.execMgr = &testingTask.ExecutionManager{}
	r.taskMgr = &testingTask.Manager{}
	r.scheduler = &testingscheduler.Scheduler{}
	r.flowCtl = &flowController{}
	r.ormCreator = &orm.Creator{}
	r.ctl = &controller{
		repMgr:     r.repMgr,
		regMgr:     r.regMgr,
		scheduler:  r.scheduler,
		execMgr:    r.execMgr,
		taskMgr:    r.taskMgr,
		flowCtl:    r.flowCtl,
		ormCreator: r.ormCreator,
		wp:         lib.NewWorkerPool(1024),
	}
}

func (r *replicationTestSuite) TestStart() {
	// policy is disabled
	id, err := r.ctl.Start(context.Background(), &repctlmodel.Policy{Enabled: false}, nil, task.ExecutionTriggerManual)
	r.Require().NotNil(err)

	// got error when running the replication flow
	r.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	r.execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{}, nil)
	r.execMgr.On("StopAndWait", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	r.execMgr.On("MarkError", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	r.flowCtl.On("Start", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	r.ormCreator.On("Create").Return(nil)
	id, err = r.ctl.Start(context.Background(), &repctlmodel.Policy{Enabled: true}, nil, task.ExecutionTriggerManual)
	r.Require().Nil(err)
	r.Equal(int64(1), id)
	time.Sleep(1 * time.Second) // wait the functions called in the goroutine
	r.execMgr.AssertExpectations(r.T())
	r.flowCtl.AssertExpectations(r.T())
	r.ormCreator.AssertExpectations(r.T())

	// reset the mocks
	r.SetupTest()

	// got no error when running the replication flow
	r.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	r.execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{}, nil)
	r.flowCtl.On("Start", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	r.ormCreator.On("Create").Return(nil)
	id, err = r.ctl.Start(context.Background(), &repctlmodel.Policy{Enabled: true}, nil, task.ExecutionTriggerManual)
	r.Require().Nil(err)
	r.Equal(int64(1), id)
	time.Sleep(1 * time.Second) // wait the functions called in the goroutine
	r.execMgr.AssertExpectations(r.T())
	r.flowCtl.AssertExpectations(r.T())
	r.ormCreator.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestStop() {
	r.execMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	err := r.ctl.Stop(nil, 1)
	r.Require().Nil(err)
	r.execMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestExecutionCount() {
	r.execMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	total, err := r.ctl.ExecutionCount(nil, nil)
	r.Require().Nil(err)
	r.Equal(int64(1), total)
	r.execMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestListExecutions() {
	r.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID:         1,
			VendorType: job.Replication,
			VendorID:   1,
			Status:     job.RunningStatus.String(),
			Metrics: &dao.Metrics{
				TaskCount:        1,
				RunningTaskCount: 1,
			},
			Trigger:   task.ExecutionTriggerManual,
			StartTime: time.Time{},
			EndTime:   time.Time{},
		},
	}, nil)
	executions, err := r.ctl.ListExecutions(nil, nil)
	r.Require().Nil(err)
	r.Require().Len(executions, 1)
	r.Equal(int64(1), executions[0].ID)
	r.Equal(int64(1), executions[0].PolicyID)
	r.execMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestGetExecution() {
	r.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID:         1,
			VendorType: job.Replication,
			VendorID:   1,
			Status:     job.RunningStatus.String(),
			Metrics: &dao.Metrics{
				TaskCount:        1,
				RunningTaskCount: 1,
			},
			Trigger:   task.ExecutionTriggerManual,
			StartTime: time.Time{},
			EndTime:   time.Time{},
		},
	}, nil)
	execution, err := r.ctl.GetExecution(nil, 1)
	r.Require().Nil(err)
	r.Equal(int64(1), execution.ID)
	r.Equal(int64(1), execution.PolicyID)
	r.execMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestTaskCount() {
	r.taskMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	total, err := r.ctl.TaskCount(nil, nil)
	r.Require().Nil(err)
	r.Equal(int64(1), total)
	r.taskMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestListTasks() {
	r.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.RunningStatus.String(),
			ExtraAttrs: map[string]interface{}{
				"resource_type":        "artifact",
				"source_resource":      "library/hello-world",
				"destination_resource": "library/hello-world",
				"operation":            "copy",
			},
		},
	}, nil)
	tasks, err := r.ctl.ListTasks(nil, nil)
	r.Require().Nil(err)
	r.Require().Len(tasks, 1)
	r.Equal(int64(1), tasks[0].ID)
	r.Equal(int64(1), tasks[0].ExecutionID)
	r.Equal("artifact", tasks[0].ResourceType)
	r.Equal("library/hello-world", tasks[0].SourceResource)
	r.Equal("library/hello-world", tasks[0].DestinationResource)
	r.Equal("copy", tasks[0].Operation)
	r.taskMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestGetTask() {
	r.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.RunningStatus.String(),
			ExtraAttrs: map[string]interface{}{
				"resource_type":        "artifact",
				"source_resource":      "library/hello-world",
				"destination_resource": "library/hello-world",
				"operation":            "copy",
			},
		},
	}, nil)
	task, err := r.ctl.GetTask(nil, 1)
	r.Require().Nil(err)
	r.Equal(int64(1), task.ID)
	r.Equal(int64(1), task.ExecutionID)
	r.Equal("artifact", task.ResourceType)
	r.Equal("library/hello-world", task.SourceResource)
	r.Equal("library/hello-world", task.DestinationResource)
	r.Equal("copy", task.Operation)
	r.taskMgr.AssertExpectations(r.T())
}

func (r *replicationTestSuite) TestGetTaskLog() {
	r.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{
		{
			ID: 1,
		},
	}, nil)
	r.taskMgr.On("GetLog", mock.Anything, mock.Anything).Return([]byte{'a'}, nil)
	data, err := r.ctl.GetTaskLog(nil, 1)
	r.Require().Nil(err)
	r.Equal([]byte{'a'}, data)
	r.taskMgr.AssertExpectations(r.T())
}

func TestReplicationTestSuite(t *testing.T) {
	suite.Run(t, &replicationTestSuite{})
}
