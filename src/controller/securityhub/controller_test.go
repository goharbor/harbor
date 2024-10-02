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
	"github.com/goharbor/harbor/src/pkg/securityhub/model"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/goharbor/harbor/src/testing/mock"
	scannerMock "github.com/goharbor/harbor/src/testing/pkg/scan/scanner"
	securityMock "github.com/goharbor/harbor/src/testing/pkg/securityhub"
	tagMock "github.com/goharbor/harbor/src/testing/pkg/tag"
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
	c          *controller
	scannerMgr *scannerMock.Manager
	secHubMgr  *securityMock.Manager
	tagMgr     *tagMock.Manager
}

// TestController is the entry of controller test suite
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// SetupTest prepares env for the controller test suite
func (suite *ControllerTestSuite) SetupTest() {
	suite.secHubMgr = &securityMock.Manager{}
	suite.scannerMgr = &scannerMock.Manager{}
	suite.tagMgr = &tagMock.Manager{}

	suite.c = &controller{
		secHubMgr:  suite.secHubMgr,
		scannerMgr: suite.scannerMgr,
		tagMgr:     suite.tagMgr,
	}
}

func (suite *ControllerTestSuite) TearDownTest() {
}

// TestSecuritySummary tests the security summary
func (suite *ControllerTestSuite) TestSecuritySummary() {
	ctx := suite.Context()

	mock.OnAnything(suite.secHubMgr, "TotalArtifactsCount").Return(int64(1234), nil)
	mock.OnAnything(suite.secHubMgr, "ScannedArtifactsCount").Return(int64(1000), nil)
	mock.OnAnything(suite.secHubMgr, "Summary").Return(sum, nil).Twice()
	mock.OnAnything(suite.scannerMgr, "DefaultScannerUUID").Return("ruuid", nil)
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
	mock.OnAnything(suite.scannerMgr, "DefaultScannerUUID").Return("ruuid", nil)
	mock.OnAnything(suite.secHubMgr, "TotalArtifactsCount").Return(int64(0), errors.New("project not found")).Once()
	mock.OnAnything(suite.secHubMgr, "ScannedArtifactsCount").Return(int64(1000), nil)
	mock.OnAnything(suite.secHubMgr, "Summary").Return(nil, errors.New("invalid project")).Once()
	summary, err := suite.c.SecuritySummary(ctx, 0, WithCVE(false), WithArtifact(false))
	suite.Error(err)
	suite.Nil(summary)
	mock.OnAnything(suite.secHubMgr, "TotalArtifactsCount").Return(int64(0), errors.New("failed to connect db")).Once()
	mock.OnAnything(suite.secHubMgr, "Summary").Return(sum, nil).Once()
	summary, err = suite.c.SecuritySummary(ctx, 0, WithCVE(false), WithArtifact(false))
	suite.Error(err)
	suite.Nil(summary)

}

func (suite *ControllerTestSuite) TestScannedArtifact() {
	ctx := suite.Context()
	mock.OnAnything(suite.scannerMgr, "DefaultScannerUUID").Return("ruuid", nil)
	mock.OnAnything(suite.secHubMgr, "ScannedArtifactsCount").Return(int64(1000), nil)
	scanned, err := suite.c.scannedArtifactCount(ctx, 0)
	suite.NoError(err)
	suite.Equal(int64(1000), scanned)
}

// TestAttachTags test the attachTags
func (suite *ControllerTestSuite) TestAttachTags() {
	ctx := suite.Context()
	tagList := []*tag.Tag{
		{ArtifactID: int64(1), Name: "latest"},
		{ArtifactID: int64(1), Name: "tag1"},
		{ArtifactID: int64(1), Name: "tag2"},
		{ArtifactID: int64(1), Name: "tag3"},
		{ArtifactID: int64(1), Name: "tag4"},
		{ArtifactID: int64(1), Name: "tag5"},
		{ArtifactID: int64(1), Name: "tag6"},
		{ArtifactID: int64(1), Name: "tag7"},
		{ArtifactID: int64(1), Name: "tag8"},
		{ArtifactID: int64(1), Name: "tag9"},
		{ArtifactID: int64(1), Name: "tag10"},
	}
	vulItems := []*model.VulnerabilityItem{
		{ArtifactID: int64(1)},
	}
	mock.OnAnything(suite.c.tagMgr, "List").Return(tagList, nil).Once()
	resultItems, err := suite.c.attachTags(ctx, vulItems)
	suite.NoError(err)
	suite.Equal(len(vulItems), len(resultItems))
	suite.Equal([]string{"latest"}, resultItems[0].Tags[:1])
	suite.Equal(10, len(resultItems[0].Tags))
}

// TestListVuls tests the list vulnerabilities
func (suite *ControllerTestSuite) TestListVuls() {
	ctx := suite.Context()
	vulItems := []*model.VulnerabilityItem{
		{ArtifactID: int64(1)},
	}
	tagList := []*tag.Tag{
		{ArtifactID: int64(1), Name: "latest"},
	}
	mock.OnAnything(suite.c.secHubMgr, "ListVuls").Return(vulItems, nil)
	mock.OnAnything(suite.c.tagMgr, "List").Return(tagList, nil).Once()
	vulResult, err := suite.c.ListVuls(ctx, "", 0, true, nil)
	suite.NoError(err)
	suite.Equal(1, len(vulResult))
	suite.Equal(int64(1), vulResult[0].ArtifactID)
}

func (suite *ControllerTestSuite) TestCountVuls() {
	ctx := suite.Context()
	mock.OnAnything(suite.c.secHubMgr, "TotalVuls").Return(int64(10), nil)
	count, err := suite.c.CountVuls(ctx, "", 0, true, nil)
	suite.NoError(err)
	suite.Equal(int64(10), count)
}
