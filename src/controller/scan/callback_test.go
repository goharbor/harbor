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

package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/task"
	taskdao "github.com/goharbor/harbor/src/pkg/task/dao"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	robottesting "github.com/goharbor/harbor/src/testing/controller/robot"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	postprocessorstesting "github.com/goharbor/harbor/src/testing/pkg/scan/postprocessors"
	reporttesting "github.com/goharbor/harbor/src/testing/pkg/scan/report"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
)

type CallbackTestSuite struct {
	suite.Suite

	ctx context.Context

	artifactCtl *artifacttesting.Controller

	execMgr *tasktesting.ExecutionManager

	robotCtl *robottesting.Controller

	reportMgr *reporttesting.Manager

	scanCtl Controller

	taskMgr         *tasktesting.Manager
	reportConverter *postprocessorstesting.NativeScanReportConverter
}

func (suite *CallbackTestSuite) SetupSuite() {
	suite.ctx = orm.NewContext(nil, &ormtesting.FakeOrmer{})
	suite.artifactCtl = &artifacttesting.Controller{}
	artifactCtl = suite.artifactCtl

	suite.execMgr = &tasktesting.ExecutionManager{}

	suite.robotCtl = &robottesting.Controller{}
	robotCtl = suite.robotCtl

	suite.reportMgr = &reporttesting.Manager{}

	suite.taskMgr = &tasktesting.Manager{}
	taskMgr = suite.taskMgr

	suite.reportConverter = &postprocessorstesting.NativeScanReportConverter{}

	suite.scanCtl = &basicController{
		makeCtx:         context.TODO,
		manager:         suite.reportMgr,
		execMgr:         suite.execMgr,
		taskMgr:         suite.taskMgr,
		reportConverter: suite.reportConverter,
	}
	scanCtl = suite.scanCtl
}

func (suite *CallbackTestSuite) TestScanTaskStatusChange() {
	{
		// get task failed
		suite.taskMgr.On("Get", mock.Anything, int64(1)).Return(nil, fmt.Errorf("not found")).Once()
		suite.Error(scanTaskStatusChange(suite.ctx, 1, job.SuccessStatus.String()))
	}

	{
		// status success
		suite.taskMgr.On("Get", mock.Anything, int64(1)).Return(
			&task.Task{
				ExtraAttrs: suite.makeExtraAttrs(0, 1),
			},
			nil,
		).Once()
		suite.robotCtl.On("Delete", mock.Anything, int64(1), mock.Anything).Return(nil).Once()
		suite.NoError(scanTaskStatusChange(suite.ctx, 1, job.SuccessStatus.String()))
	}

	{
		// status success, delete robot failed
		suite.taskMgr.On("Get", mock.Anything, int64(1)).Return(
			&task.Task{
				ExtraAttrs: suite.makeExtraAttrs(0, 1),
			},
			nil,
		).Once()
		suite.robotCtl.On("Delete", mock.Anything, int64(1), mock.Anything).Return(fmt.Errorf("failed")).Once()
		suite.NoError(scanTaskStatusChange(suite.ctx, 1, job.SuccessStatus.String()))
	}

	{
		// status success, artifact not found
		suite.taskMgr.On("Get", mock.Anything, int64(1)).Return(
			&task.Task{
				ExtraAttrs: suite.makeExtraAttrs(1, 0),
			},
			nil,
		).Once()
		suite.artifactCtl.On("Get", mock.Anything, int64(1), (*artifact.Option)(nil)).Return(nil, fmt.Errorf("not found")).Once()
		suite.NoError(scanTaskStatusChange(suite.ctx, 1, job.SuccessStatus.String()))
	}

	{
		// status success
		suite.taskMgr.On("Get", mock.Anything, int64(1)).Return(
			&task.Task{
				ExtraAttrs: suite.makeExtraAttrs(1, 0),
			},
			nil,
		).Once()
		suite.artifactCtl.On("Get", mock.Anything, int64(1), (*artifact.Option)(nil)).Return(&artifact.Artifact{}, nil).Once()
		suite.NoError(scanTaskStatusChange(suite.ctx, 1, job.SuccessStatus.String()))
	}
}

func (suite *CallbackTestSuite) TestScanAllCallback() {
	{
		// create execution failed
		suite.execMgr.On(
			"Create", mock.Anything, "SCAN_ALL", int64(0), "SCHEDULE",
			mock.Anything).Return(int64(0), fmt.Errorf("failed")).Once()

		suite.Error(scanAllCallback(suite.ctx, ""))
	}

	{
		executionID := int64(1)

		suite.execMgr.On(
			"Create", mock.Anything, "SCAN_ALL", int64(0), "SCHEDULE",
			mock.Anything).Return(executionID, nil).Once()

		suite.execMgr.On(
			"Get", mock.Anything, executionID,
		).Return(&task.Execution{}, nil)

		mock.OnAnything(suite.artifactCtl, "List").Return([]*artifact.Artifact{}, nil).Once()

		mock.OnAnything(suite.execMgr, "UpdateExtraAttrs").Return(nil).Once()

		suite.execMgr.On("MarkDone", mock.Anything, executionID, mock.Anything).Return(nil).Once()

		suite.NoError(scanAllCallback(suite.ctx, ""))
	}
}

func (suite *CallbackTestSuite) makeExtraAttrs(artifactID, robotID int64) map[string]any {
	b, _ := json.Marshal(map[string]any{artifactIDKey: artifactID, robotIDKey: robotID})

	daoTask := &taskdao.Task{
		ExtraAttrs: string(b),
	}
	tsk := &task.Task{}
	tsk.From(daoTask)

	return tsk.ExtraAttrs
}

func TestGetRobotID_LargeID(t *testing.T) {
	// 2^53 + 1, cannot be represented exactly in float64.
	// The robot ID sequence is capped at 2^53 - 1, so this value is intentionally
	// out of range for real robots; it verifies the extraction path is lossless for any int64
	largeID := int64(9007199254740993)

	b, _ := json.Marshal(map[string]any{robotIDKey: largeID, artifactIDKey: largeID})
	daoTask := &taskdao.Task{
		ExtraAttrs: string(b),
	}
	tsk := &task.Task{}
	tsk.From(daoTask)

	gotRobotID := getRobotID(tsk.ExtraAttrs)
	if gotRobotID != largeID {
		t.Errorf("getRobotID() = %d, want %d", gotRobotID, largeID)
	}

	gotArtifactID := getArtifactID(tsk.ExtraAttrs)
	if gotArtifactID != largeID {
		t.Errorf("getArtifactID() = %d, want %d", gotArtifactID, largeID)
	}
}

func TestGetRobotID_LegacyFloat64(t *testing.T) {
	// old data decoded without UseNumber comes in as float64
	extraAttrs := map[string]any{robotIDKey: float64(12345), artifactIDKey: float64(67890)}

	if got := getRobotID(extraAttrs); got != 12345 {
		t.Errorf("getRobotID() = %d, want 12345", got)
	}
	if got := getArtifactID(extraAttrs); got != 67890 {
		t.Errorf("getArtifactID() = %d, want 67890", got)
	}
}

func TestCallbackTestSuite(t *testing.T) {
	suite.Run(t, &CallbackTestSuite{})
}
