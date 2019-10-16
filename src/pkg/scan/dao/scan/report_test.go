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
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/q"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ReportTestSuite is test suite of testing report DAO.
type ReportTestSuite struct {
	suite.Suite
}

// TestReport is the entry of ReportTestSuite.
func TestReport(t *testing.T) {
	suite.Run(t, &ReportTestSuite{})
}

// SetupSuite prepares env for test suite.
func (suite *ReportTestSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()
}

// SetupTest prepares env for test case.
func (suite *ReportTestSuite) SetupTest() {
	r := &Report{
		UUID:             "uuid",
		TrackID:          "track-uuid",
		Digest:           "digest1001",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
		Status:           job.PendingStatus.String(),
		StatusCode:       job.PendingStatus.Code(),
	}

	id, err := CreateReport(r)
	require.NoError(suite.T(), err)
	require.Condition(suite.T(), func() (success bool) {
		success = id > 0
		return
	})
}

// TearDownTest clears enf for test case.
func (suite *ReportTestSuite) TearDownTest() {
	err := DeleteReport("uuid")
	require.NoError(suite.T(), err)
}

// TestReportList tests list reports with query parameters.
func (suite *ReportTestSuite) TestReportList() {
	query1 := &q.Query{
		PageSize:   1,
		PageNumber: 1,
		Keywords: map[string]interface{}{
			"digest":            "digest1001",
			"registration_uuid": "ruuid",
			"mime_type":         v1.MimeTypeNativeReport,
		},
	}
	l, err := ListReports(query1)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	query2 := &q.Query{
		PageSize:   1,
		PageNumber: 1,
		Keywords: map[string]interface{}{
			"digest": "digest1002",
		},
	}
	l, err = ListReports(query2)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, len(l))
}

// TestReportUpdateJobID tests update job ID of the report.
func (suite *ReportTestSuite) TestReportUpdateJobID() {
	err := UpdateJobID("track-uuid", "jobid001")
	require.NoError(suite.T(), err)

	l, err := ListReports(nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))
	assert.Equal(suite.T(), "jobid001", l[0].JobID)
}

// TestReportUpdateReportData tests update the report data.
func (suite *ReportTestSuite) TestReportUpdateReportData() {
	err := UpdateReportData("uuid", "{}", 1000)
	require.NoError(suite.T(), err)

	l, err := ListReports(nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))
	assert.Equal(suite.T(), "{}", l[0].Report)

	err = UpdateReportData("uuid", "{\"a\": 900}", 900)
	require.Error(suite.T(), err)
}

// TestReportUpdateStatus tests update the report status.
func (suite *ReportTestSuite) TestReportUpdateStatus() {
	err := UpdateReportStatus("track-uuid", job.RunningStatus.String(), job.RunningStatus.Code(), 1000)
	require.NoError(suite.T(), err)

	err = UpdateReportStatus("track-uuid", job.RunningStatus.String(), job.RunningStatus.Code(), 900)
	require.Error(suite.T(), err)

	err = UpdateReportStatus("track-uuid", job.PendingStatus.String(), job.PendingStatus.Code(), 1000)
	require.Error(suite.T(), err)
}
