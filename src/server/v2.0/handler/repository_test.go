// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact/processor/cnai"
	"github.com/goharbor/harbor/src/lib/q"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
)

type RepositoryTestSuite struct {
	suite.Suite
	artCtl *artifacttesting.Controller
	api    *repositoryAPI
}

func (suite *RepositoryTestSuite) SetupTest() {
	suite.artCtl = &artifacttesting.Controller{}
	suite.api = &repositoryAPI{artCtl: suite.artCtl}
}

func (suite *RepositoryTestSuite) TestAssembleRepositoryTypeImage() {
	repo := model.NewRepoRecord(&repomodel.RepoRecord{
		RepositoryID: 1,
		Name:         "library/nginx",
		ProjectID:    1,
	})

	mock.OnAnything(suite.artCtl, "Count").Return(int64(3), nil).Once().
		Run(func(args mock.Arguments) {
			query := args.Get(1).(*q.Query)
			suite.Equal(int64(1), query.Keywords["RepositoryID"])
			suite.Nil(query.Keywords["Type"])
		})
	mock.OnAnything(suite.artCtl, "Count").Return(int64(0), nil).Once().
		Run(func(args mock.Arguments) {
			query := args.Get(1).(*q.Query)
			suite.Equal(int64(1), query.Keywords["RepositoryID"])
			suite.Equal(cnai.ArtifactTypeCNAI, query.Keywords["Type"])
		})

	got := suite.api.assembleRepository(context.Background(), repo)
	suite.Equal(int64(3), got.ArtifactCount)
	suite.Equal(repositoryTypeImage, got.Type)
	suite.artCtl.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestAssembleRepositoryTypeModel() {
	repo := model.NewRepoRecord(&repomodel.RepoRecord{
		RepositoryID: 2,
		Name:         "library/llama",
		ProjectID:    1,
	})

	mock.OnAnything(suite.artCtl, "Count").Return(int64(2), nil).Once()
	mock.OnAnything(suite.artCtl, "Count").Return(int64(1), nil).Once().
		Run(func(args mock.Arguments) {
			query := args.Get(1).(*q.Query)
			suite.Equal(cnai.ArtifactTypeCNAI, query.Keywords["Type"])
		})

	got := suite.api.assembleRepository(context.Background(), repo)
	suite.Equal(int64(2), got.ArtifactCount)
	suite.Equal(repositoryTypeModel, got.Type)
	suite.artCtl.AssertExpectations(suite.T())
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
