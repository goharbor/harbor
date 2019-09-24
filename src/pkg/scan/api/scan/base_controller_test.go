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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/job/models"
	jm "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/q"
	sca "github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ControllerTestSuite is the test suite for scan controller.
type ControllerTestSuite struct {
	suite.Suite

	registration *scanner.Registration
	artifact     *v1.Artifact
	rawReport    string
	c            Controller
}

// TestController is the entry point of ControllerTestSuite.
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// SetupSuite ...
func (suite *ControllerTestSuite) SetupSuite() {
	suite.registration = &scanner.Registration{
		ID:        1,
		UUID:      "uuid001",
		Name:      "Test-scan-controller",
		URL:       "http://testing.com:3128",
		IsDefault: true,
	}

	suite.artifact = &v1.Artifact{
		NamespaceID: 1,
		Repository:  "scan",
		Tag:         "golang",
		Digest:      "digest-code",
		MimeType:    v1.MimeTypeDockerArtifact,
	}

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

	sc := &MockScannerController{}
	sc.On("GetRegistrationByProject", suite.artifact.NamespaceID).Return(suite.registration, nil)
	sc.On("Ping", suite.registration).Return(m, nil)

	mgr := &MockReportManager{}
	mgr.On("Create", &scan.Report{
		Digest:           "digest-code",
		RegistrationUUID: "uuid001",
		MimeType:         "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0",
		Status:           "Pending",
		StatusCode:       0,
		TrackID:          "the-uuid-123",
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

	dep := &MockDepManager{}
	dep.On("UUID").Return("the-uuid-123", nil)
	dep.On("GetRegistryEndpoint").Return("https://core.com", nil)
	dep.On("GetInternalCoreAddr").Return("http://core:8080", nil)
	dep.On("MakeRobotAccount", suite.artifact.NamespaceID, (int64)(1800)).Return("robot-account", nil)
	dep.On("GetJobLog", "the-job-id").Return([]byte("job log"), nil)

	// Set job parameters
	req := &v1.ScanRequest{
		Registry: &v1.Registry{
			URL:           "https://core.com",
			Authorization: "robot-account",
		},
		Artifact: suite.artifact,
	}

	rJSON, err := req.ToJSON()
	require.NoError(suite.T(), err)

	regJSON, err := suite.registration.ToJSON()
	require.NoError(suite.T(), err)

	params := make(map[string]interface{})
	params[sca.JobParamRegistration] = regJSON
	params[sca.JobParameterRequest] = rJSON
	params[sca.JobParameterMimes] = []string{v1.MimeTypeNativeReport}

	j := &jm.JobData{
		Name: job.ImageScanJob,
		Metadata: &jm.JobMetadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/scan/%s", "http://core:8080", "the-uuid-123"),
	}
	dep.On("SubmitJob", j).Return("the-job-id", nil)

	suite.c = &basicController{
		manager: mgr,
		sc:      sc,
		dep:     dep,
	}
}

// TearDownSuite ...
func (suite *ControllerTestSuite) TearDownSuite() {}

// TestScanControllerScan ...
func (suite *ControllerTestSuite) TestScanControllerScan() {
	err := suite.c.Scan(suite.artifact)
	require.NoError(suite.T(), err)
}

// TestScanControllerGetReport ...
func (suite *ControllerTestSuite) TestScanControllerGetReport() {
	rep, err := suite.c.GetReport(suite.artifact, []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(rep))
}

// TestScanControllerGetSummary ...
func (suite *ControllerTestSuite) TestScanControllerGetSummary() {
	sum, err := suite.c.GetSummary(suite.artifact, []string{v1.MimeTypeNativeReport})
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

// MockReportManager ...
type MockReportManager struct {
	mock.Mock
}

// Create ...
func (mrm *MockReportManager) Create(r *scan.Report) (string, error) {
	args := mrm.Called(r)

	return args.String(0), args.Error(1)
}

// UpdateScanJobID ...
func (mrm *MockReportManager) UpdateScanJobID(trackID string, jobID string) error {
	args := mrm.Called(trackID, jobID)

	return args.Error(0)
}

func (mrm *MockReportManager) UpdateStatus(trackID string, status string, rev int64) error {
	args := mrm.Called(trackID, status, rev)

	return args.Error(0)
}

func (mrm *MockReportManager) UpdateReportData(uuid string, report string, rev int64) error {
	args := mrm.Called(uuid, report, rev)

	return args.Error(0)
}

func (mrm *MockReportManager) GetBy(digest string, registrationUUID string, mimeTypes []string) ([]*scan.Report, error) {
	args := mrm.Called(digest, registrationUUID, mimeTypes)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*scan.Report), args.Error(1)
}

func (mrm *MockReportManager) Get(uuid string) (*scan.Report, error) {
	args := mrm.Called(uuid)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*scan.Report), args.Error(1)
}

// MockScannerController ...
type MockScannerController struct {
	mock.Mock
}

// ListRegistrations ...
func (msc *MockScannerController) ListRegistrations(query *q.Query) ([]*scanner.Registration, error) {
	args := msc.Called(query)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*scanner.Registration), args.Error(1)
}

// CreateRegistration ...
func (msc *MockScannerController) CreateRegistration(registration *scanner.Registration) (string, error) {
	args := msc.Called(registration)

	return args.String(0), args.Error(1)
}

// GetRegistration ...
func (msc *MockScannerController) GetRegistration(registrationUUID string) (*scanner.Registration, error) {
	args := msc.Called(registrationUUID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*scanner.Registration), args.Error(1)
}

// RegistrationExists ...
func (msc *MockScannerController) RegistrationExists(registrationUUID string) bool {
	args := msc.Called(registrationUUID)

	return args.Bool(0)
}

// UpdateRegistration ...
func (msc *MockScannerController) UpdateRegistration(registration *scanner.Registration) error {
	args := msc.Called(registration)

	return args.Error(0)
}

// DeleteRegistration ...
func (msc *MockScannerController) DeleteRegistration(registrationUUID string) (*scanner.Registration, error) {
	args := msc.Called(registrationUUID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*scanner.Registration), args.Error(1)
}

// SetDefaultRegistration ...
func (msc *MockScannerController) SetDefaultRegistration(registrationUUID string) error {
	args := msc.Called(registrationUUID)

	return args.Error(0)
}

// SetRegistrationByProject ...
func (msc *MockScannerController) SetRegistrationByProject(projectID int64, scannerID string) error {
	args := msc.Called(projectID, scannerID)

	return args.Error(0)
}

// GetRegistrationByProject ...
func (msc *MockScannerController) GetRegistrationByProject(projectID int64) (*scanner.Registration, error) {
	args := msc.Called(projectID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*scanner.Registration), args.Error(1)
}

// Ping ...
func (msc *MockScannerController) Ping(registration *scanner.Registration) (*v1.ScannerAdapterMetadata, error) {
	args := msc.Called(registration)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*v1.ScannerAdapterMetadata), args.Error(1)
}

// GetMetadata ...
func (msc *MockScannerController) GetMetadata(registrationUUID string) (*v1.ScannerAdapterMetadata, error) {
	args := msc.Called(registrationUUID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*v1.ScannerAdapterMetadata), args.Error(1)
}

// MockDepManager ...
type MockDepManager struct {
	mock.Mock
}

// UUID ...
func (mdm *MockDepManager) UUID() (string, error) {
	args := mdm.Called()

	return args.String(0), args.Error(1)
}

func (mdm *MockDepManager) SubmitJob(jobData *models.JobData) (string, error) {
	args := mdm.Called(jobData)

	return args.String(0), args.Error(1)
}

func (mdm *MockDepManager) GetRegistryEndpoint() (string, error) {
	args := mdm.Called()

	return args.String(0), args.Error(1)
}

func (mdm *MockDepManager) GetInternalCoreAddr() (string, error) {
	args := mdm.Called()

	return args.String(0), args.Error(1)
}

// MakeRobotAccount ...
func (mdm *MockDepManager) MakeRobotAccount(pid int64, ttl int64) (string, error) {
	args := mdm.Called(pid, ttl)

	return args.String(0), args.Error(1)
}

// GetJobLog ...
func (mdm *MockDepManager) GetJobLog(uuid string) ([]byte, error) {
	args := mdm.Called(uuid)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}
