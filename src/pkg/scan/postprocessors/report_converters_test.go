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

package postprocessors

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const sampleReport = `{
	"generated_at": "2020-08-01T18:28:49.072885592Z",
	"artifact": {
	  "repository": "library/ubuntu",
	  "digest": "sha256:d5b40885539615b9aeb7119516427959a158386af13e00d79a7da43ad1b3fb87",
	  "mime_type": "application/vnd.docker.distribution.manifest.v2+json"
	},
	"scanner": {
	  "name": "Trivy",
	  "vendor": "Aqua Security",
	  "version": "v0.9.1"
	},
	"severity": "Medium",
	"vulnerabilities": [
	  {
		"id": "CVE-2019-18276",
		"package": "bash",
		"version": "5.0-6ubuntu1.1",
		"severity": "Low",
		"description": "An issue was discovered in disable_priv_mode in shell.c in GNU Bash through 5.0 patch 11. By default, if Bash is run with its effective UID not equal to its real UID, it will drop privileges by setting its effective UID to its real UID. However, it does so incorrectly. On Linux and other systems that support \"saved UID\" functionality, the saved UID is not dropped. An attacker with command execution in the shell can use \"enable -f\" for runtime loading of a new builtin, which can be a shared object that calls setuid() and therefore regains privileges. However, binaries running with an effective UID of 0 are unaffected.",
		"links": [
		  "http://packetstormsecurity.com/files/155498/Bash-5.0-Patch-11-Privilege-Escalation.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-18276",
		  "https://github.com/bminor/bash/commit/951bdaad7a18cc0dc1036bba86b18b90874d39ff",
		  "https://security.netapp.com/advisory/ntap-20200430-0003/",
		  "https://www.youtube.com/watch?v=-wGtxJ8opa8"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		}
	  }
	]
}`

const sampleReportWithCWEAndCVSS = `{
	"generated_at": "2020-08-01T18:28:49.072885592Z",
	"artifact": {
	  "repository": "library/ubuntu",
	  "digest": "sha256:d5b40885539615b9aeb7119516427959a158386af13e00d79a7da43ad1b3fb87",
	  "mime_type": "application/vnd.docker.distribution.manifest.v2+json"
	},
	"scanner": {
	  "name": "Trivy",
	  "vendor": "Aqua Security",
	  "version": "v0.9.1"
	},
	"severity": "Medium",
	"vulnerabilities": [
	  {
		"id": "CVE-2019-18276",
		"package": "bash",
		"version": "5.0-6ubuntu1.1",
		"severity": "Low",
		"description": "An issue was discovered in disable_priv_mode in shell.c in GNU Bash through 5.0 patch 11. By default, if Bash is run with its effective UID not equal to its real UID, it will drop privileges by setting its effective UID to its real UID. However, it does so incorrectly. On Linux and other systems that support \"saved UID\" functionality, the saved UID is not dropped. An attacker with command execution in the shell can use \"enable -f\" for runtime loading of a new builtin, which can be a shared object that calls setuid() and therefore regains privileges. However, binaries running with an effective UID of 0 are unaffected.",
		"links": [
		  "http://packetstormsecurity.com/files/155498/Bash-5.0-Patch-11-Privilege-Escalation.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-18276",
		  "https://github.com/bminor/bash/commit/951bdaad7a18cc0dc1036bba86b18b90874d39ff",
		  "https://security.netapp.com/advisory/ntap-20200430-0003/",
		  "https://www.youtube.com/watch?v=-wGtxJ8opa8"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		},
		"cwe_ids": ["CWE-476", "CWE-345"],
		"preferred_cvss":{
			"score_v3": 3.2,
			"score_v2": 2.3,
			"vector_v3": "CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N",
			"vector_v2": "AV:L/AC:M/Au:N/C:P/I:N/A:N"
		}
	  }
	]
}`

const sampleReportWithCompleteVulnData = `{
	"generated_at": "2020-08-01T18:28:49.072885592Z",
	"artifact": {
	  "repository": "library/ubuntu",
	  "digest": "sha256:d5b40885539615b9aeb7119516427959a158386af13e00d79a7da43ad1b3fb87",
	  "mime_type": "application/vnd.docker.distribution.manifest.v2+json"
	},
	"scanner": {
	  "name": "Trivy",
	  "vendor": "Aqua Security",
	  "version": "v0.9.1"
	},
	"severity": "Medium",
	"vulnerabilities": [
	  {
		"id": "CVE-2019-18276",
		"package": "bash",
		"version": "5.0-6ubuntu1.1",
		"severity": "Low",
		"description": "An issue was discovered in disable_priv_mode in shell.c in GNU Bash through 5.0 patch 11. By default, if Bash is run with its effective UID not equal to its real UID, it will drop privileges by setting its effective UID to its real UID. However, it does so incorrectly. On Linux and other systems that support \"saved UID\" functionality, the saved UID is not dropped. An attacker with command execution in the shell can use \"enable -f\" for runtime loading of a new builtin, which can be a shared object that calls setuid() and therefore regains privileges. However, binaries running with an effective UID of 0 are unaffected.",
		"links": [
		  "http://packetstormsecurity.com/files/155498/Bash-5.0-Patch-11-Privilege-Escalation.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-18276",
		  "https://github.com/bminor/bash/commit/951bdaad7a18cc0dc1036bba86b18b90874d39ff",
		  "https://security.netapp.com/advisory/ntap-20200430-0003/",
		  "https://www.youtube.com/watch?v=-wGtxJ8opa8"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		},
		"cwe_ids": ["CWE-476", "CWE-345"],
		"preferred_cvss":{
			"score_v3": 3.2,
			"score_v2": 2.3,
			"vector_v3": "CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N",
			"vector_v2": "AV:L/AC:M/Au:N/C:P/I:N/A:N"
		},
		"vendor_attributes":{
			"CVSS":{
				"nvd" : {
					"V2Score": 7.1,
					"V2Vector": "AV:L/AC:M/Au:N/C:P/I:N/A:N",
					"V3Score": 6.5,
					"V3Vector":"CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N"
				}
			}
		}
	  }
	]
}`

const sampleReportWithMixedSeverity = `{
	"generated_at": "2020-08-01T18:28:49.072885592Z",
	"artifact": {
	  "repository": "library/ubuntu",
	  "digest": "sha256:d5b40885539615b9aeb7119516427959a158386af13e00d79a7da43ad1b3fb87",
	  "mime_type": "application/vnd.docker.distribution.manifest.v2+json"
	},
	"scanner": {
	  "name": "Trivy",
	  "vendor": "Aqua Security",
	  "version": "v0.9.1"
	},
	"severity": "Medium",
	"vulnerabilities": [
	  {
		"id": "CVE-2019-18276",
		"package": "bash",
		"version": "5.0-6ubuntu1.1",
		"severity": "Low",
		"description": "An issue was discovered in disable_priv_mode in shell.c in GNU Bash through 5.0 patch 11. By default, if Bash is run with its effective UID not equal to its real UID, it will drop privileges by setting its effective UID to its real UID. However, it does so incorrectly. On Linux and other systems that support \"saved UID\" functionality, the saved UID is not dropped. An attacker with command execution in the shell can use \"enable -f\" for runtime loading of a new builtin, which can be a shared object that calls setuid() and therefore regains privileges. However, binaries running with an effective UID of 0 are unaffected.",
		"links": [
		  "http://packetstormsecurity.com/files/155498/Bash-5.0-Patch-11-Privilege-Escalation.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-18276",
		  "https://github.com/bminor/bash/commit/951bdaad7a18cc0dc1036bba86b18b90874d39ff",
		  "https://security.netapp.com/advisory/ntap-20200430-0003/",
		  "https://www.youtube.com/watch?v=-wGtxJ8opa8"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		},
		"cwe_ids": ["CWE-476", "CWE-345"],
		"preferred_cvss":{
			"score_v3": 3.2,
			"score_v2": 2.3,
			"vector_v3": "CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N",
			"vector_v2": "AV:L/AC:M/Au:N/C:P/I:N/A:N"
		},
		"vendor_attributes":{
			"CVSS":{
				"nvd" : {
					"V2Score": 7.1,
					"V2Vector": "AV:L/AC:M/Au:N/C:P/I:N/A:N",
					"V3Score": 6.5,
					"V3Vector":"CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N"
				}
			}
		}
	  },
	  	{
		"id": "CVE-2013-7445",
		"package": "linux",
		"version": "4.9.189-3+deb9u2",
		"severity": "High",
		"description": "An issue was discovered in disable_priv_mode in shell.c in GNU Bash through 5.0 patch 11. By default, if Bash is run with its effective UID not equal to its real UID, it will drop privileges by setting its effective UID to its real UID. However, it does so incorrectly. On Linux and other systems that support \"saved UID\" functionality, the saved UID is not dropped. An attacker with command execution in the shell can use \"enable -f\" for runtime loading of a new builtin, which can be a shared object that calls setuid() and therefore regains privileges. However, binaries running with an effective UID of 0 are unaffected.",
		"links": [
		  "http://packetstormsecurity.com/files/155498/Bash-5.0-Patch-11-Privilege-Escalation.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-18276",
		  "https://github.com/bminor/bash/commit/951bdaad7a18cc0dc1036bba86b18b90874d39ff",
		  "https://security.netapp.com/advisory/ntap-20200430-0003/",
		  "https://www.youtube.com/watch?v=-wGtxJ8opa8"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		},
		"cwe_ids": ["CWE-476", "CWE-345"],
		"preferred_cvss":{
			"score_v3": 3.2,
			"score_v2": 2.3,
			"vector_v3": "CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N",
			"vector_v2": "AV:L/AC:M/Au:N/C:P/I:N/A:N"
		},
		"vendor_attributes":{
			"CVSS":{
				"nvd" : {
					"V2Score": 7.1,
					"V2Vector": "AV:L/AC:M/Au:N/C:P/I:N/A:N",
					"V3Score": 6.5,
					"V3Vector":"CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N"
				}
			}
		}
	  },
      {
		"id": "CVE-2019-2182",
		"package": "bash",
		"version": "5.0-6ubuntu1.1",
		"severity": "Medium",
		"description": "An issue was discovered in disable_priv_mode in shell.c in GNU Bash through 5.0 patch 11. By default, if Bash is run with its effective UID not equal to its real UID, it will drop privileges by setting its effective UID to its real UID. However, it does so incorrectly. On Linux and other systems that support \"saved UID\" functionality, the saved UID is not dropped. An attacker with command execution in the shell can use \"enable -f\" for runtime loading of a new builtin, which can be a shared object that calls setuid() and therefore regains privileges. However, binaries running with an effective UID of 0 are unaffected.",
		"links": [
		  "http://packetstormsecurity.com/files/155498/Bash-5.0-Patch-11-Privilege-Escalation.html",
		  "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2019-18276",
		  "https://github.com/bminor/bash/commit/951bdaad7a18cc0dc1036bba86b18b90874d39ff",
		  "https://security.netapp.com/advisory/ntap-20200430-0003/",
		  "https://www.youtube.com/watch?v=-wGtxJ8opa8"
		],
		"layer": {
		  "digest": "sha256:4739cd2f4f486596c583c79f6033f1a9dee019389d512603609494678c8ccd53",
		  "diff_id": "sha256:f66829086c450acd5f67d0529a58f7120926c890f04e17aa7f0e9365da86480a"
		},
		"cwe_ids": ["CWE-476", "CWE-345"],
		"preferred_cvss":{
			"score_v3": 3.2,
			"score_v2": 2.3,
			"vector_v3": "CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N",
			"vector_v2": "AV:L/AC:M/Au:N/C:P/I:N/A:N"
		},
		"vendor_attributes":{
			"CVSS":{
				"nvd" : {
					"V2Score": 7.1,
					"V2Vector": "AV:L/AC:M/Au:N/C:P/I:N/A:N",
					"V3Score": 6.5,
					"V3Vector":"CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N"
				}
			}
		}
	  }

	]
}`

// TestReportConverterSuite
type TestReportConverterSuite struct {
	htesting.Suite
	rc                     NativeScanReportConverter
	rpUUID                 string
	vulnerabilityRecordDao scan.VulnerabilityRecordDao
	reportDao              scan.DAO
	registrationID         string
}

// SetupTest prepares env for test cases.
func (suite *TestReportConverterSuite) SetupTest() {
	suite.rpUUID = "reportUUID"
	suite.registrationID = "ruuid"
	r := &scanner.Registration{
		UUID:        suite.registrationID,
		Name:        "forUT",
		Description: "sample registration",
		URL:         "https://sample.scanner.com",
	}

	_, err := scanner.AddRegistration(suite.Context(), r)
	require.NoError(suite.T(), err, "add new registration")
}

// TestReportConverterTests specifies the test suite
func TestReportConverterTests(t *testing.T) {
	suite.Run(t, &TestReportConverterSuite{})
}

// SetupSuite sets up the report converter suite test cases
func (suite *TestReportConverterSuite) SetupSuite() {
	suite.rc = NewNativeToRelationalSchemaConverter()
	suite.Suite.SetupSuite()
	suite.vulnerabilityRecordDao = scan.NewVulnerabilityRecordDao()
	suite.reportDao = scan.New()
}

// TearDownTest clears test env for test cases.
func (suite *TestReportConverterSuite) TearDownTest() {
	// No delete method defined in manager as no requirement,
	// so, to clear env, call dao method here
	scanner.DeleteRegistration(suite.Context(), suite.registrationID)
	reports, err := suite.reportDao.List(orm.Context(), &q.Query{})
	require.True(suite.T(), err == nil, "Failed to delete vulnerability records")
	for _, report := range reports {
		_, err := suite.reportDao.DeleteMany(orm.Context(), q.Query{Keywords: q.KeyWords{"uuid": report.UUID}})
		require.NoError(suite.T(), err)
		_, err = suite.vulnerabilityRecordDao.DeleteForReport(orm.Context(), report.UUID)
		require.NoError(suite.T(), err, "Failed to delete vulnerability records")
		_, err = suite.vulnerabilityRecordDao.DeleteForScanner(orm.Context(), report.RegistrationUUID)
		require.NoError(suite.T(), err, "Failed to delete vulnerability records")
	}

}

// TestConvertReport tests the report conversion logic
func (suite *TestReportConverterSuite) TestConvertReport() {
	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
		Report:           sampleReport,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1000),
		UUID:             "reportUUID",
	}
	suite.create(rp)
	ruuid, summary, err := suite.rc.ToRelationalSchema(orm.Context(), rp.UUID, rp.RegistrationUUID, rp.Digest, rp.Report)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), rp.UUID, ruuid)
	suite.validateReportSummary(summary, sampleReport)
}

// TestConvertReportWithCWEAndCVSS tests the report conversion with CVSS and CWE information
func (suite *TestReportConverterSuite) TestConvertReportWithCWEAndCVSS() {
	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
		Report:           sampleReportWithCWEAndCVSS,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1000),
		UUID:             "reportUUID1",
	}
	suite.create(rp)
	ruuid, summary, err := suite.rc.ToRelationalSchema(orm.Context(), rp.UUID, rp.RegistrationUUID, rp.Digest, rp.Report)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), rp.UUID, ruuid)
	suite.validateReportSummary(summary, sampleReportWithCWEAndCVSS)
}

// TestConvertReportWithCompleteVulnData tests report conversion with complete vulnerability data
func (suite *TestReportConverterSuite) TestConvertReportWithCompleteVulnData() {
	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
		Report:           sampleReportWithCompleteVulnData,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1000),
		UUID:             "reportUUID2",
	}
	suite.create(rp)
	ruuid, summary, err := suite.rc.ToRelationalSchema(orm.Context(), rp.UUID, rp.RegistrationUUID, rp.Digest, rp.Report)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), rp.UUID, ruuid)
	suite.validateReportSummary(summary, sampleReportWithCompleteVulnData)
}

func (suite *TestReportConverterSuite) TestConvertToNativeReport() {

	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
		Report:           sampleReportWithCompleteVulnData,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1000),
		UUID:             "reportUUID2",
	}
	suite.create(rp)
	_, summary, err := suite.rc.ToRelationalSchema(orm.Context(), rp.UUID, rp.RegistrationUUID, rp.Digest, sampleReportWithCompleteVulnData)
	completeReport, err := suite.rc.FromRelationalSchema(orm.Context(), rp.UUID, rp.Digest, summary)
	require.NoError(suite.T(), err)
	v := new(vuln.Report)
	err = json.Unmarshal([]byte(sampleReportWithCompleteVulnData), v)
	require.NoError(suite.T(), err)
	v.WithArtifactDigest(rp.Digest)
	data, err := json.Marshal(v)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), string(data), completeReport)
}

func (suite *TestReportConverterSuite) TestNativeReportSummaryAfterConversion() {

	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeGenericVulnerabilityReport,
		Report:           sampleReportWithMixedSeverity,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1000),
		UUID:             "reportUUID2",
	}
	suite.create(rp)
	_, summary, err := suite.rc.ToRelationalSchema(orm.Context(), rp.UUID, rp.RegistrationUUID, rp.Digest, rp.Report)
	require.NoError(suite.T(), err)
	completeReport, err := suite.rc.FromRelationalSchema(orm.Context(), rp.UUID, rp.Digest, summary)
	require.NoError(suite.T(), err)
	v := new(vuln.Report)
	err = json.Unmarshal([]byte(rp.Report), v)
	require.NoError(suite.T(), err)
	v.WithArtifactDigest(rp.Digest)
	data, err := json.Marshal(v)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), string(data), completeReport)
	// validate that summarization happens correctly on this report
	rp.Report = completeReport
	rp.Status = job.SuccessStatus.String()
	summ, err := report.GenerateSummary(rp)
	require.NoError(suite.T(), err)
	nativeReportSummary := summ.(*vuln.NativeReportSummary)
	sevMapping := nativeReportSummary.Summary.Summary
	assert.True(suite.T(), len(sevMapping) == 3, "Expected entries in severity mapping for 'High', 'Low', 'Medium'")
	assert.Equal(suite.T(), 1, sevMapping[vuln.Low])
	assert.Equal(suite.T(), 1, sevMapping[vuln.High])
	assert.Equal(suite.T(), 1, sevMapping[vuln.Medium])
}

func (suite *TestReportConverterSuite) TestGenericVulnReportSummaryAfterConversion() {

	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
		Report:           sampleReportWithMixedSeverity,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(1000),
		UUID:             "reportUUID2",
	}
	suite.create(rp)
	_, summary, err := suite.rc.ToRelationalSchema(orm.Context(), rp.UUID, rp.RegistrationUUID, rp.Digest, rp.Report)
	require.NoError(suite.T(), err)
	completeReport, err := suite.rc.FromRelationalSchema(orm.Context(), rp.UUID, rp.Digest, summary)
	require.NoError(suite.T(), err)
	v := new(vuln.Report)
	err = json.Unmarshal([]byte(rp.Report), v)
	require.NoError(suite.T(), err)
	v.WithArtifactDigest(rp.Digest)
	data, err := json.Marshal(v)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), string(data), completeReport)
	// validate that summarization happens correctly on this report
	rp.Report = completeReport
	rp.Status = job.SuccessStatus.String()
	summ, err := report.GenerateSummary(rp)
	require.NoError(suite.T(), err)
	nativeReportSummary := summ.(*vuln.NativeReportSummary)
	sevMapping := nativeReportSummary.Summary.Summary
	assert.True(suite.T(), len(sevMapping) == 3, "Expected entries in severity mapping for 'High', 'Low', 'Medium'")
	assert.Equal(suite.T(), 1, sevMapping[vuln.Low])
	assert.Equal(suite.T(), 1, sevMapping[vuln.High])
	assert.Equal(suite.T(), 1, sevMapping[vuln.Medium])
}

func (suite *TestReportConverterSuite) create(r *scan.Report) {
	id, err := suite.reportDao.Create(orm.Context(), r)
	require.NoError(suite.T(), err)
	require.Condition(suite.T(), func() (success bool) {
		success = id > 0
		return
	})
}

func (suite *TestReportConverterSuite) validateReportSummary(summary string, rawReport string) {
	expectedReport := new(vuln.Report)
	err := json.Unmarshal([]byte(rawReport), expectedReport)
	require.NoError(suite.T(), err)
	expectedReport.Vulnerabilities = nil
	data, err := json.Marshal(expectedReport)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), string(data), summary)
}
