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
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan"
	dscan "github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/task"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	robottesting "github.com/goharbor/harbor/src/testing/controller/robot"
	"github.com/goharbor/harbor/src/testing/mock"
	postprocessorstesting "github.com/goharbor/harbor/src/testing/pkg/scan/postprocessors"
	reporttesting "github.com/goharbor/harbor/src/testing/pkg/scan/report"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
	"testing"
)

type CallbackTestSuite struct {
	suite.Suite

	artifactCtl         *artifacttesting.Controller
	originalArtifactCtl artifact.Controller

	execMgr *tasktesting.ExecutionManager

	robotCtl         *robottesting.Controller
	originalRobotCtl robot.Controller

	reportMgr *reporttesting.Manager

	scanCtl         Controller
	originalScanCtl Controller

	taskMgr         *tasktesting.Manager
	originalTaskMgr task.Manager
	reportConverter *postprocessorstesting.ScanReportV1ToV2Converter
}

func (suite *CallbackTestSuite) SetupSuite() {
	suite.originalArtifactCtl = artifact.Ctl
	suite.artifactCtl = &artifacttesting.Controller{}
	artifact.Ctl = suite.artifactCtl

	suite.execMgr = &tasktesting.ExecutionManager{}

	suite.originalRobotCtl = robot.Ctl
	suite.robotCtl = &robottesting.Controller{}
	robot.Ctl = suite.robotCtl

	suite.reportMgr = &reporttesting.Manager{}

	suite.originalTaskMgr = task.Mgr
	suite.taskMgr = &tasktesting.Manager{}
	task.Mgr = suite.taskMgr

	suite.originalScanCtl = DefaultController
	suite.reportConverter = &postprocessorstesting.ScanReportV1ToV2Converter{}

	suite.scanCtl = &basicController{
		makeCtx:         context.TODO,
		manager:         suite.reportMgr,
		execMgr:         suite.execMgr,
		taskMgr:         suite.taskMgr,
		reportConverter: suite.reportConverter,
	}
	DefaultController = suite.scanCtl
}

func (suite *CallbackTestSuite) TearDownSuite() {
	DefaultController = suite.originalScanCtl

	artifact.Ctl = suite.originalArtifactCtl
	robot.Ctl = suite.originalRobotCtl
	task.Mgr = suite.originalTaskMgr
}

func (suite *CallbackTestSuite) TestScanTaskStatusChange() {
	{
		// get task failed
		suite.taskMgr.On("Get", context.TODO(), int64(1)).Return(nil, fmt.Errorf("not found")).Once()
		suite.Error(scanTaskStatusChange(context.TODO(), 1, job.SuccessStatus.String()))
	}

	{
		// status success
		suite.taskMgr.On("Get", context.TODO(), int64(1)).Return(
			&task.Task{
				ExtraAttrs: suite.makeExtraAttrs(0, 1),
			},
			nil,
		).Once()
		suite.robotCtl.On("Delete", context.TODO(), int64(1)).Return(nil).Once()
		suite.NoError(scanTaskStatusChange(context.TODO(), 1, job.SuccessStatus.String()))
	}

	{
		// status success, delete robot failed
		suite.taskMgr.On("Get", context.TODO(), int64(1)).Return(
			&task.Task{
				ExtraAttrs: suite.makeExtraAttrs(0, 1),
			},
			nil,
		).Once()
		suite.robotCtl.On("Delete", context.TODO(), int64(1)).Return(fmt.Errorf("failed")).Once()
		suite.NoError(scanTaskStatusChange(context.TODO(), 1, job.SuccessStatus.String()))
	}

	{
		// status success, artifact not found
		suite.taskMgr.On("Get", context.TODO(), int64(1)).Return(
			&task.Task{
				ExtraAttrs: suite.makeExtraAttrs(1, 0),
			},
			nil,
		).Once()
		suite.artifactCtl.On("Get", context.TODO(), int64(1), (*artifact.Option)(nil)).Return(nil, fmt.Errorf("not found")).Once()
		suite.NoError(scanTaskStatusChange(context.TODO(), 1, job.SuccessStatus.String()))
	}

	{
		// status success
		suite.taskMgr.On("Get", context.TODO(), int64(1)).Return(
			&task.Task{
				ExtraAttrs: suite.makeExtraAttrs(1, 0),
			},
			nil,
		).Once()
		suite.artifactCtl.On("Get", context.TODO(), int64(1), (*artifact.Option)(nil)).Return(&artifact.Artifact{}, nil).Once()
		suite.NoError(scanTaskStatusChange(context.TODO(), 1, job.SuccessStatus.String()))
	}
}

func (suite *CallbackTestSuite) TestScanTaskCheckInProcessor() {
	{
		suite.Error(scanTaskCheckInProcessor(context.TODO(), &task.Task{}, "report"))
	}

	{
		suite.reportMgr.On("GetBy", context.TODO(), "digest", "ruuid", []string{"mime_type"}).Return(
			[]*dscan.Report{
				{UUID: "uuid"},
			},
			nil,
		).Once()

		suite.reportMgr.On("UpdateReportData", context.TODO(), "uuid", "raw_report").Return(nil)

		report := scan.CheckInReport{
			Digest:           "digest",
			RegistrationUUID: "ruuid",
			MimeType:         "mime_type",
			RawReport:        "raw_report",
		}

		r, _ := json.Marshal(report)
		suite.NoError(scanTaskCheckInProcessor(context.TODO(), &task.Task{}, string(r)))
	}
}

func (suite *CallbackTestSuite) TestScanAllCallback() {
	{
		// create execution failed
		suite.execMgr.On(
			"Create", context.TODO(), "IMAGE_SCAN_ALL", int64(0), "SCHEDULE", map[string]interface{}{},
		).Return(int64(0), fmt.Errorf("failed")).Once()

		suite.Error(scanAllCallback(context.TODO(), ""))
	}

	{
		executionID := int64(1)

		suite.execMgr.On(
			"Create", context.TODO(), "IMAGE_SCAN_ALL", int64(0), "SCHEDULE", map[string]interface{}{},
		).Return(executionID, nil).Once()

		suite.execMgr.On(
			"Get", context.TODO(), executionID,
		).Return(&task.Execution{}, nil)

		mock.OnAnything(suite.artifactCtl, "List").Return([]*artifact.Artifact{}, nil).Once()

		suite.taskMgr.On("Count", context.TODO(), q.New(q.KeyWords{"execution_id": executionID})).Return(int64(0), nil).Once()

		suite.execMgr.On("MarkDone", context.TODO(), executionID, "no artifact found").Return(nil).Once()

		suite.NoError(scanAllCallback(context.TODO(), ""))
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
