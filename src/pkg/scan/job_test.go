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
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	htesting "github.com/goharbor/harbor/src/testing"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	mocktesting "github.com/goharbor/harbor/src/testing/mock"
	v1testing "github.com/goharbor/harbor/src/testing/pkg/scan/rest/v1"
)

// JobTestSuite is a test suite to test the scan job.
type JobTestSuite struct {
	htesting.Suite

	defaultClientPool v1.ClientPool
	mcp               *v1testing.ClientPool
	reportIDs         []string
}

// TestJob is the entry of JobTestSuite.
func TestJob(t *testing.T) {
	suite.Run(t, &JobTestSuite{})
}

// SetupSuite sets up test env for JobTestSuite.
func (suite *JobTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	mcp := &v1testing.ClientPool{}
	suite.defaultClientPool = v1.DefaultClientPool
	v1.DefaultClientPool = mcp

	suite.mcp = mcp
}

// TearDownSuite clears test env for TearDownSuite.
func (suite *JobTestSuite) TearDownSuite() {
	v1.DefaultClientPool = suite.defaultClientPool
	for _, id := range suite.reportIDs {
		_ = report.Mgr.Delete(suite.Context(), id)
	}
}

// TestJob tests the scan job
func (suite *JobTestSuite) TestJob() {
	ctx := &mockjobservice.MockJobContext{}
	lg := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(lg)
	ctx.On("OPCommand").Return(job.NilCommand, false)

	r := &scanner.Registration{
		ID:   0,
		UUID: "uuid",
		Name: "TestJob",
		URL:  "https://trivy.com:8080",
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

	robot := &robot.Robot{
		Robot: model.Robot{
			ID:     1,
			Name:   "robot",
			Secret: "token",
		},
		Level: "project",
	}

	robotData, err := robot.ToJSON()
	require.NoError(suite.T(), err)

	mimeTypes := []string{v1.MimeTypeNativeReport, v1.MimeTypeGenericVulnerabilityReport}

	jp := make(job.Parameters)
	jp[JobParamRegistration] = rData
	jp[JobParameterRequest] = sData
	jp[JobParameterMimes] = mimeTypes
	jp[JobParameterAuthType] = "Basic"
	jp[JobParameterRobot] = robotData

	mc := &v1testing.Client{}
	sre := &v1.ScanResponse{
		ID: "scan_id",
	}
	mc.On("SubmitScan", sr).Return(sre, nil)

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

	jRep, err := json.Marshal(rp)
	require.NoError(suite.T(), err)

	mc.On("GetScanReport", "scan_id", v1.MimeTypeNativeReport, v1.MimeTypeGenericVulnerabilityReport).Return(string(jRep), nil)
	mocktesting.OnAnything(suite.mcp, "Get").Return(mc, nil)

	j := &Job{}
	err = j.Run(ctx, jp)
	require.NoError(suite.T(), err)
}

func (suite *JobTestSuite) TestfetchScanReportFromScanner() {
	vulnRpt := &vuln.Report{
		GeneratedAt: time.Now().UTC().String(),
		Scanner: &v1.Scanner{
			Name:    "Trivy",
			Vendor:  "Harbor",
			Version: "0.1.0",
		},
		Severity: vuln.High,
	}
	rptContent, err := json.Marshal(vulnRpt)
	require.NoError(suite.T(), err)
	rawContent := string(rptContent)
	ctx := suite.Context()
	dgst := "sha256:mydigest"
	uuid := `7f20b1b9-6117-4a2e-820b-e4cc0401f15a`
	scannerUUID := `7f20b1b9-6117-4a2e-820b-e4cc0401f15b`
	rpt := &scan.Report{
		UUID:             uuid,
		RegistrationUUID: scannerUUID,
		Digest:           dgst,
		MimeType:         v1.MimeTypeDockerArtifact,
		Report:           rawContent,
	}

	ctx = suite.Context()
	rptID, err := report.Mgr.Create(ctx, rpt)
	suite.reportIDs = append(suite.reportIDs, rptID)
	require.NoError(suite.T(), err)
	client := &v1testing.Client{}
	client.On("GetScanReport", mock.Anything, v1.MimeTypeGenericVulnerabilityReport, mock.Anything).Return(rawContent, nil)
	parameters := "sbom_media_type=application/spdx+json"
	rawRept, err := fetchScanReportFromScanner(client, "abc", v1.MimeTypeGenericVulnerabilityReport, parameters)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), rawContent, rawRept)
}
