package dao

import (
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/job/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MysqlDaoTestSuite struct {
	htesting.Suite
	dao DAO

	jobID1 int64
	jobID2 int64
	jobID3 int64
}

func (suite *MysqlDaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = NewMysqlDao()
	suite.Suite.ClearTables = []string{"notification_job"}
	suite.jobs()
}

func (suite *MysqlDaoTestSuite) jobs() {
	var err error
	suite.jobID1, err = suite.dao.Create(orm.Context(), testJob1)
	suite.Nil(err)

	suite.jobID2, err = suite.dao.Create(orm.Context(), testJob2)
	suite.Nil(err)

	suite.jobID3, err = suite.dao.Create(orm.Context(), testJob3)
	suite.Nil(err)
}

func (suite *MysqlDaoTestSuite) TestCreate() {
	_, err := suite.dao.Create(orm.Context(), nil)
	suite.NotNil(err)
}

func (suite *MysqlDaoTestSuite) TestDelete() {
	err := suite.dao.Delete(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	err = suite.dao.Delete(orm.Context(), suite.jobID2)
	suite.Nil(err)
}

func (suite *MysqlDaoTestSuite) TestList() {
	jobs, err := suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"EventType": "pushImage",
		},
	})
	suite.Require().Nil(err)
	suite.Equal(len(jobs), 1)
	suite.Equal(suite.jobID1, jobs[0].ID)
}

func (suite *MysqlDaoTestSuite) TestGet() {
	_, err := suite.dao.Get(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	id, err := suite.dao.Create(orm.Context(), &model.Job{
		PolicyID:   2222,
		EventType:  "pushChart",
		NotifyType: "http",
		Status:     "pending",
		JobDetail:  "{\"type\":\"pushImage\",\"occur_at\":1563536782,\"event_data\":{\"resources\":[{\"digest\":\"sha256:bf1684a6e3676389ec861c602e97f27b03f14178e5bc3f70dce198f9f160cce9\",\"tag\":\"v1.0\",\"resource_url\":\"10.194.32.23/myproj/alpine:v1.0\"}],\"repository\":{\"date_created\":1563505587,\"name\":\"alpine\",\"namespace\":\"myproj\",\"repo_full_name\":\"myproj/alpine\",\"repo_type\":\"private\"}},\"operator\":\"admin\"}",
		UUID:       "00000000",
	})
	suite.Nil(err)

	r, err := suite.dao.Get(orm.Context(), id)
	suite.Nil(err)
	suite.Equal("pushChart", r.EventType)
}

func (suite *MysqlDaoTestSuite) TestUpdate() {
	j := &model.Job{
		ID:     suite.jobID1,
		Status: "success",
	}

	err := suite.dao.Update(orm.Context(), j)
	suite.Nil(err)

	r1, err := suite.dao.Get(orm.Context(), j.ID)
	suite.Equal("success", r1.Status)
}

func (suite *MysqlDaoTestSuite) TestCount() {
	// nil query
	total, err := suite.dao.Count(orm.Context(), nil)
	suite.Nil(err)
	suite.True(total > 0)

	// query by name
	total, err = suite.dao.Count(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"EventType": "deleteImage",
		},
	})
	suite.Nil(err)
	suite.Equal(int64(1), total)
}

func (suite *MysqlDaoTestSuite) TestDeleteByPolicyID() {
	jobs, err := suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"PolicyID": 111,
		},
	})
	suite.True(len(jobs) > 0)

	err = suite.dao.DeleteByPolicyID(orm.Context(), 111)
	suite.Nil(err)

	jobs, err = suite.dao.List(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"PolicyID": 111,
		},
	})
	suite.Equal(0, len(jobs))
}

func (suite *MysqlDaoTestSuite) TestGetLastTriggerJobsGroupByEventType() {
	_, err := suite.dao.Create(orm.Context(), &model.Job{
		PolicyID:   3333,
		EventType:  "pushChart",
		NotifyType: "http",
		Status:     "pending",
		JobDetail:  "{\"type\":\"pushImage\",\"occur_at\":1563536782,\"event_data\":{\"resources\":[{\"digest\":\"sha256:bf1684a6e3676389ec861c602e97f27b03f14178e5bc3f70dce198f9f160cce9\",\"tag\":\"v1.0\",\"resource_url\":\"10.194.32.23/myproj/alpine:v1.0\"}],\"repository\":{\"date_created\":1563505587,\"name\":\"alpine\",\"namespace\":\"myproj\",\"repo_full_name\":\"myproj/alpine\",\"repo_type\":\"private\"}},\"operator\":\"admin\"}",
		UUID:       "00000000",
	})
	suite.Nil(err)
	_, err = suite.dao.Create(orm.Context(), &model.Job{
		PolicyID:   3333,
		EventType:  "pullChart",
		NotifyType: "http",
		Status:     "pending",
		JobDetail:  "{\"type\":\"pushImage\",\"occur_at\":1563536782,\"event_data\":{\"resources\":[{\"digest\":\"sha256:bf1684a6e3676389ec861c602e97f27b03f14178e5bc3f70dce198f9f160cce9\",\"tag\":\"v1.0\",\"resource_url\":\"10.194.32.23/myproj/alpine:v1.0\"}],\"repository\":{\"date_created\":1563505587,\"name\":\"alpine\",\"namespace\":\"myproj\",\"repo_full_name\":\"myproj/alpine\",\"repo_type\":\"private\"}},\"operator\":\"admin\"}",
		UUID:       "00000000",
	})
	suite.Nil(err)
	jobs, err := suite.dao.GetLastTriggerJobsGroupByEventType(orm.Context(), 3333)
	suite.Nil(err)
	suite.Equal(2, len(jobs))
}

func TestMysqlDaoTestSuite(t *testing.T) {
	if !utils.IsDBMysql() {
		return
	}
	suite.Run(t, &MysqlDaoTestSuite{})
}
