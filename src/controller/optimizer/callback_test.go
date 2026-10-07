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

package optimizer

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/jobservice/job"
	dockerfileoptdao "github.com/goharbor/harbor/src/pkg/dockerfileoptimization/dao"
	"github.com/goharbor/harbor/src/pkg/task"
	robottesting "github.com/goharbor/harbor/src/testing/controller/robot"
	"github.com/goharbor/harbor/src/testing/mock"
	optdaotesting "github.com/goharbor/harbor/src/testing/pkg/dockerfileoptimization/dao"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
)

// CallbackTestSuite tests optimizeTaskStatusChange. The dependencies it exercises
// (robotCtl, taskMgr, execMgr, optDAO) are package-level vars in callback.go, so
// each test swaps them for mocks and restores the originals afterwards.
type CallbackTestSuite struct {
	suite.Suite

	origRobotCtl robot.Controller
	origTaskMgr  task.Manager
	origExecMgr  task.ExecutionManager
	origOptDAO   dockerfileoptdao.DAO

	mRC      *robottesting.Controller
	mTaskMgr *tasktesting.Manager
	mExecMgr *tasktesting.ExecutionManager
	mOptDAO  *optdaotesting.DAO
}

func TestCallback(t *testing.T) {
	suite.Run(t, new(CallbackTestSuite))
}

func (suite *CallbackTestSuite) SetupSuite() {
	suite.origRobotCtl = robotCtl
	suite.origTaskMgr = taskMgr
	suite.origExecMgr = execMgr
	suite.origOptDAO = optDAO
}

func (suite *CallbackTestSuite) TearDownSuite() {
	robotCtl = suite.origRobotCtl
	taskMgr = suite.origTaskMgr
	execMgr = suite.origExecMgr
	optDAO = suite.origOptDAO
}

func (suite *CallbackTestSuite) SetupTest() {
	suite.mRC = &robottesting.Controller{}
	suite.mTaskMgr = &tasktesting.Manager{}
	suite.mExecMgr = &tasktesting.ExecutionManager{}
	suite.mOptDAO = &optdaotesting.DAO{}

	robotCtl = suite.mRC
	taskMgr = suite.mTaskMgr
	execMgr = suite.mExecMgr
	optDAO = suite.mOptDAO
}

func (suite *CallbackTestSuite) TestNonFinalStatus_NoOp() {
	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.RunningStatus))
	suite.Require().NoError(err)
	// No expectations set on any mock: if a call were made, testify would panic
	// with "I don't know what to return" since no .On() was registered.
}

func (suite *CallbackTestSuite) TestTaskMgrGetError() {
	suite.mTaskMgr.On("Get", mock.Anything, int64(1)).Return(nil, fmt.Errorf("db error")).Once()

	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.SuccessStatus))
	suite.Require().Error(err)
}

func (suite *CallbackTestSuite) TestExecMgrGetError() {
	suite.mTaskMgr.On("Get", mock.Anything, int64(1)).Return(&task.Task{
		ID:          1,
		ExecutionID: 10,
	}, nil).Once()
	suite.mExecMgr.On("Get", mock.Anything, int64(10)).Return(nil, fmt.Errorf("db error")).Once()

	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.SuccessStatus))
	suite.Require().Error(err)
}

func (suite *CallbackTestSuite) TestSuccessStatus_DeletesRobotNoDAOUpdate() {
	extraAttrs := map[string]any{robotIDKey: float64(5)}
	suite.mTaskMgr.On("Get", mock.Anything, int64(1)).Return(&task.Task{
		ID:          1,
		ExecutionID: 10,
		ExtraAttrs:  extraAttrs,
	}, nil).Once()
	suite.mExecMgr.On("Get", mock.Anything, int64(10)).Return(&task.Execution{
		ID:         10,
		ExtraAttrs: map[string]any{"artifact": map[string]any{"repository_name": "library/photon", "digest": "sha256:d"}},
	}, nil).Once()
	suite.mRC.On("Delete", mock.Anything, int64(5), &robot.Option{Operator: "harbor-jobservice"}).Return(nil).Once()

	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.SuccessStatus))
	suite.Require().NoError(err)
	// Success is final but not Error/Stopped, so the safety net must not touch optDAO.
	suite.mOptDAO.AssertNotCalled(suite.T(), "GetByArtifact", mock.Anything, mock.Anything, mock.Anything)
	suite.mOptDAO.AssertNotCalled(suite.T(), "UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *CallbackTestSuite) TestSuccessStatus_NoRobotID_SkipsDelete() {
	suite.mTaskMgr.On("Get", mock.Anything, int64(1)).Return(&task.Task{
		ID:          1,
		ExecutionID: 10,
		ExtraAttrs:  map[string]any{},
	}, nil).Once()
	suite.mExecMgr.On("Get", mock.Anything, int64(10)).Return(&task.Execution{ID: 10}, nil).Once()

	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.SuccessStatus))
	suite.Require().NoError(err)
	suite.mRC.AssertNotCalled(suite.T(), "Delete", mock.Anything, mock.Anything, mock.Anything)
}

func (suite *CallbackTestSuite) TestSuccessStatus_RobotDeleteFailsIsLoggedNotReturned() {
	extraAttrs := map[string]any{robotIDKey: float64(5)}
	suite.mTaskMgr.On("Get", mock.Anything, int64(1)).Return(&task.Task{
		ID:          1,
		ExecutionID: 10,
		ExtraAttrs:  extraAttrs,
	}, nil).Once()
	suite.mExecMgr.On("Get", mock.Anything, int64(10)).Return(&task.Execution{ID: 10}, nil).Once()
	suite.mRC.On("Delete", mock.Anything, int64(5), mock.Anything).Return(fmt.Errorf("robot already gone")).Once()

	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.SuccessStatus))
	suite.Require().NoError(err)
}

func (suite *CallbackTestSuite) TestErrorStatus_PendingRecordMarkedErrored() {
	extraAttrs := map[string]any{robotIDKey: float64(5)}
	suite.mTaskMgr.On("Get", mock.Anything, int64(1)).Return(&task.Task{
		ID:          1,
		ExecutionID: 10,
		ExtraAttrs:  extraAttrs,
	}, nil).Once()
	suite.mExecMgr.On("Get", mock.Anything, int64(10)).Return(&task.Execution{
		ID: 10,
		ExtraAttrs: map[string]any{
			"artifact": map[string]any{"repository_name": "library/photon", "digest": "sha256:d"},
		},
	}, nil).Once()
	suite.mRC.On("Delete", mock.Anything, int64(5), mock.Anything).Return(nil).Once()
	suite.mOptDAO.On("GetByArtifact", mock.Anything, "library/photon", "sha256:d").Return(&dockerfileoptdao.DockerfileOptimization{
		Status: dockerfileoptdao.StatusPending,
	}, nil).Once()
	suite.mOptDAO.On("UpdateStatus", mock.Anything, "library/photon", "sha256:d", dockerfileoptdao.StatusError, mock.Anything).Return(nil).Once()

	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.ErrorStatus))
	suite.Require().NoError(err)
}

func (suite *CallbackTestSuite) TestErrorStatus_AlreadyTerminalRecordNotOverwritten() {
	suite.mTaskMgr.On("Get", mock.Anything, int64(1)).Return(&task.Task{
		ID:          1,
		ExecutionID: 10,
		ExtraAttrs:  map[string]any{},
	}, nil).Once()
	suite.mExecMgr.On("Get", mock.Anything, int64(10)).Return(&task.Execution{
		ID: 10,
		ExtraAttrs: map[string]any{
			"artifact": map[string]any{"repository_name": "library/photon", "digest": "sha256:d"},
		},
	}, nil).Once()
	suite.mOptDAO.On("GetByArtifact", mock.Anything, "library/photon", "sha256:d").Return(&dockerfileoptdao.DockerfileOptimization{
		Status: dockerfileoptdao.StatusSuccess,
	}, nil).Once()

	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.ErrorStatus))
	suite.Require().NoError(err)
	suite.mOptDAO.AssertNotCalled(suite.T(), "UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *CallbackTestSuite) TestStoppedStatus_RunningRecordMarkedErrored() {
	suite.mTaskMgr.On("Get", mock.Anything, int64(1)).Return(&task.Task{
		ID:          1,
		ExecutionID: 10,
		ExtraAttrs:  map[string]any{},
	}, nil).Once()
	suite.mExecMgr.On("Get", mock.Anything, int64(10)).Return(&task.Execution{
		ID: 10,
		ExtraAttrs: map[string]any{
			"artifact": map[string]any{"repository_name": "library/photon", "digest": "sha256:d"},
		},
	}, nil).Once()
	suite.mOptDAO.On("GetByArtifact", mock.Anything, "library/photon", "sha256:d").Return(&dockerfileoptdao.DockerfileOptimization{
		Status: dockerfileoptdao.StatusRunning,
	}, nil).Once()
	suite.mOptDAO.On("UpdateStatus", mock.Anything, "library/photon", "sha256:d", dockerfileoptdao.StatusError, mock.Anything).Return(nil).Once()

	err := optimizeTaskStatusChange(context.TODO(), 1, string(job.StoppedStatus))
	suite.Require().NoError(err)
}

func TestGetRobotID(t *testing.T) {
	suite.Run(t, new(getRobotIDSuite))
}

type getRobotIDSuite struct{ suite.Suite }

func (suite *getRobotIDSuite) TestNilExtraAttrs() {
	suite.Equal(int64(0), getRobotID(nil))
}

func (suite *getRobotIDSuite) TestMissingKey() {
	suite.Equal(int64(0), getRobotID(map[string]any{}))
}

func (suite *getRobotIDSuite) TestFloat64Value() {
	suite.Equal(int64(7), getRobotID(map[string]any{robotIDKey: float64(7)}))
}

func (suite *getRobotIDSuite) TestInt64Value() {
	suite.Equal(int64(7), getRobotID(map[string]any{robotIDKey: int64(7)}))
}

func TestGetArtifactCoords(t *testing.T) {
	suite.Run(t, new(getArtifactCoordsSuite))
}

type getArtifactCoordsSuite struct{ suite.Suite }

func (suite *getArtifactCoordsSuite) TestNilExtraAttrs() {
	repo, digest := getArtifactCoords(nil)
	suite.Empty(repo)
	suite.Empty(digest)
}

func (suite *getArtifactCoordsSuite) TestMissingArtifact() {
	repo, digest := getArtifactCoords(map[string]any{})
	suite.Empty(repo)
	suite.Empty(digest)
}

func (suite *getArtifactCoordsSuite) TestPresent() {
	repo, digest := getArtifactCoords(map[string]any{
		"artifact": map[string]any{"repository_name": "library/photon", "digest": "sha256:d"},
	})
	suite.Equal("library/photon", repo)
	suite.Equal("sha256:d", digest)
}
