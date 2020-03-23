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

	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/common"
	cj "github.com/goharbor/harbor/src/common/job"
	jm "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	sca "github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	artifacttesting "github.com/goharbor/harbor/src/testing/api/artifact"
	scannertesting "github.com/goharbor/harbor/src/testing/api/scanner"
	mocktesting "github.com/goharbor/harbor/src/testing/mock"
	reporttesting "github.com/goharbor/harbor/src/testing/pkg/scan/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ControllerTestSuite is the test suite for scan controller.
type ControllerTestSuite struct {
	suite.Suite

	registration *scanner.Registration
	artifact     *artifact.Artifact
	rawReport    string

	ar artifact.Controller
	c  Controller
}

// TestController is the entry point of ControllerTestSuite.
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// SetupSuite ...
func (suite *ControllerTestSuite) SetupSuite() {
	suite.artifact = &artifact.Artifact{}
	suite.artifact.Type = "IMAGE"
	suite.artifact.ProjectID = 1
	suite.artifact.RepositoryName = "library/photon"
	suite.artifact.Digest = "digest-code"
	suite.artifact.ManifestMediaType = v1.MimeTypeDockerArtifact

	m := &v1.ScannerAdapterMetadata{
		Scanner: &v1.Scanner{
			Name:    "Clair",
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
	sc.On("GetRegistrationByProject", suite.artifact.ProjectID).Return(suite.registration, nil)
	sc.On("Ping", suite.registration).Return(m, nil)

	mgr := &reporttesting.Manager{}
	mgr.On("Create", &scan.Report{
		Digest:           "digest-code",
		RegistrationUUID: "uuid001",
		MimeType:         "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0",
		Status:           "Pending",
		StatusCode:       0,
		TrackID:          "the-uuid-123",
		Requester:        "the-uuid-123",
	}).Return("r-uuid", nil)
	mgr.On("UpdateScanJobID", "the-uuid-123", "the-job-id").Return(nil)

	rp := vuln.Report{
		GeneratedAt: time.Now().UTC().String(),
		Scanner: &v1.Scanner{
			Name:    "Clair",
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
			StatusCode:       3,
			TrackID:          "the-uuid-123",
			JobID:            "the-job-id",
			StatusRevision:   time.Now().Unix(),
			Report:           suite.rawReport,
			StartTime:        time.Now(),
			EndTime:          time.Now().Add(2 * time.Second),
		},
	}

	mgr.On("GetBy", suite.artifact.Digest, suite.registration.UUID, []string{v1.MimeTypeNativeReport}).Return(reports, nil)
	mgr.On("Get", "rp-uuid-001").Return(reports[0], nil)
	mgr.On("UpdateReportData", "rp-uuid-001", suite.rawReport, (int64)(10000)).Return(nil)
	mgr.On("UpdateStatus", "the-uuid-123", "Success", (int64)(10000)).Return(nil)

	rc := &MockRobotController{}

	resource := fmt.Sprintf("/project/%d/repository", suite.artifact.ProjectID)
	access := []*types.Policy{
		{Resource: types.Resource(resource), Action: rbac.ActionPull},
		{Resource: types.Resource(resource), Action: rbac.ActionScannerPull},
	}

	rname := "the-uuid-123"
	account := &model.RobotCreate{
		Name:        rname,
		Description: "for scan",
		ProjectID:   suite.artifact.ProjectID,
		Access:      access,
	}
	rc.On("CreateRobotAccount", account).Return(&model.Robot{
		ID:          1,
		Name:        common.RobotPrefix + rname,
		Token:       "robot-account",
		Description: "for scan",
		ProjectID:   suite.artifact.ProjectID,
	}, nil)

	// Set job parameters
	req := &v1.ScanRequest{
		Registry: &v1.Registry{
			URL:           "https://core.com",
			Authorization: "Basic " + base64.StdEncoding.EncodeToString([]byte(common.RobotPrefix+"the-uuid-123:robot-account")),
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

	jc := &MockJobServiceClient{}
	params := make(map[string]interface{})
	params[sca.JobParamRegistration] = regJSON
	params[sca.JobParameterRequest] = rJSON
	params[sca.JobParameterMimes] = []string{v1.MimeTypeNativeReport}
	params[sca.JobParameterRobotID] = int64(1)

	j := &jm.JobData{
		Name: job.ImageScanJob,
		Metadata: &jm.JobMetadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/scan/%s", "http://core:8080", "the-uuid-123"),
	}
	jc.On("SubmitJob", j).Return("the-job-id", nil)
	jc.On("GetJobLog", "the-job-id").Return([]byte("job log"), nil)

	suite.ar = &artifacttesting.Controller{}
	mocktesting.OnAnything(suite.ar, "Walk").Return(nil).Run(func(args mock.Arguments) {
		walkFn := args.Get(2).(func(*artifact.Artifact) error)
		walkFn(suite.artifact)
	})

	suite.c = &basicController{
		manager: mgr,
		ar:      suite.ar,
		sc:      sc,
		jc: func() cj.Client {
			return jc
		},
		rc: rc,
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
	}
}

// TearDownSuite ...
func (suite *ControllerTestSuite) TearDownSuite() {}

// TestScanControllerScan ...
func (suite *ControllerTestSuite) TestScanControllerScan() {
	err := suite.c.Scan(context.TODO(), suite.artifact)
	require.NoError(suite.T(), err)
}

// TestScanControllerGetReport ...
func (suite *ControllerTestSuite) TestScanControllerGetReport() {
	rep, err := suite.c.GetReport(context.TODO(), suite.artifact, []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(rep))
}

// TestScanControllerGetSummary ...
func (suite *ControllerTestSuite) TestScanControllerGetSummary() {
	sum, err := suite.c.GetSummary(context.TODO(), suite.artifact, []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(sum))
}

// TestScanControllerGetScanLog ...
func (suite *ControllerTestSuite) TestScanControllerGetScanLog() {
	bytes, err := suite.c.GetScanLog("rp-uuid-001")
	require.NoError(suite.T(), err)
	assert.Condition(suite.T(), func() (success bool) {
		success = len(bytes) > 0
		return
	})
}

// TestScanControllerHandleJobHooks ...
func (suite *ControllerTestSuite) TestScanControllerHandleJobHooks() {
	cReport := &sca.CheckInReport{
		Digest:           "digest-code",
		RegistrationUUID: suite.registration.UUID,
		MimeType:         v1.MimeTypeNativeReport,
		RawReport:        suite.rawReport,
	}

	cRpJSON, err := cReport.ToJSON()
	require.NoError(suite.T(), err)

	statusChange := &job.StatusChange{
		JobID:   "the-job-id",
		Status:  "Success",
		CheckIn: string(cRpJSON),
		Metadata: &job.StatsInfo{
			Revision: (int64)(10000),
		},
	}

	err = suite.c.HandleJobHooks("the-uuid-123", statusChange)
	require.NoError(suite.T(), err)
}

// Mock things

// MockJobServiceClient ...
type MockJobServiceClient struct {
	mock.Mock
}

// SubmitJob ...
func (mjc *MockJobServiceClient) SubmitJob(jData *jm.JobData) (string, error) {
	args := mjc.Called(jData)

	return args.String(0), args.Error(1)
}

// GetJobLog ...
func (mjc *MockJobServiceClient) GetJobLog(uuid string) ([]byte, error) {
	args := mjc.Called(uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

// PostAction ...
func (mjc *MockJobServiceClient) PostAction(uuid, action string) error {
	args := mjc.Called(uuid, action)

	return args.Error(0)
}

func (mjc *MockJobServiceClient) GetExecutions(uuid string) ([]job.Stats, error) {
	args := mjc.Called(uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]job.Stats), args.Error(1)
}

// MockRobotController ...
type MockRobotController struct {
	mock.Mock
}

// GetRobotAccount ...
func (mrc *MockRobotController) GetRobotAccount(id int64) (*model.Robot, error) {
	args := mrc.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Robot), args.Error(1)
}

// CreateRobotAccount ...
func (mrc *MockRobotController) CreateRobotAccount(robotReq *model.RobotCreate) (*model.Robot, error) {
	args := mrc.Called(robotReq)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.Robot), args.Error(1)
}

// DeleteRobotAccount ...
func (mrc *MockRobotController) DeleteRobotAccount(id int64) error {
	args := mrc.Called(id)

	return args.Error(0)
}

// UpdateRobotAccount ...
func (mrc *MockRobotController) UpdateRobotAccount(r *model.Robot) error {
	args := mrc.Called(r)

	return args.Error(0)
}

// ListRobotAccount ...
func (mrc *MockRobotController) ListRobotAccount(query *q.Query) ([]*model.Robot, error) {
	args := mrc.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*model.Robot), args.Error(1)
}
