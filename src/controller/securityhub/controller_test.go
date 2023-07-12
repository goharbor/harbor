//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package securityhub

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/securityhub/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/goharbor/harbor/src/testing/mock"
	artifactMock "github.com/goharbor/harbor/src/testing/pkg/artifact"
	scannerMock "github.com/goharbor/harbor/src/testing/pkg/scan/scanner"
	securityMock "github.com/goharbor/harbor/src/testing/pkg/securityhub"
)

var sum = &model.Summary{
	CriticalCnt: 50,
	HighCnt:     40,
	MediumCnt:   30,
	LowCnt:      20,
	NoneCnt:     10,
	FixableCnt:  90,
}

type ControllerTestSuite struct {
	htesting.Suite
	c           *controller
	artifactMgr *artifactMock.Manager
	scannerMgr  *scannerMock.Manager
	secHubMgr   *securityMock.Manager
}

// TestController is the entry of controller test suite
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// SetupTest prepares env for the controller test suite
func (suite *ControllerTestSuite) SetupTest() {
	suite.artifactMgr = &artifactMock.Manager{}
	suite.secHubMgr = &securityMock.Manager{}
	suite.scannerMgr = &scannerMock.Manager{}
	suite.c = &controller{
		artifactMgr: suite.artifactMgr,
		secHubMgr:   suite.secHubMgr,
		scannerMgr:  suite.scannerMgr,
	}
}

func (suite *ControllerTestSuite) TearDownTest() {
}

// TestSecuritySummary tests the security summary
func (suite *ControllerTestSuite) TestSecuritySummary() {
	ctx := suite.Context()

	mock.OnAnything(suite.artifactMgr, "Count").Return(int64(1234), nil)
	mock.OnAnything(suite.secHubMgr, "ScannedArtifactsCount").Return(int64(1000), nil)
	mock.OnAnything(suite.secHubMgr, "Summary").Return(sum, nil).Twice()
	mock.OnAnything(suite.scannerMgr, "GetDefault").Return(&scanner.Registration{UUID: "ruuid"}, nil)
	summary, err := suite.c.SecuritySummary(ctx, 0, WithArtifact(false), WithCVE(false))
	suite.NoError(err)
	suite.NotNil(summary)
	suite.Equal(int64(1234), summary.TotalArtifactCnt)
	suite.Equal(int64(1000), summary.ScannedCnt)
	suite.Equal(int64(50), summary.CriticalCnt)
	suite.Equal(int64(40), summary.HighCnt)
	suite.Equal(int64(30), summary.MediumCnt)
	suite.Equal(int64(20), summary.LowCnt)
	suite.Equal(int64(10), summary.NoneCnt)
	suite.Equal(int64(90), summary.FixableCnt)
	sum.DangerousCVEs = []*scan.VulnerabilityRecord{
		{CVEID: "CVE-2020-1234", Severity: "CRITICAL"},
		{CVEID: "CVE-2020-1235", Severity: "HIGH"},
		{CVEID: "CVE-2020-1236", Severity: "MEDIUM"},
		{CVEID: "CVE-2020-1237", Severity: "LOW"},
		{CVEID: "CVE-2020-1238", Severity: "NONE"},
	}
	sum.DangerousArtifacts = []*model.DangerousArtifact{
		{Project: 1, Repository: "library/busybox"},
		{Project: 1, Repository: "library/nginx"},
		{Project: 1, Repository: "library/hello-world"},
		{Project: 1, Repository: "library/harbor-jobservice"},
		{Project: 1, Repository: "library/harbor-core"},
	}
	mock.OnAnything(suite.secHubMgr, "Summary").Return(sum, nil).Once()
	mock.OnAnything(suite.secHubMgr, "DangerousCVEs").Return(sum.DangerousCVEs, nil).Once()
	mock.OnAnything(suite.secHubMgr, "DangerousArtifacts").Return(sum.DangerousArtifacts, nil).Once()
	sum2, err := suite.c.SecuritySummary(ctx, 0, WithCVE(false), WithArtifact(false))
	suite.NoError(err)
	suite.NotNil(sum2)
	suite.NotNil(sum2.DangerousCVEs)
	suite.NotNil(sum2.DangerousArtifacts)

	sum3, err := suite.c.SecuritySummary(ctx, 0, WithCVE(true), WithArtifact(true))
	suite.NoError(err)
	suite.NotNil(sum3)
	suite.True(len(sum3.DangerousCVEs) > 0)
	suite.True(len(sum3.DangerousArtifacts) > 0)
}

// TestSecuritySummaryError tests the security summary with error
func (suite *ControllerTestSuite) TestSecuritySummaryError() {
	ctx := suite.Context()
	mock.OnAnything(suite.scannerMgr, "GetDefault").Return(&scanner.Registration{UUID: "ruuid"}, nil)
	mock.OnAnything(suite.secHubMgr, "ScannedArtifactsCount").Return(int64(1000), nil)
	mock.OnAnything(suite.secHubMgr, "Summary").Return(nil, errors.New("invalid project")).Once()
	summary, err := suite.c.SecuritySummary(ctx, 0, WithCVE(false), WithArtifact(false))
	suite.Error(err)
	suite.Nil(summary)
	mock.OnAnything(suite.artifactMgr, "Count").Return(int64(0), errors.New("failed to connect db")).Once()
	mock.OnAnything(suite.secHubMgr, "Summary").Return(sum, nil).Once()
	summary, err = suite.c.SecuritySummary(ctx, 0, WithCVE(false), WithArtifact(false))
	suite.Error(err)
	suite.Nil(summary)

}

// TestGetDefaultScanner tests the get default scanner
func (suite *ControllerTestSuite) TestGetDefaultScanner() {
	ctx := suite.Context()
	mock.OnAnything(suite.scannerMgr, "GetDefault").Return(&scanner.Registration{UUID: ""}, nil).Once()
	scanner, err := suite.c.defaultScannerUUID(ctx)
	suite.NoError(err)
	suite.Equal("", scanner)

	mock.OnAnything(suite.scannerMgr, "GetDefault").Return(nil, errors.New("failed to get scanner")).Once()
	scanner, err = suite.c.defaultScannerUUID(ctx)
	suite.Error(err)
	suite.Equal("", scanner)
}

func (suite *ControllerTestSuite) TestScannedArtifact() {
	ctx := suite.Context()
	mock.OnAnything(suite.scannerMgr, "GetDefault").Return(&scanner.Registration{UUID: "ruuid"}, nil)
	mock.OnAnything(suite.secHubMgr, "ScannedArtifactsCount").Return(int64(1000), nil)
	scanned, err := suite.c.scannedArtifactCount(ctx, 0)
	suite.NoError(err)
	suite.Equal(int64(1000), scanned)
}
