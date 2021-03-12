package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/allowlist/models"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	htesting.Suite
	dao DAO
}

func (s *testSuite) SetupSuite() {
	s.Suite.SetupSuite()
	s.Suite.ClearSQLs = []string{
		"DELETE FROM cve_allowlist WHERE 1 = 1",
	}
	s.dao = New()
}

func (s *testSuite) TestSetAndGet() {
	s.TearDownSuite()
	l, err := s.dao.QueryByProjectID(s.Context(), 5)
	s.Nil(err)
	s.Nil(l)
	var longList []models.CVEAllowlistItem
	for i := 0; i < 50; i++ {
		longList = append(longList, models.CVEAllowlistItem{CVEID: "CVE-1999-0067"})
	}

	e := int64(1573254000)
	in1 := models.CVEAllowlist{ProjectID: 3, Items: longList, ExpiresAt: &e}
	_, err = s.dao.Set(s.Context(), in1)
	s.Nil(err)
	// assert.Equal(t, int64(1), n)
	out1, err := s.dao.QueryByProjectID(s.Context(), 3)
	s.Nil(err)
	s.Equal(int64(3), out1.ProjectID)
	s.Equal(longList, out1.Items)
	s.Equal(e, *out1.ExpiresAt)

	sysCVEs := []models.CVEAllowlistItem{
		{CVEID: "CVE-2019-10164"},
		{CVEID: "CVE-2017-12345"},
	}
	in3 := models.CVEAllowlist{Items: sysCVEs}
	_, err = s.dao.Set(s.Context(), in3)
	s.Nil(err)

}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
