package scandataexport

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	"github.com/goharbor/harbor/src/pkg/task"
	exporttesting "github.com/goharbor/harbor/src/testing/controller/scan/export"
	"github.com/goharbor/harbor/src/testing/mock"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/pkg/errors"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type ScanDataExportExecutionTestSuite struct {
	suite.Suite
	execMgr    *testingTask.ExecutionManager
	taskMgr    *testingTask.Manager
	cleanupMgr *exporttesting.CleanupManager
	ctl        *controller
}

func (suite *ScanDataExportExecutionTestSuite) SetupSuite() {
}

func (suite *ScanDataExportExecutionTestSuite) TestGetTask() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &exporttesting.CleanupManager{}
	suite.ctl = &controller{
		execMgr:    suite.execMgr,
		taskMgr:    suite.taskMgr,
		cleanupMgr: suite.cleanupMgr,
	}
	// valid task execution record exists for an execution id
	{
		t := task.Task{
			ID:             1,
			VendorType:     "SCAN_DATA_EXPORT",
			ExecutionID:    100,
			Status:         "Success",
			StatusMessage:  "",
			RunCount:       1,
			JobID:          "TestJobId",
			ExtraAttrs:     nil,
			CreationTime:   time.Time{},
			StartTime:      time.Time{},
			UpdateTime:     time.Time{},
			EndTime:        time.Time{},
			StatusRevision: 0,
		}

		tasks := make([]*task.Task, 0)
		tasks = append(tasks, &t)
		mock.OnAnything(suite.taskMgr, "List").Return(tasks, nil).Once()
		returnedTask, err := suite.ctl.GetTask(context.Background(), 100)
		suite.NoError(err)
		suite.Equal(t, *returnedTask)
	}

	// no task records exist for an execution id
	{
		tasks := make([]*task.Task, 0)
		mock.OnAnything(suite.taskMgr, "List").Return(tasks, nil).Once()
		_, err := suite.ctl.GetTask(context.Background(), 100)
		suite.Error(err)
	}

}

func (suite *ScanDataExportExecutionTestSuite) TestStart() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.ctl = &controller{
		execMgr: suite.execMgr,
		taskMgr: suite.taskMgr,
	}
	// execution manager and task manager return successfully
	{
		suite.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(10), nil)
		suite.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(20), nil)
		ctx := context.Background()
		ctx = context.WithValue(ctx, "vendorId", int(-1))
		criteria := export.Criteria{}
		executionId, err := suite.ctl.Start(ctx, criteria)
		suite.NoError(err)
		suite.Equal(int64(10), executionId)
		suite.validateExecutionManagerInvocation(ctx)
	}

}

func (suite *ScanDataExportExecutionTestSuite) TestStartWithExecManagerError() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.ctl = &controller{
		execMgr: suite.execMgr,
		taskMgr: suite.taskMgr,
	}
	// execution manager returns an error
	{
		ctx := context.Background()
		ctx = context.WithValue(ctx, "vendorId", int(-1))
		mock.OnAnything(suite.execMgr, "Create").Return(int64(-1), errors.New("Test Error"))
		_, err := suite.ctl.Start(ctx, export.Criteria{})
		suite.Error(err)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStartWithTaskManagerError() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.ctl = &controller{
		execMgr: suite.execMgr,
		taskMgr: suite.taskMgr,
	}
	// execution manager is successful but task manager returns an error
	// execution manager and task manager return successfully
	{
		ctx := context.Background()
		ctx = context.WithValue(ctx, "vendorId", int(-1))
		suite.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(10), nil)
		suite.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(-1), errors.New("Test Error"))
		mock.OnAnything(suite.execMgr, "StopAndWait").Return(nil)
		mock.OnAnything(suite.execMgr, "MarkError").Return(nil)
		_, err := suite.ctl.Start(ctx, export.Criteria{})
		suite.Error(err)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStartCleanup() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &exporttesting.CleanupManager{}
	suite.ctl = &controller{
		execMgr:    suite.execMgr,
		taskMgr:    suite.taskMgr,
		cleanupMgr: suite.cleanupMgr,
	}

	{

		ctx := context.TODO()

		executionID := int64(1)

		suite.execMgr.On(
			"Create", ctx, "EXPORT_DATA_CLEANUP", int64(0), "SCHEDULE",
		).Return(executionID, nil).Once()

		suite.cleanupMgr.On("Execute", mock.Anything).Return(nil).Once()

		suite.execMgr.On("MarkDone", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.StartCleanup(ctx, "SCHEDULE", TriggerParam{}, false)
		suite.NoError(err)
		suite.cleanupMgr.AssertNotCalled(suite.T(), "Configure")
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStartCleanupErrorDuringCreate() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &exporttesting.CleanupManager{}
	suite.ctl = &controller{
		execMgr:    suite.execMgr,
		taskMgr:    suite.taskMgr,
		cleanupMgr: suite.cleanupMgr,
	}

	{

		ctx := context.TODO()

		executionID := int64(1)

		suite.execMgr.On(
			"Create", ctx, "EXPORT_DATA_CLEANUP", int64(0), "SCHEDULE",
		).Return(int64(0), errors.New("test error")).Once()

		suite.cleanupMgr.On("Execute", mock.Anything).Return(nil).Once()

		suite.execMgr.On("MarkDone", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.StartCleanup(ctx, "SCHEDULE", TriggerParam{}, false)
		suite.Error(err)
		suite.cleanupMgr.AssertNotCalled(suite.T(), "Configure")
		suite.cleanupMgr.AssertNotCalled(suite.T(), "Execute", mock.Anything)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStartCleanupErrorDuringExecute() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &exporttesting.CleanupManager{}
	suite.ctl = &controller{
		execMgr:    suite.execMgr,
		taskMgr:    suite.taskMgr,
		cleanupMgr: suite.cleanupMgr,
	}

	{
		ctx := context.TODO()

		executionID := int64(1)

		suite.execMgr.On(
			"Create", ctx, "EXPORT_DATA_CLEANUP", int64(0), "SCHEDULE",
		).Return(int64(1), nil).Once()

		suite.cleanupMgr.On("Execute", mock.Anything).Return(errors.New("test error")).Once()

		suite.execMgr.On("MarkDone", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.StartCleanup(ctx, "SCHEDULE", TriggerParam{}, false)
		suite.Error(err)
		suite.cleanupMgr.AssertNotCalled(suite.T(), "Configure")
		suite.cleanupMgr.AssertCalled(suite.T(), "Execute", mock.Anything)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStartCleanupWithTimeWindowOverride() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &exporttesting.CleanupManager{}
	suite.ctl = &controller{
		execMgr:    suite.execMgr,
		taskMgr:    suite.taskMgr,
		cleanupMgr: suite.cleanupMgr,
	}

	{

		ctx := context.TODO()

		executionID := int64(1)

		suite.execMgr.On(
			"Create", ctx, "EXPORT_DATA_CLEANUP", int64(0), "SCHEDULE",
		).Return(executionID, nil).Once()

		suite.cleanupMgr.On("Configure", mock.Anything).Once()
		suite.cleanupMgr.On("Execute", mock.Anything).Return(nil).Once()

		suite.execMgr.On("MarkDone", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.StartCleanup(ctx, "SCHEDULE", TriggerParam{TimeWindowMinutes: 10}, false)
		suite.NoError(err)

		cleanupSettingsMatcher := testifymock.MatchedBy(func(cs *export.CleanupSettings) bool {
			return cs.Get(export.TimeWindowKey).(int) == 10 && cs.Get(export.PageSizeKey) == nil
		})
		suite.cleanupMgr.AssertCalled(suite.T(), "Configure", cleanupSettingsMatcher)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStartCleanupWithPageSizeOverride() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &exporttesting.CleanupManager{}
	suite.ctl = &controller{
		execMgr:    suite.execMgr,
		taskMgr:    suite.taskMgr,
		cleanupMgr: suite.cleanupMgr,
	}

	{

		ctx := context.TODO()

		executionID := int64(1)

		suite.execMgr.On(
			"Create", ctx, "EXPORT_DATA_CLEANUP", int64(0), "SCHEDULE",
		).Return(executionID, nil).Once()

		suite.cleanupMgr.On("Configure", mock.Anything).Once()
		suite.cleanupMgr.On("Execute", mock.Anything).Return(nil).Once()

		suite.execMgr.On("MarkDone", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.StartCleanup(ctx, "SCHEDULE", TriggerParam{PageSize: 10}, false)
		suite.NoError(err)

		cleanupSettingsMatcher := testifymock.MatchedBy(func(cs *export.CleanupSettings) bool {
			return cs.Get(export.PageSizeKey).(int) == 10 && cs.Get(export.TimeWindowKey) == nil
		})
		suite.cleanupMgr.AssertCalled(suite.T(), "Configure", cleanupSettingsMatcher)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStartCleanupWithCompleteOverride() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.cleanupMgr = &exporttesting.CleanupManager{}
	suite.ctl = &controller{
		execMgr:    suite.execMgr,
		taskMgr:    suite.taskMgr,
		cleanupMgr: suite.cleanupMgr,
	}

	{

		ctx := context.TODO()

		executionID := int64(1)

		suite.execMgr.On(
			"Create", ctx, "EXPORT_DATA_CLEANUP", int64(0), "SCHEDULE",
		).Return(executionID, nil).Once()

		suite.cleanupMgr.On("Configure", mock.Anything).Once()
		suite.cleanupMgr.On("Execute", mock.Anything).Return(nil).Once()

		suite.execMgr.On("MarkDone", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.StartCleanup(ctx, "SCHEDULE", TriggerParam{PageSize: 10, TimeWindowMinutes: 10}, false)
		suite.NoError(err)

		cleanupSettingsMatcher := testifymock.MatchedBy(func(cs *export.CleanupSettings) bool {
			return cs.Get(export.PageSizeKey).(int) == 10 && cs.Get(export.TimeWindowKey) == 10
		})
		suite.cleanupMgr.AssertCalled(suite.T(), "Configure", cleanupSettingsMatcher)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TearDownSuite() {
	suite.execMgr = nil
	suite.taskMgr = nil
}

func (suite *ScanDataExportExecutionTestSuite) validateExecutionManagerInvocation(ctx context.Context) {
	// validate that execution manager has been called with the specified
	suite.execMgr.AssertCalled(suite.T(), "Create", ctx, job.ScanDataExport, int64(-1), task.ExecutionTriggerManual)
}

func TestScanDataExportExecutionTestSuite(t *testing.T) {
	suite.Run(t, &ScanDataExportExecutionTestSuite{})
}
