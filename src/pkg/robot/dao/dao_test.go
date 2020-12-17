package dao

import (
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO

	robotID1 int64
	robotID2 int64
	robotID3 int64
	robotID4 int64
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = New()
	suite.Suite.ClearTables = []string{"robot"}
	suite.robots()
}

func (suite *DaoTestSuite) robots() {
	var err error
	suite.robotID1, err = suite.dao.Create(orm.Context(), &model.Robot{
		Name:        "test1",
		Description: "test1 description",
		ProjectID:   1,
		Secret:      suite.RandString(10),
	})
	suite.Nil(err)

	suite.robotID2, err = suite.dao.Create(orm.Context(), &model.Robot{
		Name:        "test2",
		Description: "test2 description",
		ProjectID:   1,
		Secret:      suite.RandString(10),
	})
	suite.Nil(err)

	suite.robotID3, err = suite.dao.Create(orm.Context(), &model.Robot{
		Name:        "test3",
		Description: "test3 description",
		ProjectID:   1,
		Secret:      suite.RandString(10),
	})
	suite.Nil(err)

	suite.robotID4, err = suite.dao.Create(orm.Context(), &model.Robot{
		Name:        "test4",
		Description: "test4 description",
		ProjectID:   2,
		Secret:      suite.RandString(10),
	})
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestCreate() {
	r := &model.Robot{
		Name:        "test1",
		Description: "test1 description",
		ProjectID:   1,
		Secret:      suite.RandString(10),
	}
	_, err := suite.dao.Create(orm.Context(), r)
	suite.NotNil(err)
	suite.True(errors.IsErr(err, errors.ConflictCode))
}

func (suite *DaoTestSuite) TestDelete() {
	err := suite.dao.Delete(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	err = suite.dao.Delete(orm.Context(), suite.robotID2)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestList() {
	robots, err := suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"name": "test3",
		},
	})
	suite.Require().Nil(err)
	suite.Equal(suite.robotID3, robots[0].ID)

	r := &model.Robot{
		Name:        "testvisible",
		Description: "test visible",
		ProjectID:   998,
		Visible:     false,
		Secret:      suite.RandString(10),
	}
	_, err = suite.dao.Create(orm.Context(), r)
	suite.Nil(err)
	robots, err = suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"name":    "testvisible",
			"visible": true,
		},
	})
	suite.Equal(len(robots), 0)
}

func (suite *DaoTestSuite) TestGet() {
	_, err := suite.dao.Get(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	r, err := suite.dao.Get(orm.Context(), suite.robotID3)
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
		Keywords: map[string]interface{}{
			"name": "test3",
		},
	})
	suite.Nil(err)
	suite.Equal(int64(1), total)
}

func (suite *DaoTestSuite) TestUpdate() {
	r := &model.Robot{
		ID:          suite.robotID3,
		Description: "after test3 update",
	}

	err := suite.dao.Update(orm.Context(), r)
	suite.Nil(err)

	r1, err := suite.dao.Get(orm.Context(), r.ID)
	suite.Equal("after test3 update", r1.Description)
}

func (suite *DaoTestSuite) TestDeleteByProjectID() {
	robots, err := suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"project_id": 2,
		},
	})
	suite.Equal(1, len(robots))

	err = suite.dao.DeleteByProjectID(orm.Context(), 2)
	suite.Nil(err)

	robots, err = suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"project_id": 2,
		},
	})
	suite.Equal(0, len(robots))
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
