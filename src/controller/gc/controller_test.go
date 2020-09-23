package gc

import (
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	schedulertesting "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
	"testing"
)

type gcCtrTestSuite struct {
	suite.Suite
	scheduler *schedulertesting.Scheduler
	execMgr   *tasktesting.FakeExecutionManager
	taskMgr   *tasktesting.FakeManager
	ctl       *controller
}

func (g *gcCtrTestSuite) SetupTest() {
	g.execMgr = &tasktesting.FakeExecutionManager{}
	g.taskMgr = &tasktesting.FakeManager{}
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

	dataMap := make(map[string]interface{})
	g.Nil(g.ctl.Start(nil, dataMap))
}

func (g *gcCtrTestSuite) TestStop() {
	g.taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	g.Nil(g.ctl.Stop(nil, 1))
}

func (g *gcCtrTestSuite) TestLog() {
	g.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{
		{
			ID:          1,
			ExecutionID: 1,
			Status:      job.SuccessStatus.String(),
		},
	}, nil)
	g.taskMgr.On("GetLog", mock.Anything, mock.Anything).Return([]byte("hello world"), nil)

	log, err := g.ctl.GetLog(nil, 1)
	g.Nil(err)
	g.Equal([]byte("hello world"), log)
}

func (g *gcCtrTestSuite) TestCount() {
	g.execMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	count, err := g.ctl.Count(nil, q.New(q.KeyWords{"VendorType": "gc"}))
	g.Nil(err)
	g.Equal(int64(1), count)
}

func (g *gcCtrTestSuite) TestHistory() {
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

	hs, err := g.ctl.History(nil, q.New(q.KeyWords{"VendorType": "gc"}))

	g.Nil(err)
	g.Equal("Manual", hs[0].Kind)
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
	g.scheduler.On("Schedule", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	dataMap := make(map[string]interface{})
	id, err := g.ctl.CreateSchedule(nil, "", dataMap)
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
