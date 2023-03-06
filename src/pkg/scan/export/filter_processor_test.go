package export

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/tag"
	artpkg "github.com/goharbor/harbor/src/pkg/artifact"
	labelmodel "github.com/goharbor/harbor/src/pkg/label/model"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	artifactctl "github.com/goharbor/harbor/src/testing/controller/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	"github.com/goharbor/harbor/src/testing/pkg/user"
)

type FilterProcessorTestSuite struct {
	suite.Suite
	artCtl          *artifactctl.Controller
	repoMgr         *repository.Manager
	usrMgr          *user.Manager
	projectMgr      *project.Manager
	filterProcessor FilterProcessor
}

func (suite *FilterProcessorTestSuite) SetupSuite() {

}

func (suite *FilterProcessorTestSuite) SetupTest() {
	suite.artCtl = &artifactctl.Controller{}
	suite.repoMgr = &repository.Manager{}
	suite.usrMgr = &user.Manager{}
	suite.projectMgr = &project.Manager{}
	suite.filterProcessor = &DefaultFilterProcessor{
		artCtl:     suite.artCtl,
		repoMgr:    suite.repoMgr,
		usrMgr:     suite.usrMgr,
		projectMgr: suite.projectMgr,
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
		RepositoryID: int64(2),
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
		suite.Equal(int64(1), candidates[0])
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
			RepositoryID: int64(3),
			Name:         "test/repo1/ubuntu",
			ProjectID:    int64(100),
			Description:  "test repo 1",
			PullCount:    1,
			StarCount:    4,
			CreationTime: time.Time{},
			UpdateTime:   time.Time{},
		}
		repoRecord4 := model.RepoRecord{
			RepositoryID: int64(4),
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
		m := map[int64]bool{}
		for _, cand := range candidates {
			m[cand] = true
		}
		_, ok := m[3]
		suite.True(ok)
		_, ok = m[4]
		suite.True(ok)
	}
}

func (suite *FilterProcessorTestSuite) TestProcessTagFilter() {
	tag1 := &tag.Tag{Tag: tagmodel.Tag{ID: int64(1), Name: "tag1"}}
	tag2 := &tag.Tag{Tag: tagmodel.Tag{ID: int64(2), Name: "tag2"}}
	arts := []*artifact.Artifact{
		{Artifact: artpkg.Artifact{Digest: "digest1"}, Tags: []*tag.Tag{tag1}},
		{Artifact: artpkg.Artifact{Digest: "digest2"}, Tags: []*tag.Tag{tag2}},
	}

	// filter required repositories haveing the specified tags
	{
		suite.artCtl.On("List", mock.Anything, mock.Anything, mock.Anything).Return(arts, nil).Once()

		candidates, err := suite.filterProcessor.ProcessTagFilter(context.TODO(), "tag2", []int64{1})
		suite.NoError(err)
		suite.Equal(1, len(candidates), "Expected 1 candidate but found ", len(candidates))
		suite.Equal("digest2", candidates[0].Digest)
		suite.Equal(int64(2), candidates[0].Tags[0].ID)
	}

	// simulate repo manager returning an error
	{
		suite.artCtl.On("List", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("test error")).Once()
		candidates, err := suite.filterProcessor.ProcessTagFilter(context.TODO(), "repo1", []int64{1})
		suite.Error(err)
		suite.Nil(candidates)
	}

}

func (suite *FilterProcessorTestSuite) TestProcessLabelFilter() {
	arts := []*artifact.Artifact{
		{Artifact: artpkg.Artifact{Digest: "digest1"}, Labels: []*labelmodel.Label{{ID: 1}}},
		{Artifact: artpkg.Artifact{Digest: "digest2"}, Labels: []*labelmodel.Label{{ID: 2}}},
	}

	// no label filter return all
	{
		candidates, err := suite.filterProcessor.ProcessLabelFilter(context.TODO(), nil, arts)
		suite.NoError(err)
		suite.Equal(2, len(candidates), "Expected 2 candidate but found ", len(candidates))
		suite.Equal("digest1", candidates[0].Digest)
		suite.Equal("digest2", candidates[1].Digest)
	}

	// filter required repositories haveing the specified label
	{
		candidates, err := suite.filterProcessor.ProcessLabelFilter(context.TODO(), []int64{2}, arts)
		suite.NoError(err)
		suite.Equal(1, len(candidates), "Expected 1 candidate but found ", len(candidates))
		suite.Equal("digest2", candidates[0].Digest)
	}
}

func TestFilterProcessorTestSuite(t *testing.T) {
	suite.Run(t, &FilterProcessorTestSuite{})
}
