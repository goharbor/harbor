package dao

import (
	"github.com/goharbor/harbor/src/pkg/joblog/models"
	htesting "github.com/goharbor/harbor/src/testing"
	"time"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"job_log"}
	suite.dao = New()
}

func (suite *DaoTestSuite) TestMethodsOfJobLog() {
	ctx := suite.Context()

	uuid := "uuid_for_unit_test"
	now := time.Now()
	content := "content for unit text"
	jobLog := &models.JobLog{
		UUID:         uuid,
		CreationTime: now,
		Content:      content,
	}

	// create
	_, err := suite.dao.Create(ctx, jobLog)
	suite.Nil(err)

	// update
	updateContent := "content for unit text update"
	jobLog.Content = updateContent
	_, err = suite.dao.Create(ctx, jobLog)
	suite.Nil(err)

	// get
	log, err := suite.dao.Get(ctx, uuid)
	suite.Nil(err)
	suite.Equal(now.Second(), log.CreationTime.Second())
	suite.Equal(updateContent, log.Content)
	suite.Equal(jobLog.LogID, log.LogID)

	// delete
	count, err := suite.dao.DeleteBefore(ctx, time.Now().Add(time.Duration(time.Minute)))
	suite.Nil(err)
	suite.Equal(int64(1), count)
}
