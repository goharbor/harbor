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
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// JobTestSuite is a test suite to test the scan job.
type JobTestSuite struct {
	suite.Suite

	defaultClientPool v1.ClientPool
	mcp               *MockClientPool
}

// TestJob is the entry of JobTestSuite.
func TestJob(t *testing.T) {
	suite.Run(t, &JobTestSuite{})
}

// SetupSuite sets up test env for JobTestSuite.
func (suite *JobTestSuite) SetupSuite() {
	mcp := &MockClientPool{}
	suite.defaultClientPool = v1.DefaultClientPool
	v1.DefaultClientPool = mcp

	suite.mcp = mcp
}

// TeraDownSuite clears test env for TeraDownSuite.
func (suite *JobTestSuite) TeraDownSuite() {
	v1.DefaultClientPool = suite.defaultClientPool
}

// TestJob tests the scan job
func (suite *JobTestSuite) TestJob() {
	ctx := &MockJobContext{}
	lg := &MockJobLogger{}

	ctx.On("GetLogger").Return(lg)

	r := &scanner.Registration{
		ID:   0,
		UUID: "uuid",
		Name: "TestJob",
		URL:  "https://clair.com:8080",
	}

	rData, err := r.ToJSON()
	require.NoError(suite.T(), err)

	sr := &v1.ScanRequest{
		Registry: &v1.Registry{
			URL:           "http://localhost:5000",
			Authorization: "Basic cm9ib3Q6dG9rZW4=",
		},
		Artifact: &v1.Artifact{
			Repository: "library/test_job",
			Digest:     "sha256:data",
			MimeType:   v1.MimeTypeDockerArtifact,
		},
	}

	sData, err := sr.ToJSON()
	require.NoError(suite.T(), err)

	robot := &model.Robot{
		ID:    1,
		Name:  "robot",
		Token: "token",
	}

	robotData, err := robot.ToJSON()
	require.NoError(suite.T(), err)

	mimeTypes := []string{v1.MimeTypeNativeReport}

	jp := make(job.Parameters)
	jp[JobParamRegistration] = rData
	jp[JobParameterRequest] = sData
	jp[JobParameterMimes] = mimeTypes
	jp[JobParameterAuthType] = "Basic"
	jp[JobParameterRobot] = robotData

	mc := &MockClient{}
	sre := &v1.ScanResponse{
		ID: "scan_id",
	}
	mc.On("SubmitScan", sr).Return(sre, nil)

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

	jRep, err := json.Marshal(rp)
	require.NoError(suite.T(), err)

	mc.On("GetScanReport", "scan_id", v1.MimeTypeNativeReport).Return(string(jRep), nil)
	suite.mcp.On("Get", r).Return(mc, nil)

	crp := &CheckInReport{
		Digest:           sr.Artifact.Digest,
		RegistrationUUID: r.UUID,
		MimeType:         v1.MimeTypeNativeReport,
		RawReport:        string(jRep),
	}

	jsonData, err := crp.ToJSON()
	require.NoError(suite.T(), err)

	ctx.On("Checkin", string(jsonData)).Return(nil)
	j := &Job{}
	err = j.Run(ctx, jp)
	require.NoError(suite.T(), err)
}

// MockJobContext mocks job context interface.
// TODO: Maybe moved to a separate `mock` pkg for sharing in future.
type MockJobContext struct {
	mock.Mock
}

// Build ...
func (mjc *MockJobContext) Build(tracker job.Tracker) (job.Context, error) {
	args := mjc.Called(tracker)
	c := args.Get(0)
	if c != nil {
		return c.(job.Context), nil
	}

	return nil, args.Error(1)
}

// Get ...
func (mjc *MockJobContext) Get(prop string) (interface{}, bool) {
	args := mjc.Called(prop)
	return args.Get(0), args.Bool(1)
}

// SystemContext ...
func (mjc *MockJobContext) SystemContext() context.Context {
	return context.TODO()
}

// Checkin ...
func (mjc *MockJobContext) Checkin(status string) error {
	args := mjc.Called(status)
	return args.Error(0)
}

// OPCommand ...
func (mjc *MockJobContext) OPCommand() (job.OPCommand, bool) {
	args := mjc.Called()
	return (job.OPCommand)(args.String(0)), args.Bool(1)
}

// GetLogger ...
func (mjc *MockJobContext) GetLogger() logger.Interface {
	return &MockJobLogger{}
}

// Tracker ...
func (mjc *MockJobContext) Tracker() job.Tracker {
	args := mjc.Called()
	if t := args.Get(0); t != nil {
		return t.(job.Tracker)
	}

	return nil
}

// MockJobLogger mocks the job logger interface.
// TODO: Maybe moved to a separate `mock` pkg for sharing in future.
type MockJobLogger struct {
	mock.Mock
}

// Debug ...
func (mjl *MockJobLogger) Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Debugf ...
func (mjl *MockJobLogger) Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Info ...
func (mjl *MockJobLogger) Info(v ...interface{}) {
	logger.Info(v...)
}

// Infof ...
func (mjl *MockJobLogger) Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Warning ...
func (mjl *MockJobLogger) Warning(v ...interface{}) {
	logger.Warning(v...)
}

// Warningf ...
func (mjl *MockJobLogger) Warningf(format string, v ...interface{}) {
	logger.Warningf(format, v...)
}

// Error ...
func (mjl *MockJobLogger) Error(v ...interface{}) {
	logger.Error(v...)
}

// Errorf ...
func (mjl *MockJobLogger) Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Fatal ...
func (mjl *MockJobLogger) Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Fatalf ...
func (mjl *MockJobLogger) Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// MockClientPool mocks the client pool
type MockClientPool struct {
	mock.Mock
}

// Get v1 client
func (mcp *MockClientPool) Get(r *scanner.Registration) (v1.Client, error) {
	args := mcp.Called(r)
	c := args.Get(0)
	if c != nil {
		return c.(v1.Client), nil
	}

	return nil, args.Error(1)
}

// MockClient mocks the v1 client
type MockClient struct {
	mock.Mock
}

// GetMetadata ...
func (mc *MockClient) GetMetadata() (*v1.ScannerAdapterMetadata, error) {
	args := mc.Called()
	s := args.Get(0)
	if s != nil {
		return s.(*v1.ScannerAdapterMetadata), nil
	}

	return nil, args.Error(1)
}

// SubmitScan ...
func (mc *MockClient) SubmitScan(req *v1.ScanRequest) (*v1.ScanResponse, error) {
	args := mc.Called(req)
	sr := args.Get(0)
	if sr != nil {
		return sr.(*v1.ScanResponse), nil
	}

	return nil, args.Error(1)
}

// GetScanReport ...
func (mc *MockClient) GetScanReport(scanRequestID, reportMIMEType string) (string, error) {
	args := mc.Called(scanRequestID, reportMIMEType)
	return args.String(0), args.Error(1)
}
