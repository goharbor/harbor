package dao

import (
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/policy/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	testPly1 = &model.Policy{
		Name:         "webhook test policy1",
		Description:  "webhook test policy1 description",
		ProjectID:    111,
		TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
		EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
		Creator:      "no one",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
		Enabled:      true,
	}
)

var (
	testPly2 = &model.Policy{
		Name:         "webhook test policy2",
		Description:  "webhook test policy2 description",
		ProjectID:    222,
		TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
		EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
		Creator:      "no one",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
		Enabled:      true,
	}
)

var (
	testPly3 = &model.Policy{
		Name:         "webhook test policy3",
		Description:  "webhook test policy3 description",
		ProjectID:    333,
		TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
		EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
		Creator:      "no one",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
		Enabled:      true,
	}
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO

	jobID1 int64
	jobID2 int64
	jobID3 int64
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = New()
	suite.Suite.ClearTables = []string{"notification_policy"}
	suite.policies()
}

func (suite *DaoTestSuite) policies() {
	var err error
	suite.jobID1, err = suite.dao.Create(orm.Context(), testPly1)
	suite.Nil(err)

	suite.jobID2, err = suite.dao.Create(orm.Context(), testPly2)
	suite.Nil(err)

	suite.jobID3, err = suite.dao.Create(orm.Context(), testPly3)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestCreate() {
	_, err := suite.dao.Create(orm.Context(), nil)
	suite.NotNil(err)
}

func (suite *DaoTestSuite) TestDelete() {
	err := suite.dao.Delete(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	err = suite.dao.Delete(orm.Context(), suite.jobID2)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestList() {
	jobs, err := suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"ProjectID": 333,
		},
	})
	suite.Require().Nil(err)
	suite.Equal(len(jobs), 1)
	suite.Equal(suite.jobID3, jobs[0].ID)
}

func (suite *DaoTestSuite) TestGet() {
	_, err := suite.dao.Get(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	id, err := suite.dao.Create(orm.Context(), &model.Policy{
		Name:         "webhook test policy4",
		Description:  "webhook test policy4 description",
		ProjectID:    444,
		TargetsDB:    "[{\"type\":\"http\",\"address\":\"http://10.173.32.58:9009\",\"token\":\"xxxxxxxxx\",\"skip_cert_verify\":true}]",
		EventTypesDB: "[\"pushImage\",\"pullImage\",\"deleteImage\",\"uploadChart\",\"deleteChart\",\"downloadChart\",\"scanningFailed\",\"scanningCompleted\"]",
		Creator:      "no one",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
		Enabled:      true,
	})
	suite.Nil(err)

	r, err := suite.dao.Get(orm.Context(), id)
	suite.Nil(err)
	suite.Equal("webhook test policy4", r.Name)
}

func (suite *DaoTestSuite) TestUpdate() {
	j := &model.Policy{
		ID:      suite.jobID1,
		Enabled: false,
	}

	err := suite.dao.Update(orm.Context(), j)
	suite.Nil(err)

	r1, err := suite.dao.Get(orm.Context(), j.ID)
	suite.False(r1.Enabled)
}

func (suite *DaoTestSuite) TestCount() {
	// nil query
	total, err := suite.dao.Count(orm.Context(), nil)
	suite.Nil(err)
	suite.True(total > 0)

	total, err = suite.dao.Count(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"ProjectID": 111,
		},
	})
	suite.Nil(err)
	suite.Equal(int64(1), total)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
