package dao

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/sbom/model"
	htesting "github.com/goharbor/harbor/src/testing"
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
	sbomReport := &model.Report{
		UUID:             "uuid",
		ArtifactID:       111,
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeSBOMReport,
		ReportSummary:    `{"sbom_digest": "sha256:abc"}`,
	}
	suite.create(sbomReport)
}

// TearDownTest clears enf for test case.
func (suite *ReportTestSuite) TearDownTest() {
	_, err := suite.dao.DeleteMany(orm.Context(), q.Query{Keywords: q.KeyWords{"uuid": "uuid"}})
	require.NoError(suite.T(), err)
	_, err = suite.dao.DeleteMany(orm.Context(), q.Query{Keywords: q.KeyWords{"uuid": "uuid-indexed"}})
	require.NoError(suite.T(), err)
}

func (suite *ReportTestSuite) TestDeleteReportBySBOMDigest() {
	l, err := suite.dao.List(orm.Context(), nil)
	suite.Require().NoError(err)
	suite.Equal(1, len(l))
	err = suite.dao.DeleteByExtraAttr(orm.Context(), v1.MimeTypeSBOMReport, "sbom_digest", "sha256:abc")
	suite.Require().NoError(err)
	l2, err := suite.dao.List(orm.Context(), nil)
	suite.Require().NoError(err)
	suite.Equal(0, len(l2))
}

// TestDeleteBySBOMDigestUsesIndexedColumn verifies deletion works via the
// indexed sbom_digest column (not the old JSONB containment predicate).
func (suite *ReportTestSuite) TestDeleteBySBOMDigestUsesIndexedColumn() {
	// First, update the report data to populate the sbom_digest column.
	err := suite.dao.UpdateReportData(orm.Context(), "uuid", `{"sbom_digest": "sha256:indexed"}`)
	suite.Require().NoError(err)

	// Verify the sbom_digest column was populated.
	l, err := suite.dao.List(orm.Context(), nil)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))
	suite.Equal("sha256:indexed", l[0].SBOMDigest)

	// Delete by the new column.
	err = suite.dao.DeleteByExtraAttr(orm.Context(), v1.MimeTypeSBOMReport, "sbom_digest", "sha256:indexed")
	suite.Require().NoError(err)

	// Verify deletion.
	l2, err := suite.dao.List(orm.Context(), nil)
	suite.Require().NoError(err)
	suite.Equal(0, len(l2))
}

// TestDeleteBySBOMDigestNoMatch ensures delete is a no-op when digest doesn't match.
func (suite *ReportTestSuite) TestDeleteBySBOMDigestNoMatch() {
	err := suite.dao.UpdateReportData(orm.Context(), "uuid", `{"sbom_digest": "sha256:abc"}`)
	suite.Require().NoError(err)

	err = suite.dao.DeleteByExtraAttr(orm.Context(), v1.MimeTypeSBOMReport, "sbom_digest", "sha256:nonexistent")
	suite.Require().NoError(err)

	l, err := suite.dao.List(orm.Context(), nil)
	suite.Require().NoError(err)
	suite.Equal(1, len(l))
}

// TestUpdateReportDataPopulatesSBOMDigest ensures that UpdateReportData
// extracts sbom_digest from the JSON and stores it in the indexed column.
func (suite *ReportTestSuite) TestUpdateReportDataPopulatesSBOMDigest() {
	err := suite.dao.UpdateReportData(orm.Context(), "uuid", `{"sbom_digest": "sha256:newdigest", "scan_status": "Success"}`)
	suite.Require().NoError(err)

	l, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"uuid": "uuid"}))
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))
	suite.Equal("sha256:newdigest", l[0].SBOMDigest)
}

func (suite *ReportTestSuite) create(r *model.Report) {
	id, err := suite.dao.Create(orm.Context(), r)
	suite.Require().NoError(err)
	suite.Require().Condition(func() (success bool) {
		success = id > 0
		return
	})
}

// TestReportUpdateReportData tests update the report data.
func (suite *ReportTestSuite) TestReportUpdateReportData() {
	err := suite.dao.UpdateReportData(orm.Context(), "uuid", "{}")
	suite.Require().NoError(err)

	l, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"uuid": "uuid"}))
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))
	suite.Equal("{}", l[0].ReportSummary)

	err = suite.dao.UpdateReportData(orm.Context(), "uuid", "{\"a\": 900}")
	suite.Require().NoError(err)
}

func (suite *ReportTestSuite) TestUpdate() {
	err := suite.dao.Update(orm.Context(), &model.Report{
		UUID:             "uuid",
		ArtifactID:       111,
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeSBOMReport,
		ReportSummary:    `{"sbom_digest": "sha256:abc"}`,
	}, "report")
	suite.Require().NoError(err)
	query1 := &q.Query{
		PageSize:   1,
		PageNumber: 1,
		Keywords: map[string]any{
			"artifact_id":       111,
			"registration_uuid": "ruuid",
			"mime_type":         v1.MimeTypeSBOMReport,
		},
	}
	l, err := suite.dao.List(orm.Context(), query1)
	suite.Require().Equal(1, len(l))
	suite.Equal(l[0].ReportSummary, `{"sbom_digest": "sha256:abc"}`)
}

// TestReportList tests list reports with query parameters.
func (suite *ReportTestSuite) TestReportList() {
	query1 := &q.Query{
		PageSize:   1,
		PageNumber: 1,
		Keywords: map[string]any{
			"artifact_id":       111,
			"registration_uuid": "ruuid",
			"mime_type":         v1.MimeTypeSBOMReport,
		},
	}
	l, err := suite.dao.List(orm.Context(), query1)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(l))

	query2 := &q.Query{
		PageSize:   1,
		PageNumber: 1,
		Keywords: map[string]any{
			"artifact_id": 222,
		},
	}
	l, err = suite.dao.List(orm.Context(), query2)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(l))
}
