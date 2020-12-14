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

package report

import (
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/suite"
)

// TestManagerSuite is a test suite for the report manager.
type TestManagerSuite struct {
	suite.Suite

	m          Manager
	reportUUID string
}

// TestManager is an entry of suite TestManagerSuite.
func TestManager(t *testing.T) {
	suite.Run(t, &TestManagerSuite{})
}

// SetupSuite prepares test env for suite TestManagerSuite.
func (suite *TestManagerSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()

	suite.m = NewManager()
}

// SetupTest prepares env for test cases.
func (suite *TestManagerSuite) SetupTest() {
	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
	}

	uuid, err := suite.m.Create(orm.Context(), rp)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(uuid)
	suite.reportUUID = uuid
}

// TearDownTest clears test env for test cases.
func (suite *TestManagerSuite) TearDownTest() {
	suite.Nil(suite.m.Delete(orm.Context(), suite.reportUUID))
}

// TestManagerGetBy tests the get by method.
func (suite *TestManagerSuite) TestManagerGetBy() {
	l, err := suite.m.GetBy(orm.Context(), "d1000", "ruuid", []string{v1.MimeTypeNativeReport})
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))
	suite.Require().Equal(suite.reportUUID, l[0].UUID)

	l, err = suite.m.GetBy(orm.Context(), "d1000", "ruuid", nil)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))
	suite.Require().Equal(suite.reportUUID, l[0].UUID)

	l, err = suite.m.GetBy(orm.Context(), "d1000", "", nil)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))
	suite.Require().Equal(suite.reportUUID, l[0].UUID)
}

// TestManagerUpdateReportData tests update job report data.
func (suite *TestManagerSuite) TestManagerUpdateReportData() {
	err := suite.m.UpdateReportData(orm.Context(), suite.reportUUID, "{\"a\":1000}")
	suite.Require().NoError(err)

	l, err := suite.m.GetBy(orm.Context(), "d1000", "ruuid", []string{v1.MimeTypeNativeReport})
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))

	suite.Equal("{\"a\":1000}", l[0].Report)
}

// TestManagerDeleteByDigests ...
func (suite *TestManagerSuite) TestManagerDeleteByDigests() {
	// Mock new data
	rp1 := &scan.Report{
		Digest:           "d2000",
		RegistrationUUID: "ruuid1",
		MimeType:         v1.MimeTypeNativeReport,
	}

	rp2 := &scan.Report{
		Digest:           "d2000",
		RegistrationUUID: "ruuid2",
		MimeType:         v1.MimeTypeNativeReport,
	}

	var reportUUIDs []string
	for _, rp := range []*scan.Report{rp1, rp2} {
		uuid, err := suite.m.Create(orm.Context(), rp)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(uuid)
		reportUUIDs = append(reportUUIDs, uuid)
	}

	l, err := suite.m.List(orm.Context(), q.New(q.KeyWords{"uuid__in": reportUUIDs}))
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(l))

	err = suite.m.DeleteByDigests(orm.Context())
	suite.Require().NoError(err)

	err = suite.m.DeleteByDigests(orm.Context(), "d2000")
	suite.Require().NoError(err)

	l, err = suite.m.List(orm.Context(), q.New(q.KeyWords{"uuid__in": reportUUIDs}))
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(l))
}
