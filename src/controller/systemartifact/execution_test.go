package systemartifact

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/orm"
	scheduler2 "github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/scheduler"
	"github.com/goharbor/harbor/src/testing/pkg/systemartifact"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
)

type SystemArtifactCleanupTestSuite struct {
	suite.Suite
	execMgr    *testingTask.ExecutionManager
	taskMgr    *testingTask.Manager
	cleanupMgr *systemartifact.Manager
	ctl        *controller
	sched      *scheduler.Scheduler
}

func (suite *SystemArtifactCleanupTestSuite) SetupSuite() {
}

func (suite *SystemArtifactCleanupTestSuite) TestStartCleanup() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &systemartifact.Manager{}
	suite.ctl = &controller{
		execMgr:           suite.execMgr,
		taskMgr:           suite.taskMgr,
		systemArtifactMgr: suite.cleanupMgr,
		makeCtx:           func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}

	{

		ctx := context.TODO()

		executionID := int64(1)
		taskId := int64(1)

		suite.execMgr.On("Create", ctx, "SYSTEM_ARTIFACT_CLEANUP", int64(0), "SCHEDULE").Return(executionID, nil).Once()

		suite.taskMgr.On("Create", ctx, executionID, mock.Anything).Return(taskId, nil).Once()

		suite.execMgr.On("MarkDone", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.Start(ctx, false, "SCHEDULE")
		suite.NoError(err)
		jobMatcher := testifymock.MatchedBy(func(j *task.Job) bool {
			return "SYSTEM_ARTIFACT_CLEANUP" == j.Name
		})
		suite.taskMgr.AssertCalled(suite.T(), "Create", ctx, executionID, jobMatcher)
	}
}

func (suite *SystemArtifactCleanupTestSuite) TestStartCleanupErrorDuringCreate() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &systemartifact.Manager{}
	suite.ctl = &controller{
		execMgr:           suite.execMgr,
		taskMgr:           suite.taskMgr,
		systemArtifactMgr: suite.cleanupMgr,
		makeCtx:           func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}

	{

		ctx := context.TODO()

		executionID := int64(1)

		suite.execMgr.On(
			"Create", ctx, "SYSTEM_ARTIFACT_CLEANUP", int64(0), "SCHEDULE",
		).Return(int64(0), errors.New("test error")).Once()

		suite.execMgr.On("MarkDone", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.Start(ctx, false, "SCHEDULE")
		suite.Error(err)
	}
}

func (suite *SystemArtifactCleanupTestSuite) TestStartCleanupErrorDuringTaskCreate() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &systemartifact.Manager{}
	suite.ctl = &controller{
		execMgr:           suite.execMgr,
		taskMgr:           suite.taskMgr,
		systemArtifactMgr: suite.cleanupMgr,
		makeCtx:           func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}

	{

		ctx := context.TODO()

		executionID := int64(1)
		taskId := int64(0)

		suite.execMgr.On(
			"Create", ctx, "SYSTEM_ARTIFACT_CLEANUP", int64(0), "SCHEDULE",
		).Return(executionID, nil).Once()

		suite.taskMgr.On("Create", ctx, executionID, mock.Anything).Return(taskId, errors.New("test error")).Once()

		suite.execMgr.On("MarkError", ctx, executionID, mock.Anything).Return(nil).Once()
		suite.execMgr.On("StopAndWaitWithError", ctx, executionID, mock.Anything, mock.Anything).Return(nil).Once()

		err := suite.ctl.Start(ctx, false, "SCHEDULE")
		suite.Error(err)
	}
}

func (suite *SystemArtifactCleanupTestSuite) TestScheduleCleanupJobNoPreviousSchedule() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &systemartifact.Manager{}
	suite.sched = &scheduler.Scheduler{}

	suite.ctl = &controller{
		execMgr:           suite.execMgr,
		taskMgr:           suite.taskMgr,
		systemArtifactMgr: suite.cleanupMgr,
		makeCtx:           func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}

	var extraAttrs map[string]interface{}
	suite.sched.On("Schedule", mock.Anything,
		job.SystemArtifactCleanupVendorType, int64(0), cronTypeDaily, cronSpec, SystemArtifactCleanupCallback, nil, extraAttrs).Return(int64(1), nil)
	suite.sched.On("ListSchedules", mock.Anything, mock.Anything).Return(make([]*scheduler2.Schedule, 0), nil)
	sched = suite.sched
	ctx := context.TODO()

	ScheduleCleanupTask(ctx)

	suite.sched.AssertCalled(suite.T(), "Schedule", mock.Anything,
		job.SystemArtifactCleanupVendorType, int64(0), cronTypeDaily, cronSpec, SystemArtifactCleanupCallback, nil, extraAttrs)
}

func (suite *SystemArtifactCleanupTestSuite) TestScheduleCleanupJobPreviousSchedule() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &systemartifact.Manager{}
	suite.sched = &scheduler.Scheduler{}

	suite.ctl = &controller{
		execMgr:           suite.execMgr,
		taskMgr:           suite.taskMgr,
		systemArtifactMgr: suite.cleanupMgr,
		makeCtx:           func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}

	var extraAttrs map[string]interface{}
	suite.sched.On("Schedule", mock.Anything,
		job.SystemArtifactCleanupVendorType, int64(0), cronTypeDaily, cronSpec, SystemArtifactCleanupCallback, nil, extraAttrs).Return(int64(1), nil)

	existingSchedule := scheduler2.Schedule{ID: int64(10)}
	suite.sched.On("ListSchedules", mock.Anything, mock.Anything).Return([]*scheduler2.Schedule{&existingSchedule}, nil)
	sched = suite.sched
	ctx := context.TODO()

	ScheduleCleanupTask(ctx)

	suite.sched.AssertNotCalled(suite.T(), "Schedule", mock.Anything,
		job.SystemArtifactCleanupVendorType, int64(0), cronTypeDaily, cronSpec, SystemArtifactCleanupCallback, nil, extraAttrs)
}

func (suite *SystemArtifactCleanupTestSuite) TestScheduleCleanupJobPreviousScheduleError() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &systemartifact.Manager{}
	suite.sched = &scheduler.Scheduler{}

	suite.ctl = &controller{
		execMgr:           suite.execMgr,
		taskMgr:           suite.taskMgr,
		systemArtifactMgr: suite.cleanupMgr,
		makeCtx:           func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}

	suite.sched.On("Schedule", mock.Anything,
		job.SystemArtifactCleanupVendorType, int64(0), cronTypeDaily, cronSpec, SystemArtifactCleanupCallback, nil, mock.Anything).Return(int64(1), nil)

	suite.sched.On("ListSchedules", mock.Anything, mock.Anything).Return(nil, errors.New("test error"))
	sched = suite.sched
	ctx := context.TODO()

	ScheduleCleanupTask(ctx)

	extraAttributesMatcher := testifymock.MatchedBy(func(attrs map[string]interface{}) bool {
		return len(attrs) == 0
	})
	suite.sched.AssertNotCalled(suite.T(), "Schedule", mock.Anything,
		job.SystemArtifactCleanupVendorType, int64(0), cronTypeDaily, cronSpec, SystemArtifactCleanupCallback, nil, extraAttributesMatcher)
}

func (suite *SystemArtifactCleanupTestSuite) TearDownSuite() {
	suite.execMgr = nil
	suite.taskMgr = nil
	suite.cleanupMgr = nil
	suite.ctl = nil
	suite.sched = nil
}

func TestScanDataExportExecutionTestSuite(t *testing.T) {
	suite.Run(t, &SystemArtifactCleanupTestSuite{})
}
