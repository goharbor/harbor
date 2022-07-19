package systemartifact

import (
	"context"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/systemartifact/dao"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type defaultCleanupCriteriaTestSuite struct {
	htesting.Suite
	dao             dao.DAO
	ctx             context.Context
	cleanupCriteria Selector
}

func (suite *defaultCleanupCriteriaTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = dao.NewSystemArtifactDao()
	suite.cleanupCriteria = DefaultSelector
	common_dao.PrepareTestForPostgresSQL()
	suite.ctx = orm.Context()
	sa := model.SystemArtifact{}
	suite.ClearTables = append(suite.ClearTables, sa.TableName())
}

func (suite *defaultCleanupCriteriaTestSuite) TestList() {
	// insert a normal system artifact
	currentTime := time.Now()

	{
		saNow := model.SystemArtifact{
			Repository: "test_repo1000",
			Digest:     "test_digest1000",
			Size:       int64(100),
			Vendor:     "test_vendor1000",
			Type:       "test_repo_type",
			CreateTime: currentTime,
			ExtraAttrs: "",
		}

		oneDayAndElevenMinutesAgo := time.Duration(96500) * time.Second

		sa1 := model.SystemArtifact{
			Repository: "test_repo2000",
			Digest:     "test_digest2000",
			Size:       int64(100),
			Vendor:     "test_vendor2000",
			Type:       "test_repo_type",
			CreateTime: currentTime.Add(-oneDayAndElevenMinutesAgo),
			ExtraAttrs: "",
		}

		twoDaysAgo := time.Duration(172800) * time.Second
		sa2 := model.SystemArtifact{
			Repository: "test_repo3000",
			Digest:     "test_digest3000",
			Size:       int64(100),
			Vendor:     "test_vendor3000",
			Type:       "test_repo_type",
			CreateTime: currentTime.Add(-twoDaysAgo),
			ExtraAttrs: "",
		}

		id1, err := suite.dao.Create(suite.ctx, &saNow)
		id2, err := suite.dao.Create(suite.ctx, &sa1)
		id3, err := suite.dao.Create(suite.ctx, &sa2)

		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, id1, "Expected a valid record identifier but was 0")
		suite.NotEqual(0, id2, "Expected a valid record identifier but was 0")
		suite.NotEqual(0, id3, "Expected a valid record identifier but was 0")

		actualSysArtifactIds := make(map[int64]bool)

		sysArtifactList, err := suite.cleanupCriteria.List(suite.ctx)

		for _, sysArtifact := range sysArtifactList {
			actualSysArtifactIds[sysArtifact.ID] = true
		}
		expectedSysArtifactIds := map[int64]bool{id2: true, id3: true}

		for k := range expectedSysArtifactIds {
			_, ok := actualSysArtifactIds[k]
			suite.Truef(ok, "Expected system artifact : %v not present in the list", k)
		}
	}
}

func (suite *defaultCleanupCriteriaTestSuite) TestListWithFilters() {
	// insert a normal system artifact
	currentTime := time.Now()

	{
		saNow := model.SystemArtifact{
			Repository: "test_repo73000",
			Digest:     "test_digest73000",
			Size:       int64(100),
			Vendor:     "test_vendor73000",
			Type:       "test_repo_type",
			CreateTime: currentTime,
			ExtraAttrs: "",
		}

		oneDayAndElevenMinutesAgo := time.Duration(96500) * time.Second

		sa1 := model.SystemArtifact{
			Repository: "test_repo29000",
			Digest:     "test_digest29000",
			Size:       int64(100),
			Vendor:     "test_vendor29000",
			Type:       "test_repo_type",
			CreateTime: currentTime.Add(-oneDayAndElevenMinutesAgo),
			ExtraAttrs: "",
		}

		twoDaysAgo := time.Duration(172800) * time.Second
		sa2 := model.SystemArtifact{
			Repository: "test_repo37000",
			Digest:     "test_digest37000",
			Size:       int64(100),
			Vendor:     "test_vendor37000",
			Type:       "test_repo_type",
			CreateTime: currentTime.Add(-twoDaysAgo),
			ExtraAttrs: "",
		}

		id1, err := suite.dao.Create(suite.ctx, &saNow)
		id2, err := suite.dao.Create(suite.ctx, &sa1)
		id3, err := suite.dao.Create(suite.ctx, &sa2)

		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, id1, "Expected a valid record identifier but was 0")
		suite.NotEqual(0, id2, "Expected a valid record identifier but was 0")
		suite.NotEqual(0, id3, "Expected a valid record identifier but was 0")

		actualSysArtifactIds := make(map[int64]bool)

		query := q.Query{Keywords: map[string]interface{}{"vendor": "test_vendor37000", "repository": "test_repo37000"}}
		sysArtifactList, err := suite.cleanupCriteria.ListWithFilters(suite.ctx, &query)

		for _, sysArtifact := range sysArtifactList {
			actualSysArtifactIds[sysArtifact.ID] = true
		}
		expectedSysArtifactIds := map[int64]bool{id3: true}

		for k := range expectedSysArtifactIds {
			_, ok := actualSysArtifactIds[k]
			suite.Truef(ok, "Expected system artifact : %v not present in the list", k)
		}
	}
}

func TestCleanupCriteriaTestSuite(t *testing.T) {
	suite.Run(t, &defaultCleanupCriteriaTestSuite{})
}
