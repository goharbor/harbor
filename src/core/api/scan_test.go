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

package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/scan/api/scan"
	dscan "github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var scanBaseURL = "/api/repositories/library/hello-world/tags/latest/scan"

// ScanAPITestSuite is the test suite for scan API.
type ScanAPITestSuite struct {
	suite.Suite

	originalC scan.Controller
	c         *MockScanAPIController

	originalDigestGetter digestGetter

	artifact *v1.Artifact
}

// TestScanAPI is the entry point of ScanAPITestSuite.
func TestScanAPI(t *testing.T) {
	suite.Run(t, new(ScanAPITestSuite))
}

// SetupSuite prepares test env for suite.
func (suite *ScanAPITestSuite) SetupSuite() {
	suite.artifact = &v1.Artifact{
		NamespaceID: (int64)(1),
		Repository:  "library/hello-world",
		Tag:         "latest",
		Digest:      "digest-code-001",
		MimeType:    v1.MimeTypeDockerArtifact,
	}
}

// SetupTest prepares test env for test cases.
func (suite *ScanAPITestSuite) SetupTest() {
	suite.originalC = scan.DefaultController
	suite.c = &MockScanAPIController{}

	scan.DefaultController = suite.c

	suite.originalDigestGetter = digestFunc
	digestFunc = func(repo, tag string, username string) (s string, e error) {
		return "digest-code-001", nil
	}
}

// TearDownTest ...
func (suite *ScanAPITestSuite) TearDownTest() {
	scan.DefaultController = suite.originalC
	digestFunc = suite.originalDigestGetter
}

// TestScanAPIBase ...
func (suite *ScanAPITestSuite) TestScanAPIBase() {
	suite.c.On("Scan", &v1.Artifact{}).Return(nil)
	// Including general cases
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				url:    scanBaseURL,
				method: http.MethodGet,
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				url:        scanBaseURL,
				method:     http.MethodPost,
				credential: projGuest,
			},
			code: http.StatusForbidden,
		},
	}

	runCodeCheckingCases(suite.T(), cases...)
}

// TestScanAPIScan ...
func (suite *ScanAPITestSuite) TestScanAPIScan() {
	suite.c.On("Scan", suite.artifact).Return(nil)

	// Including general cases
	cases := []*codeCheckingCase{
		// 202
		{
			request: &testingRequest{
				url:        scanBaseURL,
				method:     http.MethodPost,
				credential: projDeveloper,
			},
			code: http.StatusAccepted,
		},
	}

	runCodeCheckingCases(suite.T(), cases...)
}

// TestScanAPIReport ...
func (suite *ScanAPITestSuite) TestScanAPIReport() {
	suite.c.On("GetReport", suite.artifact, []string{v1.MimeTypeNativeReport}).Return([]*dscan.Report{}, nil)

	vulItems := make(map[string]interface{})

	header := make(http.Header)
	header.Add("Accept", v1.MimeTypeNativeReport)
	err := handleAndParse(
		&testingRequest{
			url:        scanBaseURL,
			method:     http.MethodGet,
			credential: projDeveloper,
			header:     header,
		}, &vulItems)
	require.NoError(suite.T(), err)
}

// TestScanAPILog ...
func (suite *ScanAPITestSuite) TestScanAPILog() {
	suite.c.On("GetScanLog", "the-uuid-001").Return([]byte(`{"log": "this is my log"}`), nil)

	logs := make(map[string]string)
	err := handleAndParse(
		&testingRequest{
			url:        fmt.Sprintf("%s/%s", scanBaseURL, "the-uuid-001/log"),
			method:     http.MethodGet,
			credential: projDeveloper,
		}, &logs)
	require.NoError(suite.T(), err)
	assert.Condition(suite.T(), func() (success bool) {
		success = len(logs) > 0
		return
	})
}

// Mock things

// MockScanAPIController ...
type MockScanAPIController struct {
	mock.Mock
}

// Scan ...
func (msc *MockScanAPIController) Scan(artifact *v1.Artifact) error {
	args := msc.Called(artifact)

	return args.Error(0)
}

func (msc *MockScanAPIController) GetReport(artifact *v1.Artifact, mimeTypes []string) ([]*dscan.Report, error) {
	args := msc.Called(artifact, mimeTypes)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*dscan.Report), args.Error(1)
}

func (msc *MockScanAPIController) GetSummary(artifact *v1.Artifact, mimeTypes []string) (map[string]interface{}, error) {
	args := msc.Called(artifact, mimeTypes)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (msc *MockScanAPIController) GetScanLog(uuid string) ([]byte, error) {
	args := msc.Called(uuid)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func (msc *MockScanAPIController) HandleJobHooks(trackID string, change *job.StatusChange) error {
	args := msc.Called(trackID, change)

	return args.Error(0)
}
