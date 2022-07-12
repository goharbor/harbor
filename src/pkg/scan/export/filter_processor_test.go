package export

import (
	"context"
	"errors"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	project3 "github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	tag2 "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	"github.com/goharbor/harbor/src/testing/pkg/tag"
	"github.com/goharbor/harbor/src/testing/pkg/user"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"reflect"
	"testing"
	"time"
)

type FilterProcessorTestSuite struct {
	suite.Suite
	repoMgr         *repository.Manager
	tagMgr          *tag.FakeManager
	usrMgr          *user.Manager
	projectMgr      *project.Manager
	filterProcessor FilterProcessor
}

func (suite *FilterProcessorTestSuite) SetupSuite() {

}

func (suite *FilterProcessorTestSuite) SetupTest() {
	suite.repoMgr = &repository.Manager{}
	suite.tagMgr = &tag.FakeManager{}
	suite.usrMgr = &user.Manager{}
	suite.projectMgr = &project.Manager{}
	suite.filterProcessor = &DefaultFilterProcessor{
		repoMgr:    suite.repoMgr,
		tagMgr:     suite.tagMgr,
		usrMgr:     suite.usrMgr,
		projectMgr: suite.projectMgr,
	}
}

func (suite *FilterProcessorTestSuite) TestProcessProjectFilter() {
	project1 := &models.Project{ProjectID: 1}

	project2 := &models.Project{ProjectID: 2}

	// no filtered projects returns all projects
	{
		suite.usrMgr.On("GetByName", mock.Anything, "test-user").Return(&commonmodels.User{UserID: 1}, nil).Once()
		suite.projectMgr.On("List", mock.Anything, mock.Anything).Return([]*models.Project{project1, project2}, nil).Once()
		projectIds, err := suite.filterProcessor.ProcessProjectFilter(context.TODO(), "test-user", []int64{})
		suite.Equal(2, len(projectIds))
		suite.NoError(err)
	}

	// filtered project
	{
		suite.usrMgr.On("GetByName", mock.Anything, "test-user").Return(&commonmodels.User{UserID: 1}, nil).Once()
		suite.projectMgr.On("List", mock.Anything, mock.Anything).Return([]*models.Project{project1, project2}, nil).Once()
		projectIds, err := suite.filterProcessor.ProcessProjectFilter(context.TODO(), "test-user", []int64{1})
		suite.Equal(1, len(projectIds))
		suite.Equal(int64(1), projectIds[0])
		suite.NoError(err)
	}

	// filtered project with group ids
	{
		groupIDs := []int{4, 5}
		suite.usrMgr.On("GetByName", mock.Anything, "test-user").Return(&commonmodels.User{UserID: 1, GroupIDs: groupIDs}, nil).Once()
		suite.projectMgr.On("List", mock.Anything, mock.Anything).Return([]*models.Project{project1, project2}, nil).Once()
		projectIds, err := suite.filterProcessor.ProcessProjectFilter(context.TODO(), "test-user", []int64{1})
		suite.Equal(1, len(projectIds))
		suite.Equal(int64(1), projectIds[0])
		suite.NoError(err)
		memberQueryMatcher := testifymock.MatchedBy(func(query *q.Query) bool {
			memberQuery := query.Keywords["member"].(*project3.MemberQuery)
			return len(memberQuery.GroupIDs) == 2 && reflect.DeepEqual(memberQuery.GroupIDs, groupIDs) && memberQuery.Role == 0
		})
		suite.projectMgr.AssertCalled(suite.T(), "List", mock.Anything, memberQueryMatcher)
	}

	// project listing for admin user
	{
		suite.usrMgr.On("GetByName", mock.Anything, "test-user").Return(&commonmodels.User{UserID: 1, SysAdminFlag: true}, nil).Once()
		suite.projectMgr.On("List", mock.Anything, mock.Anything).Return([]*models.Project{project1, project2}, nil).Once()
		_, err := suite.filterProcessor.ProcessProjectFilter(context.TODO(), "test-user", []int64{1})
		suite.NoError(err)
		queryArgumentMatcher := testifymock.MatchedBy(func(query *q.Query) bool {
			return len(query.Keywords) == 0
		})
		suite.projectMgr.AssertCalled(suite.T(), "List", mock.Anything, queryArgumentMatcher)
	}

	// project listing returns an error
	// filtered project
	{
		suite.usrMgr.On("GetByName", mock.Anything, "test-user").Return(&commonmodels.User{UserID: 1}, nil).Once()
		suite.projectMgr.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("test-error")).Once()
		projectIds, err := suite.filterProcessor.ProcessProjectFilter(context.TODO(), "test-user", []int64{1})
		suite.Error(err)
		suite.Nil(projectIds)
	}

}

func (suite *FilterProcessorTestSuite) TestProcessRepositoryFilter() {

	repoRecord1 := model.RepoRecord{
		RepositoryID: int64(1),
		Name:         "test/repo1",
		ProjectID:    int64(100),
		Description:  "test repo 1",
		PullCount:    1,
		StarCount:    4,
		CreationTime: time.Time{},
		UpdateTime:   time.Time{},
	}
	repoRecord2 := model.RepoRecord{
		RepositoryID: int64(1),
		Name:         "test/repo2",
		ProjectID:    int64(100),
		Description:  "test repo 2",
		PullCount:    1,
		StarCount:    4,
		CreationTime: time.Time{},
		UpdateTime:   time.Time{},
	}

	allRepos := make([]*model.RepoRecord, 0)
	allRepos = append(allRepos, &repoRecord1, &repoRecord2)

	// filter required repositories
	{
		suite.repoMgr.On("List", mock.Anything, mock.Anything).Return(allRepos, nil).Once()
		candidates, err := suite.filterProcessor.ProcessRepositoryFilter(context.TODO(), "repo1", []int64{100})
		suite.NoError(err)
		suite.Equal(1, len(candidates), "Expected 1 candidate but found ", len(candidates))
		suite.Equal("repo1", candidates[0].Repository)
	}

	// simulate repo manager returning an error
	{
		suite.repoMgr.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("test error")).Once()
		candidates, err := suite.filterProcessor.ProcessRepositoryFilter(context.TODO(), "repo1", []int64{100})
		suite.Error(err)
		suite.Nil(candidates)
	}

	// simulate doublestar filtering
	{
		repoRecord3 := model.RepoRecord{
			RepositoryID: int64(1),
			Name:         "test/repo1/ubuntu",
			ProjectID:    int64(100),
			Description:  "test repo 1",
			PullCount:    1,
			StarCount:    4,
			CreationTime: time.Time{},
			UpdateTime:   time.Time{},
		}
		repoRecord4 := model.RepoRecord{
			RepositoryID: int64(1),
			Name:         "test/repo1/centos",
			ProjectID:    int64(100),
			Description:  "test repo 2",
			PullCount:    1,
			StarCount:    4,
			CreationTime: time.Time{},
			UpdateTime:   time.Time{},
		}
		allRepos = append(allRepos, &repoRecord3, &repoRecord4)
		suite.repoMgr.On("List", mock.Anything, mock.Anything).Return(allRepos, nil).Once()
		candidates, err := suite.filterProcessor.ProcessRepositoryFilter(context.TODO(), "repo1/**", []int64{100})
		suite.NoError(err)
		suite.Equal(2, len(candidates), "Expected 2 candidate but found ", len(candidates))
		m := map[string]bool{}
		for _, cand := range candidates {
			m[cand.Repository] = true
		}
		_, ok := m["repo1/ubuntu"]
		suite.True(ok)
		_, ok = m["repo1/centos"]
		suite.True(ok)
	}
}

func (suite *FilterProcessorTestSuite) TestProcessTagFilter() {

	testTag1 := tag2.Tag{
		ID:           int64(1),
		RepositoryID: int64(1),
		ArtifactID:   int64(1),
		Name:         "test-tag1",
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}

	testTag2 := tag2.Tag{
		ID:           int64(2),
		RepositoryID: int64(1),
		ArtifactID:   int64(1),
		Name:         "test-tag2",
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}

	testTag3 := tag2.Tag{
		ID:           int64(3),
		RepositoryID: int64(2),
		ArtifactID:   int64(2),
		Name:         "test-tag3",
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}

	allTags := make([]*tag2.Tag, 0)

	allTags = append(allTags, &testTag1, &testTag2)

	// filter required repositories haveing the specified tags
	{
		suite.tagMgr.On("List", mock.Anything, mock.Anything).Return([]*tag2.Tag{&testTag1, &testTag2}, nil).Once()
		suite.tagMgr.On("List", mock.Anything, mock.Anything).Return([]*tag2.Tag{&testTag3}, nil).Once()

		candidates, err := suite.filterProcessor.ProcessTagFilter(context.TODO(), "*tag2", []int64{1, 2})
		suite.NoError(err)
		suite.Equal(1, len(candidates), "Expected 1 candidate but found ", len(candidates))
		suite.Equal(int64(1), candidates[0].NamespaceID)
	}

	// simulate repo manager returning an error
	{
		suite.tagMgr.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("test error")).Once()
		candidates, err := suite.filterProcessor.ProcessTagFilter(context.TODO(), "repo1", []int64{1, 2})
		suite.Error(err)
		suite.Nil(candidates)
	}

}

func TestFilterProcessorTestSuite(t *testing.T) {
	suite.Run(t, &FilterProcessorTestSuite{})
}
