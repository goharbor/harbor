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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	accessoryModel "github.com/goharbor/harbor/src/pkg/accessory/model"
	art "github.com/goharbor/harbor/src/pkg/artifact"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	sca "github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/pkg/task"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	robottesting "github.com/goharbor/harbor/src/testing/controller/robot"
	scannertesting "github.com/goharbor/harbor/src/testing/controller/scanner"
	tagtesting "github.com/goharbor/harbor/src/testing/controller/tag"
	mockcache "github.com/goharbor/harbor/src/testing/lib/cache"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
	postprocessorstesting "github.com/goharbor/harbor/src/testing/pkg/scan/postprocessors"
	reporttesting "github.com/goharbor/harbor/src/testing/pkg/scan/report"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
)

// ControllerTestSuite is the test suite for scan controller.
type ControllerTestSuite struct {
	suite.Suite

	artifactCtl         *artifacttesting.Controller
	accessoryMgr        *accessorytesting.Manager
	originalArtifactCtl artifact.Controller

	tagCtl *tagtesting.FakeController

	registration *scanner.Registration
	artifact     *artifact.Artifact
	rawReport    string

	execMgr         *tasktesting.ExecutionManager
	taskMgr         *tasktesting.Manager
	reportMgr       *reporttesting.Manager
	ar              artifact.Controller
	c               Controller
	reportConverter *postprocessorstesting.ScanReportV1ToV2Converter
	cache           *mockcache.Cache
}

// TestController is the entry point of ControllerTestSuite.
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// SetupSuite ...
func (suite *ControllerTestSuite) SetupSuite() {
	suite.originalArtifactCtl = artifact.Ctl
	suite.artifactCtl = &artifacttesting.Controller{}
	artifact.Ctl = suite.artifactCtl

	suite.artifact = &artifact.Artifact{Artifact: art.Artifact{ID: 1}}
	suite.artifact.Type = "IMAGE"
	suite.artifact.ProjectID = 1
	suite.artifact.RepositoryName = "library/photon"
	suite.artifact.Digest = "digest-code"
	suite.artifact.ManifestMediaType = v1.MimeTypeDockerArtifact

	m := &v1.ScannerAdapterMetadata{
		Scanner: &v1.Scanner{
			Name:    "Trivy",
			Vendor:  "Harbor",
			Version: "0.1.0",
		},
		Capabilities: []*v1.ScannerCapability{{
			ConsumesMimeTypes: []string{
				v1.MimeTypeOCIArtifact,
				v1.MimeTypeDockerArtifact,
			},
			ProducesMimeTypes: []string{
				v1.MimeTypeNativeReport,
			},
		}},
		Properties: v1.ScannerProperties{
			"extra": "testing",
		},
	}

	suite.registration = &scanner.Registration{
		ID:        1,
		UUID:      "uuid001",
		Name:      "Test-scan-controller",
		URL:       "http://testing.com:3128",
		IsDefault: true,
		Metadata:  m,
	}

	sc := &scannertesting.Controller{}
	sc.On("GetRegistrationByProject", mock.Anything, suite.artifact.ProjectID).Return(suite.registration, nil)
	sc.On("Ping", suite.registration).Return(m, nil)

	mgr := &reporttesting.Manager{}
	mgr.On("Create", mock.Anything, &scan.Report{
		Digest:           "digest-code",
		RegistrationUUID: "uuid001",
		MimeType:         "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0",
	}).Return("r-uuid", nil)

	rp := vuln.Report{
		GeneratedAt: time.Now().UTC().String(),
		Scanner: &v1.Scanner{
			Name:    "Trivy",
			Vendor:  "Harbor",
			Version: "0.1.0",
		},
		Severity: vuln.High,
		Vulnerabilities: []*vuln.VulnerabilityItem{
			{
				ID:          "2019-0980-0909",
				Package:     "dpkg",
				Version:     "0.9.1",
				FixVersion:  "0.9.2",
				Severity:    vuln.High,
				Description: "mock one",
				Links:       []string{"https://vuln.com"},
			},
		},
	}

	jsonData, err := json.Marshal(rp)
	require.NoError(suite.T(), err)
	suite.rawReport = string(jsonData)

	reports := []*scan.Report{
		{
			ID:               11,
			UUID:             "rp-uuid-001",
			Digest:           "digest-code",
			RegistrationUUID: "uuid001",
			MimeType:         "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0",
			Status:           "Success",
			Report:           suite.rawReport,
			StartTime:        time.Now(),
			EndTime:          time.Now().Add(2 * time.Second),
		},
	}

	mgr.On("GetBy", mock.Anything, suite.artifact.Digest, suite.registration.UUID, []string{v1.MimeTypeNativeReport}).Return(reports, nil)
	mgr.On("Get", mock.Anything, "rp-uuid-001").Return(reports[0], nil)
	mgr.On("UpdateReportData", "rp-uuid-001", suite.rawReport, (int64)(10000)).Return(nil)
	mgr.On("UpdateStatus", "the-uuid-123", "Success", (int64)(10000)).Return(nil)
	suite.reportMgr = mgr

	rc := &robottesting.Controller{}

	rname := fmt.Sprintf("%s-%s-%s", config.ScannerRobotPrefix(context.TODO()), suite.registration.Name, "the-uuid-123")

	conf := map[string]interface{}{
		common.RobotTokenDuration: "30",
	}
	config.InitWithSettings(conf)

	account := &robot.Robot{
		Robot: model.Robot{
			Name:        rname,
			Description: "for scan",
			ProjectID:   suite.artifact.ProjectID,
			Duration:    -1,
		},
		Level: robot.LEVELPROJECT,
		Permissions: []*robot.Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: "repository",
						Action:   rbac.ActionPull,
					},
					{
						Resource: "repository",
						Action:   rbac.ActionScannerPull,
					},
				},
			},
		},
	}

	rc.On("Create", mock.Anything, account).Return(int64(1), "robot-account", nil)
	rc.On("Get", mock.Anything, int64(1), &robot.Option{
		WithPermission: false,
	}).Return(&robot.Robot{
		Robot: model.Robot{
			ID:          1,
			Name:        rname,
			Secret:      "robot-account",
			Description: "for scan",
			ProjectID:   suite.artifact.ProjectID,
			Duration:    -1,
		},
		Level: "project",
	}, nil)

	// Set job parameters
	req := &v1.ScanRequest{
		Registry: &v1.Registry{
			URL: "https://core.com",
		},
		Artifact: &v1.Artifact{
			NamespaceID: suite.artifact.ProjectID,
			Digest:      suite.artifact.Digest,
			Repository:  suite.artifact.RepositoryName,
			MimeType:    suite.artifact.ManifestMediaType,
		},
	}

	rJSON, err := req.ToJSON()
	require.NoError(suite.T(), err)

	regJSON, err := suite.registration.ToJSON()
	require.NoError(suite.T(), err)

	id, _, _ := rc.Create(context.TODO(), account)
	rb, _ := rc.Get(context.TODO(), id, &robot.Option{WithPermission: false})
	robotJSON, err := rb.ToJSON()
	require.NoError(suite.T(), err)

	params := make(map[string]interface{})
	params[sca.JobParamRegistration] = regJSON
	params[sca.JobParameterRequest] = rJSON
	params[sca.JobParameterMimes] = []string{v1.MimeTypeNativeReport}
	params[sca.JobParameterAuthType] = "Basic"
	params[sca.JobParameterRobot] = robotJSON

	suite.ar = &artifacttesting.Controller{}
	suite.accessoryMgr = &accessorytesting.Manager{}

	suite.tagCtl = &tagtesting.FakeController{}
	suite.tagCtl.On("List", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	suite.execMgr = &tasktesting.ExecutionManager{}

	suite.taskMgr = &tasktesting.Manager{}

	suite.cache = &mockcache.Cache{}

	suite.c = &basicController{
		manager: mgr,
		ar:      suite.ar,
		sc:      sc,
		rc:      rc,
		acc:     suite.accessoryMgr,
		tagCtl:  suite.tagCtl,
		uuid: func() (string, error) {
			return "the-uuid-123", nil
		},
		config: func(cfg string) (string, error) {
			switch cfg {
			case configRegistryEndpoint:
				return "https://core.com", nil
			case configCoreInternalAddr:
				return "http://core:8080", nil
			}

			return "", nil
		},

		cloneCtx: func(ctx context.Context) context.Context { return ctx },
		makeCtx:  func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },

		execMgr:         suite.execMgr,
		taskMgr:         suite.taskMgr,
		reportConverter: &postprocessorstesting.ScanReportV1ToV2Converter{},
		cache:           func() cache.Cache { return suite.cache },
	}
}

// TearDownSuite ...
func (suite *ControllerTestSuite) TearDownSuite() {
	artifact.Ctl = suite.originalArtifactCtl
}

// TestScanControllerScan ...
func (suite *ControllerTestSuite) TestScanControllerScan() {
	{
		// artifact not provieded
		suite.Require().Error(suite.c.Scan(context.TODO(), nil))
	}

	{
		mock.OnAnything(suite.accessoryMgr, "List").Return([]accessoryModel.Accessory{}, nil).Once()
		// success
		mock.OnAnything(suite.ar, "Walk").Return(nil).Run(func(args mock.Arguments) {
			walkFn := args.Get(2).(func(*artifact.Artifact) error)
			walkFn(suite.artifact)
		}).Once()

		mock.OnAnything(suite.taskMgr, "ListScanTasksByReportUUID").Return([]*task.Task{
			{ExtraAttrs: suite.makeExtraAttrs(int64(1), "rp-uuid-001"), Status: "Success"},
		}, nil).Once()

		mock.OnAnything(suite.reportMgr, "Delete").Return(nil).Once()

		mock.OnAnything(suite.execMgr, "Create").Return(int64(1), nil).Once()
		mock.OnAnything(suite.taskMgr, "Create").Return(int64(1), nil).Once()

		ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})

		suite.Require().NoError(suite.c.Scan(ctx, suite.artifact))
	}

	{
		mock.OnAnything(suite.accessoryMgr, "List").Return([]accessoryModel.Accessory{}, nil).Once()
		// delete old report failed
		mock.OnAnything(suite.ar, "Walk").Return(nil).Run(func(args mock.Arguments) {
			walkFn := args.Get(2).(func(*artifact.Artifact) error)
			walkFn(suite.artifact)
		}).Once()

		mock.OnAnything(suite.taskMgr, "ListScanTasksByReportUUID").Return([]*task.Task{
			{ExtraAttrs: suite.makeExtraAttrs(int64(1), "rp-uuid-001"), Status: "Success"},
		}, nil).Once()

		mock.OnAnything(suite.reportMgr, "Delete").Return(fmt.Errorf("delete failed")).Once()

		suite.Require().Error(suite.c.Scan(context.TODO(), suite.artifact))
	}

	{
		mock.OnAnything(suite.accessoryMgr, "List").Return([]accessoryModel.Accessory{}, nil).Once()
		// a previous scan process is ongoing
		mock.OnAnything(suite.ar, "Walk").Return(nil).Run(func(args mock.Arguments) {
			walkFn := args.Get(2).(func(*artifact.Artifact) error)
			walkFn(suite.artifact)
		}).Once()

		mock.OnAnything(suite.taskMgr, "ListScanTasksByReportUUID").Return([]*task.Task{
			{ExtraAttrs: suite.makeExtraAttrs(int64(1), "rp-uuid-001"), Status: "Running"},
		}, nil).Once()

		suite.Require().Error(suite.c.Scan(context.TODO(), suite.artifact))
	}
}

// TestScanControllerStop ...
func (suite *ControllerTestSuite) TestScanControllerStop() {
	{
		// artifact not provieded
		suite.Require().Error(suite.c.Stop(context.TODO(), nil))
	}

	{
		// success
		mock.OnAnything(suite.execMgr, "List").Return([]*task.Execution{
			{ExtraAttrs: suite.makeExtraAttrs(int64(1), "rp-uuid-001"), Status: "Running"},
		}, nil).Once()
		mock.OnAnything(suite.execMgr, "Stop").Return(nil).Once()

		ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})

		suite.Require().NoError(suite.c.Stop(ctx, suite.artifact))
	}

	{
		// failed due to no execution returned by List
		mock.OnAnything(suite.execMgr, "List").Return([]*task.Execution{}, nil).Once()
		mock.OnAnything(suite.execMgr, "Stop").Return(nil).Once()

		ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})

		suite.Require().Error(suite.c.Stop(ctx, suite.artifact))
	}

	{
		// failed due to execMgr.List() errored out
		mock.OnAnything(suite.execMgr, "List").Return([]*task.Execution{}, fmt.Errorf("failed to call execMgr.List()")).Once()

		ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})

		suite.Require().Error(suite.c.Stop(ctx, suite.artifact))
	}
}

// TestScanControllerGetReport ...
func (suite *ControllerTestSuite) TestScanControllerGetReport() {
	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	mock.OnAnything(suite.ar, "Walk").Return(nil).Run(func(args mock.Arguments) {
		walkFn := args.Get(2).(func(*artifact.Artifact) error)
		walkFn(suite.artifact)
	}).Once()

	mock.OnAnything(suite.taskMgr, "ListScanTasksByReportUUID").Return([]*task.Task{
		{ExtraAttrs: suite.makeExtraAttrs(int64(1), "rp-uuid-001")},
	}, nil).Once()
	mock.OnAnything(suite.accessoryMgr, "List").Return(nil, nil)
	rep, err := suite.c.GetReport(ctx, suite.artifact, []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(rep))
}

// TestScanControllerGetSummary ...
func (suite *ControllerTestSuite) TestScanControllerGetSummary() {
	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	mock.OnAnything(suite.accessoryMgr, "List").Return([]accessoryModel.Accessory{}, nil).Once()
	mock.OnAnything(suite.ar, "Walk").Return(nil).Run(func(args mock.Arguments) {
		walkFn := args.Get(2).(func(*artifact.Artifact) error)
		walkFn(suite.artifact)
	}).Once()
	mock.OnAnything(suite.taskMgr, "ListScanTasksByReportUUID").Return(nil, nil).Once()

	sum, err := suite.c.GetSummary(ctx, suite.artifact, []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(sum))
}

// TestScanControllerGetScanLog ...
func (suite *ControllerTestSuite) TestScanControllerGetScanLog() {
	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	mock.OnAnything(suite.taskMgr, "ListScanTasksByReportUUID").Return([]*task.Task{
		{
			ID:         1,
			ExtraAttrs: suite.makeExtraAttrs(int64(1), "rp-uuid-001"),
		},
	}, nil).Once()

	mock.OnAnything(suite.taskMgr, "GetLog").Return([]byte("log"), nil).Once()

	bytes, err := suite.c.GetScanLog(ctx, &artifact.Artifact{Artifact: art.Artifact{ID: 1, ProjectID: 1}}, "rp-uuid-001")
	require.NoError(suite.T(), err)
	assert.Condition(suite.T(), func() (success bool) {
		success = len(bytes) > 0
		return
	})
}

func (suite *ControllerTestSuite) TestScanControllerGetMultiScanLog() {
	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	suite.taskMgr.On("ListScanTasksByReportUUID", ctx, "rp-uuid-001").Return([]*task.Task{
		{
			ID:         1,
			ExtraAttrs: suite.makeExtraAttrs(int64(1), "rp-uuid-001"),
		},
	}, nil).Times(4)
	mock.OnAnything(suite.ar, "Walk").Return(nil).Run(func(args mock.Arguments) {
		walkFn := args.Get(2).(func(*artifact.Artifact) error)
		walkFn(suite.artifact)
	})
	mock.OnAnything(suite.accessoryMgr, "List").Return(nil, nil)
	suite.taskMgr.On("ListScanTasksByReportUUID", ctx, "rp-uuid-002").Return([]*task.Task{
		{
			ID:         2,
			ExtraAttrs: suite.makeExtraAttrs(int64(1), "rp-uuid-002"),
		},
	}, nil).Times(4)
	{
		// Both success
		mock.OnAnything(suite.taskMgr, "GetLog").Return([]byte("log"), nil).Twice()

		bytes, err := suite.c.GetScanLog(ctx, &artifact.Artifact{Artifact: art.Artifact{ID: 1, ProjectID: 1}}, base64.StdEncoding.EncodeToString([]byte("rp-uuid-001|rp-uuid-002")))
		suite.Nil(err)
		suite.NotEmpty(bytes)
		suite.Contains(string(bytes), "Logs of report rp-uuid-001")
		suite.Contains(string(bytes), "Logs of report rp-uuid-002")
	}

	{
		// One successfully, one failed
		suite.taskMgr.On("GetLog", ctx, int64(1)).Return([]byte("log"), nil).Once()
		suite.taskMgr.On("GetLog", ctx, int64(2)).Return(nil, fmt.Errorf("failed")).Once()

		bytes, err := suite.c.GetScanLog(ctx, &artifact.Artifact{Artifact: art.Artifact{ID: 1, ProjectID: 1}}, base64.StdEncoding.EncodeToString([]byte("rp-uuid-001|rp-uuid-002")))
		suite.Nil(err)
		suite.NotEmpty(bytes)
		suite.NotContains(string(bytes), "Logs of report rp-uuid-001")
	}

	{
		// Both failed
		mock.OnAnything(suite.taskMgr, "GetLog").Return(nil, fmt.Errorf("failed")).Twice()

		bytes, err := suite.c.GetScanLog(ctx, &artifact.Artifact{Artifact: art.Artifact{ID: 1, ProjectID: 1}}, base64.StdEncoding.EncodeToString([]byte("rp-uuid-001|rp-uuid-002")))
		suite.Error(err)
		suite.Empty(bytes)
	}

	{
		// Both empty
		mock.OnAnything(suite.taskMgr, "GetLog").Return(nil, nil).Twice()

		bytes, err := suite.c.GetScanLog(ctx, &artifact.Artifact{Artifact: art.Artifact{ID: 1, ProjectID: 1}}, base64.StdEncoding.EncodeToString([]byte("rp-uuid-001|rp-uuid-002")))
		suite.Nil(err)
		suite.Empty(bytes)
	}
}

func (suite *ControllerTestSuite) TestScanAll() {
	{
		// no artifacts found when scan all
		executionID := int64(1)

		suite.execMgr.On(
			"Create", mock.Anything, "SCAN_ALL", int64(0), "SCHEDULE",
			mock.Anything).Return(executionID, nil).Once()
		suite.execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{ID: executionID}, nil).Once()

		mock.OnAnything(suite.accessoryMgr, "List").Return([]accessoryModel.Accessory{}, nil).Once()

		mock.OnAnything(suite.artifactCtl, "List").Return([]*artifact.Artifact{}, nil).Once()

		suite.taskMgr.On("Count", mock.Anything, q.New(q.KeyWords{"execution_id": executionID})).Return(int64(0), nil).Once()

		mock.OnAnything(suite.execMgr, "UpdateExtraAttrs").Return(nil).Once()

		suite.execMgr.On("MarkDone", mock.Anything, executionID, mock.Anything).Return(nil).Once()

		suite.cache.On("Contains", mock.Anything, scanAllStoppedKey(1)).Return(false).Once()

		_, err := suite.c.ScanAll(context.TODO(), "SCHEDULE", false)
		suite.NoError(err)
	}

	{
		// artifacts found, but scan it failed when scan all
		ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})

		executionID := int64(1)

		suite.execMgr.On(
			"Create", mock.Anything, "SCAN_ALL", int64(0), "SCHEDULE",
			mock.Anything).Return(executionID, nil).Once()
		suite.execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{ID: executionID}, nil).Once()

		mock.OnAnything(suite.accessoryMgr, "List").Return([]accessoryModel.Accessory{}, nil).Once()

		mock.OnAnything(suite.artifactCtl, "List").Return([]*artifact.Artifact{suite.artifact}, nil).Once()
		mock.OnAnything(suite.ar, "Walk").Return(nil).Run(func(args mock.Arguments) {
			walkFn := args.Get(2).(func(*artifact.Artifact) error)
			walkFn(suite.artifact)
		}).Once()

		mock.OnAnything(suite.taskMgr, "ListScanTasksByReportUUID").Return(nil, nil).Once()

		mock.OnAnything(suite.reportMgr, "Delete").Return(nil).Once()
		mock.OnAnything(suite.reportMgr, "Create").Return("uuid", nil).Once()
		mock.OnAnything(suite.taskMgr, "Create").Return(int64(0), fmt.Errorf("failed")).Once()
		mock.OnAnything(suite.execMgr, "UpdateExtraAttrs").Return(nil).Once()
		suite.execMgr.On("MarkError", mock.Anything, executionID, mock.Anything).Return(nil).Once()

		_, err := suite.c.ScanAll(ctx, "SCHEDULE", false)
		suite.NoError(err)
	}
}

func (suite *ControllerTestSuite) TestStopScanAll() {
	mockExecID := int64(100)
	// mock error case
	mockErr := fmt.Errorf("stop scan all error")
	suite.cache.On("Save", mock.Anything, scanAllStoppedKey(mockExecID), mock.Anything, mock.Anything).Return(mockErr).Once()
	err := suite.c.StopScanAll(context.TODO(), mockExecID, false)
	suite.EqualError(err, mockErr.Error())

	// mock normal case
	suite.cache.On("Save", mock.Anything, scanAllStoppedKey(mockExecID), mock.Anything, mock.Anything).Return(nil).Once()
	suite.execMgr.On("Stop", mock.Anything, mockExecID).Return(nil).Once()
	err = suite.c.StopScanAll(context.TODO(), mockExecID, false)
	suite.NoError(err)
}

func (suite *ControllerTestSuite) TestDeleteReports() {
	suite.reportMgr.On("DeleteByDigests", context.TODO(), "digest").Return(nil).Once()

	suite.NoError(suite.c.DeleteReports(context.TODO(), "digest"))

	suite.reportMgr.On("DeleteByDigests", context.TODO(), "digest").Return(fmt.Errorf("failed")).Once()

	suite.Error(suite.c.DeleteReports(context.TODO(), "digest"))
}

func (suite *ControllerTestSuite) makeExtraAttrs(artifactID int64, reportUUIDs ...string) map[string]interface{} {
	b, _ := json.Marshal(map[string]interface{}{reportUUIDsKey: reportUUIDs})

	extraAttrs := map[string]interface{}{}
	json.Unmarshal(b, &extraAttrs)
	extraAttrs[artifactIDKey] = float64(artifactID)

	return extraAttrs
}
