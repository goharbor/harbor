package handler

import (
	"testing"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/q"
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

func isAuthOnlyQuery(query *q.Query) bool {
	return query != nil && query.Keywords["public"] == "auth_only"
}

func (suite *StatisticsTestSuite) TestGetStatistic() {
	// List is called once (for public projects); the auth_only count now goes
	// through Count.  Return an empty slice so the public count is 0.
	suite.projectCtl.On("List", mock.Anything, mock.Anything, mock.Anything).Return([]*models.Project{}, nil)
	// auth_only count query returns 0; the total-project count (nil query) returns 10.
	suite.projectCtl.On("Count", mock.Anything, testifymock.MatchedBy(isAuthOnlyQuery)).Return(int64(0), nil)
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
	suite.Equal(int64(0), statistics.PublicProjectCount)
	suite.Equal(int64(0), statistics.AuthOnlyProjectCount)
	// private = total(10) - public(0) - auth_only(0)
	suite.Equal(int64(10), statistics.PrivateProjectCount)
	suite.Equal(int64(20), statistics.PrivateRepoCount)
	suite.Equal(int64(20), statistics.TotalRepoCount)
}

func (suite *StatisticsTestSuite) TestGetStatisticWithAuthOnly() {
	publicProject := &models.Project{ProjectID: 2, Name: "public-proj"}

	// Clear accumulated expectations from previous tests so the ordered Once()
	// expectations below are matched in the right sequence.
	suite.projectCtl.ExpectedCalls = nil
	suite.repoCtl.ExpectedCalls = nil
	suite.blobCtl.ExpectedCalls = nil
	suite.sysArtifactMgr.ExpectedCalls = nil

	// List returns the public project; the auth_only count now goes through Count.
	suite.projectCtl.On("List", mock.Anything, mock.Anything, mock.Anything).
		Return([]*models.Project{publicProject}, nil).Once()
	// auth_only count query returns 1; the total-project count (nil query) returns 5.
	suite.projectCtl.On("Count", mock.Anything, testifymock.MatchedBy(isAuthOnlyQuery)).Return(int64(1), nil)
	suite.projectCtl.On("Count", mock.Anything, mock.Anything).Return(int64(5), nil)
	suite.repoCtl.On("Count", mock.Anything, mock.Anything).Return(int64(10), nil)
	suite.blobCtl.On("CalculateTotalSize", mock.Anything, true).Return(int64(500), nil)
	suite.sysArtifactMgr.On("GetStorageSize", mock.Anything).Return(int64(500), nil)

	suite.Security.On("IsAuthenticated").Return(true)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.Security.On("IsSysAdmin").Return(true)

	var statistics models2.Statistic
	res, err := suite.GetJSON("/statistics", &statistics)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal(int64(1), statistics.PublicProjectCount)
	suite.Equal(int64(1), statistics.AuthOnlyProjectCount)
	// private = total(5) - public(1) - auth_only(1) = 3
	suite.Equal(int64(3), statistics.PrivateProjectCount)
}
func TestStatisticsSuite(t *testing.T) {
	suite.Run(t, &StatisticsTestSuite{})
}
