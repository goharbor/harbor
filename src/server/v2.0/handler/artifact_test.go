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

package handler

import (
	"testing"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	scantesting "github.com/goharbor/harbor/src/testing/controller/scan"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestParse(t *testing.T) {
	// with tag
	input := "library/hello-world:latest"
	repository, reference, err := parse(input)
	require.Nil(t, err)
	assert.Equal(t, "library/hello-world", repository)
	assert.Equal(t, "latest", reference)

	// with digest
	input = "library/hello-world@sha256:9572f7cdcee8591948c2963463447a53466950b3fc15a247fcad1917ca215a2f"
	repository, reference, err = parse(input)
	require.Nil(t, err)
	assert.Equal(t, "library/hello-world", repository)
	assert.Equal(t, "sha256:9572f7cdcee8591948c2963463447a53466950b3fc15a247fcad1917ca215a2f", reference)

	// invalid digest
	input = "library/hello-world@sha256:invalid_digest"
	repository, reference, err = parse(input)
	require.NotNil(t, err)

	// invalid character
	input = "library/hello-world?#:latest"
	repository, reference, err = parse(input)
	require.NotNil(t, err)

	// empty input
	input = ""
	repository, reference, err = parse(input)
	require.NotNil(t, err)
}

type ArtifactTestSuite struct {
	htesting.Suite

	artCtl  *artifacttesting.Controller
	scanCtl *scantesting.Controller

	report1 *scan.Report
	report2 *scan.Report
}

func (suite *ArtifactTestSuite) SetupSuite() {
	suite.artCtl = &artifacttesting.Controller{}
	suite.scanCtl = &scantesting.Controller{}

	suite.Config = &restapi.Config{
		ArtifactAPI: &artifactAPI{
			artCtl:  suite.artCtl,
			scanCtl: suite.scanCtl,
		},
	}

	suite.Suite.SetupSuite()

	mock.OnAnything(projectCtlMock, "GetByName").Return(&project.Project{ProjectID: 1}, nil)

	suite.report1 = &scan.Report{
		MimeType: v1.MimeTypeNativeReport,
		Report:   "{}",
	}

	suite.report2 = &scan.Report{
		MimeType: v1.MimeTypeGenericVulnerabilityReport,
		Report:   "{}",
	}
}

func (suite *ArtifactTestSuite) onGetReport(mimeType string, reports ...*scan.Report) {
	suite.scanCtl.On("GetReport", mock.Anything, mock.Anything, []string{mimeType}).Return(reports, nil).Once()
}

func (suite *ArtifactTestSuite) TestGetVulnerabilitiesAddition() {
	times := 6
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("IsSysAdmin").Return(true).Times(times)
	mock.OnAnything(suite.Security, "Can").Return(true).Times(times)
	mock.OnAnything(suite.artCtl, "GetByReference").Return(&artifact.Artifact{}, nil).Times(times)

	url := "/projects/library/repositories/photon/artifacts/2.0/additions/vulnerabilities"

	{
		// report not found for the default X-Accept-Vulnerabilities
		suite.onGetReport(v1.MimeTypeNativeReport)

		var body map[string]interface{}
		res, err := suite.GetJSON(url, &body)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Empty(body)
	}

	{
		// report found for the default X-Accept-Vulnerabilities
		suite.onGetReport(v1.MimeTypeNativeReport, suite.report1)

		var body map[string]interface{}
		res, err := suite.GetJSON(url, &body)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.NotEmpty(body)
		suite.Contains(body, v1.MimeTypeNativeReport)
	}

	{
		// report found for the X-Accept-Vulnerabilities of "application/vnd.security.vulnerability.report; version=1.1"
		suite.onGetReport(v1.MimeTypeGenericVulnerabilityReport, suite.report2)

		var body map[string]interface{}
		res, err := suite.GetJSON(url, &body, map[string]string{"X-Accept-Vulnerabilities": v1.MimeTypeGenericVulnerabilityReport})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.NotEmpty(body)
		suite.Contains(body, v1.MimeTypeGenericVulnerabilityReport)
	}

	{
		// report found for "application/vnd.security.vulnerability.report; version=1.1"
		// and the X-Accept-Vulnerabilities is "application/vnd.security.vulnerability.report; version=1.1, application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"
		suite.onGetReport(v1.MimeTypeGenericVulnerabilityReport, suite.report2)

		var body map[string]interface{}
		res, err := suite.GetJSON(url, &body, map[string]string{"X-Accept-Vulnerabilities": v1.MimeTypeGenericVulnerabilityReport + "," + v1.MimeTypeNativeReport})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.NotEmpty(body)
		suite.Contains(body, v1.MimeTypeGenericVulnerabilityReport)
	}

	{
		// report not found for "application/vnd.security.vulnerability.report; version=1.1"
		// report found for "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"
		// and the X-Accept-Vulnerabilities is "application/vnd.security.vulnerability.report; version=1.1, application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"
		suite.onGetReport(v1.MimeTypeGenericVulnerabilityReport)
		suite.onGetReport(v1.MimeTypeNativeReport, suite.report1)

		var body map[string]interface{}
		res, err := suite.GetJSON(url, &body, map[string]string{"X-Accept-Vulnerabilities": v1.MimeTypeGenericVulnerabilityReport + "," + v1.MimeTypeNativeReport})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.NotEmpty(body)
		suite.Contains(body, v1.MimeTypeNativeReport)
	}

	{
		// report not found for "application/vnd.security.vulnerability.report; version=1.1"
		// report not found for "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"
		// and the X-Accept-Vulnerabilities is "application/vnd.security.vulnerability.report; version=1.1, application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"
		suite.onGetReport(v1.MimeTypeGenericVulnerabilityReport)
		suite.onGetReport(v1.MimeTypeNativeReport)

		var body map[string]interface{}
		res, err := suite.GetJSON(url, &body, map[string]string{"X-Accept-Vulnerabilities": v1.MimeTypeGenericVulnerabilityReport + "," + v1.MimeTypeNativeReport})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Empty(body)
	}
}

func TestArtifactTestSuite(t *testing.T) {
	suite.Run(t, &ArtifactTestSuite{})
}
