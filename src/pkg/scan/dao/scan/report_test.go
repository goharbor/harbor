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
	htesting "github.com/goharbor/harbor/src/testing"
	"testing"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ReportTestSuite is test suite of testing report DAO.
type ReportTestSuite struct {
	htesting.Suite

	dao DAO
}

// TestReport is the entry of ReportTestSuite.
func TestReport(t *testing.T) {
	suite.Run(t, &ReportTestSuite{})
}

// SetupSuite prepares env for test suite.
func (suite *ReportTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()

	suite.dao = New()
}

// SetupTest prepares env for test case.
func (suite *ReportTestSuite) SetupTest() {
	r := &Report{
		UUID:             "uuid",
		Digest:           "digest1001",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
	}

	suite.create(r)
}

// TearDownTest clears enf for test case.
func (suite *ReportTestSuite) TearDownTest() {
	_, err := suite.dao.DeleteMany(orm.Context(), q.Query{Keywords: q.KeyWords{"uuid": "uuid"}})
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
	l, err := suite.dao.List(orm.Context(), query1)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))

	query2 := &q.Query{
		PageSize:   1,
		PageNumber: 1,
		Keywords: map[string]interface{}{
			"digest": "digest1002",
		},
	}
	l, err = suite.dao.List(orm.Context(), query2)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(l))
}

// TestReportUpdateReportData tests update the report data.
func (suite *ReportTestSuite) TestReportUpdateReportData() {
	err := suite.dao.UpdateReportData(orm.Context(), "uuid", "{}")
	suite.Require().NoError(err)

	l, err := suite.dao.List(orm.Context(), nil)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))
	suite.Equal("{}", l[0].Report)

	err = suite.dao.UpdateReportData(orm.Context(), "uuid", "{\"a\": 900}")
	suite.Require().NoError(err)
}

func (suite *ReportTestSuite) create(r *Report) {
	id, err := suite.dao.Create(orm.Context(), r)
	suite.Require().NoError(err)
	suite.Require().Condition(func() (success bool) {
		success = id > 0
		return
	})
}
