package scandataexport

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	"github.com/goharbor/harbor/src/pkg/task"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	systemartifacttesting "github.com/goharbor/harbor/src/testing/pkg/systemartifact"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
)

type ScanDataExportExecutionTestSuite struct {
	suite.Suite
	execMgr        *testingTask.ExecutionManager
	taskMgr        *testingTask.Manager
	sysArtifactMgr *systemartifacttesting.Manager
	ctl            *controller
}

func (suite *ScanDataExportExecutionTestSuite) SetupSuite() {
}

func (suite *ScanDataExportExecutionTestSuite) TestGetTask() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.sysArtifactMgr = &systemartifacttesting.Manager{}
	suite.ctl = &controller{
		execMgr:        suite.execMgr,
		taskMgr:        suite.taskMgr,
		makeCtx:        func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
		sysArtifactMgr: suite.sysArtifactMgr,
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

	// listing of tasks returns an error
	{
		mock.OnAnything(suite.taskMgr, "List").Return(nil, errors.New("test error")).Once()
		_, err := suite.ctl.GetTask(context.Background(), 100)
		suite.Error(err)
	}

}

func (suite *ScanDataExportExecutionTestSuite) TestGetExecution() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.sysArtifactMgr = &systemartifacttesting.Manager{}
	suite.ctl = &controller{
		execMgr:        suite.execMgr,
		taskMgr:        suite.taskMgr,
		makeCtx:        func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
		sysArtifactMgr: suite.sysArtifactMgr,
	}
	// get execution succeeds
	attrs := make(map[string]interface{})
	attrs[export.JobNameAttribute] = "test-job"
	attrs[export.UserNameAttribute] = "test-user"
	attrs[export.DigestKey] = "sha256:d04b98f48e8f8bcc15c6ae5ac050801cd6dcfd428fb5f9e65c4e16e7807340fa"
	attrs["status_message"] = "test-message"
	{
		exec := task.Execution{
			ID:            100,
			VendorType:    "SCAN_DATA_EXPORT",
			VendorID:      -1,
			Status:        "Success",
			StatusMessage: "",
			Metrics:       nil,
			Trigger:       "Manual",
			ExtraAttrs:    attrs,
			StartTime:     time.Time{},
			UpdateTime:    time.Time{},
			EndTime:       time.Time{},
		}
		mock.OnAnything(suite.execMgr, "Get").Return(&exec, nil).Once()
		mock.OnAnything(suite.sysArtifactMgr, "Exists").Return(true, nil).Once()

		exportExec, err := suite.ctl.GetExecution(context.TODO(), 100)
		suite.NoError(err)
		suite.Equal(exec.ID, exportExec.ID)
		suite.Equal("test-user", exportExec.UserName)
		suite.Equal("test-job", exportExec.JobName)
		suite.Equal("test-message", exportExec.StatusMessage)
		suite.Equal(true, exportExec.FilePresent)
	}

	// get execution fails
	{
		mock.OnAnything(suite.execMgr, "Get").Return(nil, errors.New("test error")).Once()
		exportExec, err := suite.ctl.GetExecution(context.TODO(), 100)
		suite.Error(err)
		suite.Nil(exportExec)
	}

	// get execution returns null
	{
		mock.OnAnything(suite.execMgr, "Get").Return(nil, nil).Once()
		exportExec, err := suite.ctl.GetExecution(context.TODO(), 100)
		suite.NoError(err)
		suite.Nil(exportExec)
	}

}

func (suite *ScanDataExportExecutionTestSuite) TestGetExecutionSysArtifactExistFail() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.sysArtifactMgr = &systemartifacttesting.Manager{}
	suite.ctl = &controller{
		execMgr:        suite.execMgr,
		taskMgr:        suite.taskMgr,
		makeCtx:        func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
		sysArtifactMgr: suite.sysArtifactMgr,
	}
	// get execution succeeds
	attrs := make(map[string]interface{})
	attrs[export.JobNameAttribute] = "test-job"
	attrs[export.UserNameAttribute] = "test-user"
	{
		exec := task.Execution{
			ID:            100,
			VendorType:    "SCAN_DATA_EXPORT",
			VendorID:      -1,
			Status:        "Success",
			StatusMessage: "",
			Metrics:       nil,
			Trigger:       "Manual",
			ExtraAttrs:    attrs,
			StartTime:     time.Time{},
			UpdateTime:    time.Time{},
			EndTime:       time.Time{},
		}
		mock.OnAnything(suite.execMgr, "Get").Return(&exec, nil).Once()
		mock.OnAnything(suite.sysArtifactMgr, "Exists").Return(false, errors.New("test error")).Once()

		exportExec, err := suite.ctl.GetExecution(context.TODO(), 100)
		suite.NoError(err)
		suite.Equal(exec.ID, exportExec.ID)
		suite.Equal("test-user", exportExec.UserName)
		suite.Equal("test-job", exportExec.JobName)
		suite.Equal(false, exportExec.FilePresent)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestGetExecutionList() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.sysArtifactMgr = &systemartifacttesting.Manager{}
	suite.ctl = &controller{
		execMgr:        suite.execMgr,
		taskMgr:        suite.taskMgr,
		makeCtx:        func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
		sysArtifactMgr: suite.sysArtifactMgr,
	}
	// get execution succeeds
	attrs := make(map[string]interface{})
	attrs[export.JobNameAttribute] = "test-job"
	attrs[export.UserNameAttribute] = "test-user"
	{
		exec := task.Execution{
			ID:            100,
			VendorType:    "SCAN_DATA_EXPORT",
			VendorID:      -1,
			Status:        "Success",
			StatusMessage: "",
			Metrics:       nil,
			Trigger:       "Manual",
			ExtraAttrs:    attrs,
			StartTime:     time.Time{},
			UpdateTime:    time.Time{},
			EndTime:       time.Time{},
		}
		execs := make([]*task.Execution, 0)
		execs = append(execs, &exec)
		mock.OnAnything(suite.execMgr, "List").Return(execs, nil).Once()
		mock.OnAnything(suite.sysArtifactMgr, "Exists").Return(true, nil).Once()
		exportExec, err := suite.ctl.ListExecutions(context.TODO(), "test-user")
		suite.NoError(err)

		suite.Equal(1, len(exportExec))
		suite.Equal("test-user", exportExec[0].UserName)
		suite.Equal("test-job", exportExec[0].JobName)
	}

	// get execution fails
	{
		mock.OnAnything(suite.execMgr, "List").Return(nil, errors.New("test error")).Once()
		exportExec, err := suite.ctl.ListExecutions(context.TODO(), "test-user")
		suite.Error(err)
		suite.Nil(exportExec)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStart() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.ctl = &controller{
		execMgr: suite.execMgr,
		taskMgr: suite.taskMgr,
		makeCtx: func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}
	// execution manager and task manager return successfully
	{
		// get execution succeeds
		attrs := make(map[string]interface{})
		attrs[export.ProjectIDsAttribute] = []int64{1}
		attrs[export.JobNameAttribute] = "test-job"
		attrs[export.UserNameAttribute] = "test-user"
		suite.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, attrs).Return(int64(10), nil)
		suite.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(20), nil)
		ctx := context.Background()
		ctx = context.WithValue(ctx, export.CsvJobVendorIDKey, int(-1))
		criteria := export.Request{}
		criteria.Projects = []int64{1}
		criteria.UserName = "test-user"
		criteria.JobName = "test-job"
		executionId, err := suite.ctl.Start(ctx, criteria)
		suite.NoError(err)
		suite.Equal(int64(10), executionId)
		suite.validateExecutionManagerInvocation(ctx)
	}

}

func (suite *ScanDataExportExecutionTestSuite) TestDeleteExecution() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.ctl = &controller{
		execMgr: suite.execMgr,
		taskMgr: suite.taskMgr,
		makeCtx: func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}
	mock.OnAnything(suite.execMgr, "Delete").Return(nil).Once()
	err := suite.ctl.DeleteExecution(context.TODO(), int64(1))
	suite.NoError(err)
}

func (suite *ScanDataExportExecutionTestSuite) TestStartWithExecManagerError() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.ctl = &controller{
		execMgr: suite.execMgr,
		taskMgr: suite.taskMgr,
		makeCtx: func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}
	// execution manager returns an error
	{
		ctx := context.Background()
		ctx = context.WithValue(ctx, export.CsvJobVendorIDKey, int(-1))
		mock.OnAnything(suite.execMgr, "Create").Return(int64(-1), errors.New("Test Error"))
		_, err := suite.ctl.Start(ctx, export.Request{JobName: "test-job", UserName: "test-user"})
		suite.Error(err)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TestStartWithTaskManagerError() {
	suite.taskMgr = &testingTask.Manager{}
	suite.execMgr = &testingTask.ExecutionManager{}
	suite.ctl = &controller{
		execMgr: suite.execMgr,
		taskMgr: suite.taskMgr,
		makeCtx: func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
	}
	// execution manager is successful but task manager returns an error
	// execution manager and task manager return successfully
	{
		ctx := context.Background()
		ctx = context.WithValue(ctx, export.CsvJobVendorIDKey, int(-1))
		attrs := make(map[string]interface{})
		attrs[export.ProjectIDsAttribute] = []int64{1}
		attrs[export.JobNameAttribute] = "test-job"
		attrs[export.UserNameAttribute] = "test-user"
		suite.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, attrs).Return(int64(10), nil)
		suite.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(-1), errors.New("Test Error"))
		mock.OnAnything(suite.execMgr, "StopAndWait").Return(nil)
		mock.OnAnything(suite.execMgr, "MarkError").Return(nil)
		_, err := suite.ctl.Start(ctx, export.Request{JobName: "test-job", UserName: "test-user", Projects: []int64{1}})
		suite.Error(err)
	}
}

func (suite *ScanDataExportExecutionTestSuite) TearDownSuite() {
	suite.execMgr = nil
	suite.taskMgr = nil
}

func (suite *ScanDataExportExecutionTestSuite) validateExecutionManagerInvocation(ctx context.Context) {
	// validate that execution manager has been called with the specified
	extraAttsMatcher := testifymock.MatchedBy(func(m map[string]interface{}) bool {
		jobName, jobNamePresent := m[export.JobNameAttribute]
		userName, userNamePresent := m[export.UserNameAttribute]
		return jobNamePresent && userNamePresent && jobName == "test-job" && userName == "test-user"
	})
	suite.execMgr.AssertCalled(suite.T(), "Create", ctx, job.ScanDataExportVendorType, int64(-1), task.ExecutionTriggerManual, extraAttsMatcher)
}

func TestScanDataExportExecutionTestSuite(t *testing.T) {
	suite.Run(t, &ScanDataExportExecutionTestSuite{})
}
