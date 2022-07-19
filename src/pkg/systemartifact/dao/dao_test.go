package dao

import (
	"context"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type daoTestSuite struct {
	htesting.Suite
	dao DAO
	ctx context.Context
	id  int64
}

func (suite *daoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = &systemArtifactDAO{}
	common_dao.PrepareTestForPostgresSQL()
	suite.ctx = orm.Context()
	sa := model.SystemArtifact{}
	suite.ClearTables = append(suite.ClearTables, sa.TableName())
}

func (suite *daoTestSuite) SetupTest() {

}

func (suite *daoTestSuite) TeardownTest() {
	suite.ExecSQL("delete from system_artifact")
	suite.TearDownSuite()
}

func (suite *daoTestSuite) TestCreate() {
	suite.ExecSQL("delete from system_artifact")

	// insert a normal system artifact
	{
		sa := model.SystemArtifact{
			Repository: "test_repo",
			Digest:     "test_digest",
			Size:       int64(100),
			Vendor:     "test_vendor",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}
		id, err := suite.dao.Create(suite.ctx, &sa)
		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, id, "Expected a valid record identifier but was 0")
	}

	// attempt to create another system artifact with same data and then create a unique constraint violation error
	{
		sa := model.SystemArtifact{
			Repository: "test_repo",
			Digest:     "test_digest",
			Size:       int64(100),
			Vendor:     "test_vendor",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}
		id, err := suite.dao.Create(suite.ctx, &sa)

		suite.Equal(int64(0), id, "Expected id to be 0 owing to unique constraint violation")
		suite.Error(err, "Expected error to be not nil")
		errWithInfo := err.(*errors.Error)
		suite.Equalf(errors.ConflictCode, errWithInfo.Code, "Expected conflict code but was %s", errWithInfo.Code)
	}
}

func (suite *daoTestSuite) TestGet() {

	// insert a normal system artifact and attempt to get it
	{
		sa := model.SystemArtifact{
			Repository: "test_repo1",
			Digest:     "test_digest1",
			Size:       int64(100),
			Vendor:     "test_vendor",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}
		id, err := suite.dao.Create(suite.ctx, &sa)
		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, id, "Expected a valid record identifier but was 0")

		saRead, err := suite.dao.Get(suite.ctx, "test_vendor", "test_repo1", "test_digest1")
		suite.NoErrorf(err, "Unexpected error when reading system artifact: %v", err)
		suite.Equalf(id, saRead.ID, "The ID for inserted system record %d is not equal to the read system record %d", id, saRead.ID)
	}

	// attempt to retrieve a non-existent system artifact record with incorrect repo name and correct digest
	{
		saRead, err := suite.dao.Get(suite.ctx, "test_vendor", "test_repo2", "test_digest1")
		suite.Errorf(err, "Expected no record found error for provided repository and digest")
		suite.Nil(saRead, "Expected system artifact record to be nil")

		errWithInfo := err.(*errors.Error)
		suite.Equalf(errors.NotFoundCode, errWithInfo.Code, "Expected not found code but was %s", errWithInfo.Code)
	}

	// attempt to retrieve a non-existent system artifact record with correct repo name and incorrect digest
	{
		saRead, err := suite.dao.Get(suite.ctx, "test_vendor", "test_repo1", "test_digest2")
		suite.Errorf(err, "Expected no record found error for provided repository and digest")
		suite.Nil(saRead, "Expected system artifact record to be nil")

		errWithInfo := err.(*errors.Error)
		suite.Equalf(errors.NotFoundCode, errWithInfo.Code, "Expected not found code but was %s", errWithInfo.Code)
	}

	// multiple system artifact records from different vendors.
	// insert a normal system artifact and attempt to get it
	{
		sa_vendor1 := model.SystemArtifact{
			Repository: "test_repo10",
			Digest:     "test_digest10",
			Size:       int64(100),
			Vendor:     "test_vendor10",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}

		sa_vendor2 := model.SystemArtifact{
			Repository: "test_repo20",
			Digest:     "test_digest20",
			Size:       int64(100),
			Vendor:     "test_vendor20",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}

		idVendor1, err := suite.dao.Create(suite.ctx, &sa_vendor1)
		idVendor2, err := suite.dao.Create(suite.ctx, &sa_vendor2)

		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, idVendor1, "Expected a valid record identifier but was 0")
		suite.NotEqual(0, idVendor2, "Expected a valid record identifier but was 0")

		saRead, err := suite.dao.Get(suite.ctx, "test_vendor10", "test_repo10", "test_digest10")
		saRead2, err := suite.dao.Get(suite.ctx, "test_vendor20", "test_repo20", "test_digest20")

		suite.NoErrorf(err, "Unexpected error when reading system artifact: %v", err)
		suite.Equalf(idVendor1, saRead.ID, "The ID for inserted system record %d is not equal to the read system record %d", idVendor1, saRead.ID)
		suite.Equalf(idVendor2, saRead2.ID, "The ID for inserted system record %d is not equal to the read system record %d", idVendor2, saRead2.ID)
	}
}

func (suite *daoTestSuite) TestDelete() {

	// insert a normal system artifact and attempt to get it
	{
		sa := model.SystemArtifact{
			Repository: "test_repo3",
			Digest:     "test_digest3",
			Size:       int64(100),
			Vendor:     "test_vendor",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}
		id, err := suite.dao.Create(suite.ctx, &sa)
		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, id, "Expected a valid record identifier but was 0")

		err = suite.dao.Delete(suite.ctx, "test_vendor", "test_repo3", "test_digest3")
		suite.NoErrorf(err, "Unexpected error when reading system artifact: %v", err)
	}

	// attempt to delete a non-existent system artifact record with incorrect repo name and correct digest
	{
		err := suite.dao.Delete(suite.ctx, "test_vendor", "test_repo4", "test_digest3")
		suite.NoErrorf(err, "Attempt to delete a non-existent system artifact should not fail")
	}

	// attempt to retrieve a non-existent system artifact record with correct repo name and incorrect digest
	{
		err := suite.dao.Delete(suite.ctx, "test_vendor", "test_repo3", "test_digest4")
		suite.NoErrorf(err, "Attempt to delete a non-existent system artifact should not fail")
	}

	// multiple system artifact records from different vendors.
	// insert a normal system artifact and attempt to get it
	{
		sa_vendor1 := model.SystemArtifact{
			Repository: "test_repo200",
			Digest:     "test_digest200",
			Size:       int64(100),
			Vendor:     "test_vendor200",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}

		sa_vendor2 := model.SystemArtifact{
			Repository: "test_repo300",
			Digest:     "test_digest300",
			Size:       int64(100),
			Vendor:     "test_vendor300",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}

		idVendor1, err := suite.dao.Create(suite.ctx, &sa_vendor1)
		idVendor2, err := suite.dao.Create(suite.ctx, &sa_vendor2)

		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, idVendor1, "Expected a valid record identifier but was 0")
		suite.NotEqual(0, idVendor2, "Expected a valid record identifier but was 0")

		saRead, err := suite.dao.Get(suite.ctx, "test_vendor200", "test_repo200", "test_digest200")
		saRead2, err := suite.dao.Get(suite.ctx, "test_vendor300", "test_repo300", "test_digest300")

		suite.NoErrorf(err, "Unexpected error when reading system artifact: %v", err)
		suite.Equalf(idVendor1, saRead.ID, "The ID for inserted system record %d is not equal to the read system record %d", idVendor1, saRead.ID)
		suite.Equalf(idVendor2, saRead2.ID, "The ID for inserted system record %d is not equal to the read system record %d", idVendor2, saRead2.ID)

		err = suite.dao.Delete(suite.ctx, "test_vendor200", "test_repo200", "test_digest200")

		suite.NoErrorf(err, "Unexpected error when reading system artifact: %v", err)
		saRead, err = suite.dao.Get(suite.ctx, "test_vendor200", "test_repo200", "test_digest200")
		suite.Errorf(err, "Expected no record found error for provided repository and digest")
		suite.Nil(saRead, "Expected system artifact record to be nil")
		errWithInfo := err.(*errors.Error)
		suite.Equalf(errors.NotFoundCode, errWithInfo.Code, "Expected not found code but was %s", errWithInfo.Code)

		saRead3, err := suite.dao.Get(suite.ctx, "test_vendor300", "test_repo300", "test_digest300")
		suite.Equalf(idVendor2, saRead2.ID, "The ID for inserted system record %d is not equal to the read system record %d", idVendor2, saRead3.ID)
	}
}

func (suite *daoTestSuite) TestList() {
	expectedSystemArtifactIds := make(map[int64]bool)

	// insert a normal system artifact
	{
		sa := model.SystemArtifact{
			Repository: "test_repo4",
			Digest:     "test_digest4",
			Size:       int64(100),
			Vendor:     "test_vendor",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}
		id, err := suite.dao.Create(suite.ctx, &sa)
		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, id, "Expected a valid record identifier but was 0")
		expectedSystemArtifactIds[id] = true
	}

	// attempt to read all the system artifact records
	{
		query := q.Query{}
		query.Keywords = map[string]interface{}{"repository": "test_repo4", "digest": "test_digest4"}

		sysArtifacts, err := suite.dao.List(suite.ctx, &query)
		suite.NotNilf(sysArtifacts, "Expected system artifacts list to be non-nil")
		suite.NoErrorf(err, "Unexpected error when listing system artifact records : %v", err)
		suite.Equalf(1, len(sysArtifacts), "Expected system artifacts list of size 1 but was: %d", len(sysArtifacts))

		// iterate through the system artifact and validate that the ids are in the expected list of ids
		for _, sysArtifact := range sysArtifacts {
			_, ok := expectedSystemArtifactIds[sysArtifact.ID]
			suite.Truef(ok, "Expected system artifact id %d to be present but was absent", sysArtifact.ID)
		}
	}
}

func (suite *daoTestSuite) TestSize() {
	// insert a normal system artifact
	{
		sa := model.SystemArtifact{
			Repository: "test_repo8",
			Digest:     "test_digest8",
			Size:       int64(100),
			Vendor:     "test_vendor",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}
		id, err := suite.dao.Create(suite.ctx, &sa)
		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, id, "Expected a valid record identifier but was 0")

		sa1 := model.SystemArtifact{
			Repository: "test_repo9",
			Digest:     "test_digest9",
			Size:       int64(500),
			Vendor:     "test_vendor",
			Type:       "test_repo_type",
			CreateTime: time.Now(),
			ExtraAttrs: "",
		}
		id, err = suite.dao.Create(suite.ctx, &sa1)
		suite.NoError(err, "Unexpected error when inserting test record")
		suite.NotEqual(0, id, "Expected a valid record identifier but was 0")

		size, err := suite.dao.Size(suite.ctx)
		suite.NoError(err, "Unexpected error when calculating record size")
		suite.Truef(size > int64(0), "Expected size to be non-zero")
	}
}
func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
