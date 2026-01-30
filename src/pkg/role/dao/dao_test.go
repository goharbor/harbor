package dao

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/role/model"
	htesting "github.com/goharbor/harbor/src/testing"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO

	roleID1 int64
	roleID2 int64
	roleID3 int64
	roleID4 int64
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = New()
	suite.Suite.ClearTables = []string{"role"}
	suite.roles()
}

func (suite *DaoTestSuite) roles() {
	var err error
	suite.roleID1, err = suite.dao.Create(orm.Context(), &model.Role{
		Name: "test1",
	})
	suite.Nil(err)

	suite.roleID2, err = suite.dao.Create(orm.Context(), &model.Role{
		Name: "test2",
	})
	suite.Nil(err)

	suite.roleID3, err = suite.dao.Create(orm.Context(), &model.Role{
		Name: "test3",
	})
	suite.Nil(err)

	suite.roleID4, err = suite.dao.Create(orm.Context(), &model.Role{
		Name: "test4",
	})
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestCreate() {
	r := &model.Role{
		Name: "test1",
	}
	_, err := suite.dao.Create(orm.Context(), r)
	suite.NotNil(err)
	suite.True(errors.IsErr(err, errors.ConflictCode))
}

func (suite *DaoTestSuite) TestDelete() {
	err := suite.dao.Delete(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	err = suite.dao.Delete(orm.Context(), suite.roleID2)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestList() {
	roles, err := suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]any{
			"name": "test3",
		},
	})
	suite.Require().Nil(err)
	suite.Equal(suite.roleID3, roles[0].ID)

	r := &model.Role{
		Name: "testvisible",
	}
	_, err = suite.dao.Create(orm.Context(), r)
	suite.Nil(err)
	roles, err = suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]any{
			"name": "testvisible",
		},
	})
	suite.Equal(len(roles), 0)
}

func (suite *DaoTestSuite) TestGet() {
	_, err := suite.dao.Get(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	r, err := suite.dao.Get(orm.Context(), suite.roleID3)
	suite.Nil(err)
	suite.Equal("test3", r.Name)
}

func (suite *DaoTestSuite) TestCount() {
	// nil query
	total, err := suite.dao.Count(orm.Context(), nil)
	suite.Nil(err)
	suite.True(total > 0)

	// query by name
	total, err = suite.dao.Count(orm.Context(), &q.Query{
		Keywords: map[string]any{
			"name": "test3",
		},
	})
	suite.Nil(err)
	suite.Equal(int64(1), total)
}

func (suite *DaoTestSuite) TestUpdate() {
	r := &model.Role{
		ID: suite.roleID3,
	}

	err := suite.dao.Update(orm.Context(), r)
	suite.Nil(err)

}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
