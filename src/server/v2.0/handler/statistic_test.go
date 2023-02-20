package handler

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/project/models"
	models2 "github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	blobtesting "github.com/goharbor/harbor/src/testing/controller/blob"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	repositorytesting "github.com/goharbor/harbor/src/testing/controller/repository"
	"github.com/goharbor/harbor/src/testing/mock"
	systemartifacttesting "github.com/goharbor/harbor/src/testing/pkg/systemartifact"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

type StatisticsTestSuite struct {
	htesting.Suite
	projectCtl     *projecttesting.Controller
	repoCtl        *repositorytesting.Controller
	blobCtl        *blobtesting.Controller
	sysArtifactMgr *systemartifacttesting.Manager
}

func (suite *StatisticsTestSuite) SetupSuite() {
	suite.projectCtl = &projecttesting.Controller{}
	suite.repoCtl = &repositorytesting.Controller{}
	suite.blobCtl = &blobtesting.Controller{}
	suite.sysArtifactMgr = &systemartifacttesting.Manager{}

	suite.Config = &restapi.Config{StatisticAPI: &statisticAPI{
		proCtl:            suite.projectCtl,
		repoCtl:           suite.repoCtl,
		blobCtl:           suite.blobCtl,
		systemArtifactMgr: suite.sysArtifactMgr,
	}}
	suite.Suite.SetupSuite()
}

func (suite *StatisticsTestSuite) TestGetStatistic() {
	projects := make([]*models.Project, 0)
	suite.projectCtl.On("List", mock.Anything, mock.Anything, mock.Anything).Return(projects, nil)
	suite.projectCtl.On("Count", mock.Anything, mock.Anything).Return(int64(10), nil)
	suite.repoCtl.On("Count", mock.Anything, mock.Anything).Return(int64(20), nil)
	suite.blobCtl.On("CalculateTotalSize", mock.Anything, true).Return(int64(1000), nil)
	suite.sysArtifactMgr.On("GetStorageSize", mock.Anything).Return(int64(1000), nil)

	suite.Security.On("IsAuthenticated").Return(true)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.Security.On("IsSysAdmin").Return(true)

	var statistics models2.Statistic
	res, err := suite.GetJSON("/statistics", &statistics)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal(int64(2000), statistics.TotalStorageConsumption)
	suite.Equal(int64(10), statistics.PrivateProjectCount)
	suite.Equal(int64(20), statistics.PrivateRepoCount)
	suite.Equal(int64(20), statistics.TotalRepoCount)
}
func TestStatisticsSuite(t *testing.T) {
	suite.Run(t, &StatisticsTestSuite{})
}
