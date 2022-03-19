package systemartifact

import (
	"context"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/task"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/systemartifact"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/pkg/errors"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SystemArtifactCleanupTestSuite struct {
	suite.Suite
	execMgr    *testingTask.ExecutionManager
	taskMgr    *testingTask.Manager
	cleanupMgr *systemartifact.Manager
	ctl        *controller
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
		suite.execMgr.On("StopAndWait", ctx, executionID, mock.Anything).Return(nil).Once()

		err := suite.ctl.Start(ctx, false, "SCHEDULE")
		suite.Error(err)
	}
}

func (suite *SystemArtifactCleanupTestSuite) TearDownSuite() {
	suite.execMgr = nil
	suite.taskMgr = nil
}

func TestScanDataExportExecutionTestSuite(t *testing.T) {
	suite.Run(t, &SystemArtifactCleanupTestSuite{})
}
