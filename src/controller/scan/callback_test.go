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
		suite.robotCtl.On("Delete", mock.Anything, int64(1)).Return(nil).Once()
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
		suite.robotCtl.On("Delete", mock.Anything, int64(1)).Return(fmt.Errorf("failed")).Once()
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

func (suite *CallbackTestSuite) makeExtraAttrs(artifactID, robotID int64) map[string]interface{} {
	b, _ := json.Marshal(map[string]interface{}{artifactIDKey: artifactID, robotIDKey: robotID})

	extraAttrs := map[string]interface{}{}
	json.Unmarshal(b, &extraAttrs)

	return extraAttrs
}

func TestCallbackTestSuite(t *testing.T) {
	suite.Run(t, &CallbackTestSuite{})
}
