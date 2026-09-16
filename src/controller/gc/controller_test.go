package gc

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	schedulertesting "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
)

type gcCtrTestSuite struct {
	suite.Suite
	scheduler *schedulertesting.Scheduler
	execMgr   *tasktesting.ExecutionManager
	taskMgr   *tasktesting.Manager
	ctl       *controller
}

func (g *gcCtrTestSuite) SetupTest() {
	g.execMgr = &tasktesting.ExecutionManager{}
	g.taskMgr = &tasktesting.Manager{}
	g.scheduler = &schedulertesting.Scheduler{}
	g.ctl = &controller{
		taskMgr:      g.taskMgr,
		exeMgr:       g.execMgr,
		schedulerMgr: g.scheduler,
	}
}

func (g *gcCtrTestSuite) TestStart() {
	g.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	g.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	g.taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)

	dataMap := make(map[string]any)
	p := Policy{
		DeleteUntagged: true,
		DeleteTag:      true,
		ExtraAttrs:     dataMap,
	}
	id, err := g.ctl.Start(nil, p, task.ExecutionTriggerManual)
	g.Nil(err)
	g.Equal(int64(1), id)
}

func (g *gcCtrTestSuite) TestStartInjectsMaxWorkers() {
	g.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	g.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	g.T().Setenv("GC_MAX_WORKERS", "8")

	p := Policy{
		DeleteUntagged: true,
		DeleteTag:      true,
		Workers:        24,
		ExtraAttrs:     make(map[string]any),
	}
	_, err := g.ctl.Start(nil, p, task.ExecutionTriggerManual)
	g.Nil(err)

	// the limit is resolved at execution time and shipped to the job, so that a
	// schedule created before the limit was configured is clamped as well
	jobArg, ok := g.taskMgr.Calls[len(g.taskMgr.Calls)-1].Arguments.Get(2).(*task.Job)
	g.True(ok)
	g.Equal(8, jobArg.Parameters["max_workers"])
	g.Equal(24, jobArg.Parameters["workers"])
}

func (g *gcCtrTestSuite) TestStartWithoutMaxWorkers() {
	g.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	g.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	p := Policy{
		DeleteUntagged: true,
		DeleteTag:      true,
		Workers:        24,
		ExtraAttrs:     make(map[string]any),
	}
	_, err := g.ctl.Start(nil, p, task.ExecutionTriggerManual)
	g.Nil(err)

	// no limit configured means the key is omitted entirely
	jobArg, ok := g.taskMgr.Calls[len(g.taskMgr.Calls)-1].Arguments.Get(2).(*task.Job)
	g.True(ok)
	_, exist := jobArg.Parameters["max_workers"]
	g.False(exist)
}

func (g *gcCtrTestSuite) TestStop() {
	g.execMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	g.Nil(g.ctl.Stop(nil, 1))
}

func (g *gcCtrTestSuite) TestGetTaskLog() {
	g.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.SuccessStatus.String(),
		},
	}, nil)
	g.taskMgr.On("GetLog", mock.Anything, mock.Anything).Return([]byte("hello world"), nil)

	log, err := g.ctl.GetTaskLog(nil, 1)
	g.Nil(err)
	g.Equal([]byte("hello world"), log)
}

func (g *gcCtrTestSuite) TestExecutionCount() {
	g.execMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	count, err := g.ctl.ExecutionCount(nil, q.New(q.KeyWords{"VendorType": "gc"}))
	g.Nil(err)
	g.Equal(int64(1), count)
}

func (g *gcCtrTestSuite) TestGetExecution() {
	g.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID:            1,
			Trigger:       "Manual",
			VendorType:    job.GarbageCollectionVendorType,
			StatusMessage: "Success",
		},
	}, nil)

	hs, err := g.ctl.GetExecution(nil, int64(1))
	g.Nil(err)

	g.Equal("Manual", hs.Trigger)
}

func (g *gcCtrTestSuite) TestListExecutions() {
	g.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID:      1,
			Trigger: "Manual",
		},
	}, nil)

	g.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{
		{
			ID:          112,
			ExecutionID: 1,
			Status:      job.SuccessStatus.String(),
		},
	}, nil)

	hs, err := g.ctl.ListExecutions(nil, q.New(q.KeyWords{"VendorType": "gc"}))

	g.Nil(err)
	g.Equal("Manual", hs[0].Trigger)
}

func (g *gcCtrTestSuite) TestListTasks() {
	g.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.RunningStatus.String(),
		},
	}, nil)
	tasks, err := g.ctl.ListTasks(nil, nil)
	g.Require().Nil(err)
	g.Require().Len(tasks, 1)
	g.Equal(int64(1), tasks[0].ID)
	g.Equal(int64(1), tasks[0].ExecutionID)
	g.taskMgr.AssertExpectations(g.T())
}

func (g *gcCtrTestSuite) TestGetSchedule() {
	g.scheduler.On("ListSchedules", mock.Anything, mock.Anything).Return([]*scheduler.Schedule{
		{
			ID:         1,
			VendorType: "gc",
		},
	}, nil)

	sche, err := g.ctl.GetSchedule(nil)
	g.Nil(err)
	g.Equal("gc", sche.VendorType)
}

func (g *gcCtrTestSuite) TestCreateSchedule() {
	g.scheduler.On("Schedule", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	dataMap := make(map[string]any)
	p := Policy{
		DeleteUntagged: true,
		DeleteTag:      true,
		ExtraAttrs:     dataMap,
		Workers:        3,
	}
	id, err := g.ctl.CreateSchedule(nil, "Daily", "* * * * * *", p)
	g.Nil(err)
	g.Equal(int64(1), id)
}

func (g *gcCtrTestSuite) TestDeleteSchedule() {
	g.scheduler.On("UnScheduleByVendor", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	g.Nil(g.ctl.DeleteSchedule(nil))
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &gcCtrTestSuite{})
}
