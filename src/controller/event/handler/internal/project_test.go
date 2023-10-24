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

package internal

import (
	"context"
	"testing"

	beegoorm "github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/suite"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/artifact"
	immutableModel "github.com/goharbor/harbor/src/pkg/immutable/model"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/pkg/tag"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
)

// ProjectHandlerTestSuite is test suite for artifact handler.
type ProjectHandlerTestSuite struct {
	suite.Suite

	ctx     context.Context
	handler *ProjectEventHandler
}

// SetupSuite prepares for running ArtifactHandlerTestSuite.
func (suite *ProjectHandlerTestSuite) SetupSuite() {
	common_dao.PrepareTestForPostgresSQL()
	config.Init()
	suite.ctx = orm.NewContext(context.TODO(), beegoorm.NewOrm())
	suite.handler = &ProjectEventHandler{}

	// mock artifact
	_, err := pkg.ArtifactMgr.Create(suite.ctx, &artifact.Artifact{ID: 1, RepositoryID: 1})
	suite.Nil(err)
	// mock repository
	_, err = pkg.RepositoryMgr.Create(suite.ctx, &model.RepoRecord{RepositoryID: 1})
	suite.Nil(err)
	// mock tag
	_, err = tag.Mgr.Create(suite.ctx, &tagmodel.Tag{ID: 1, RepositoryID: 1, ArtifactID: 1, Name: "latest"})
	suite.Nil(err)
}

// TearDownSuite cleans environment.
func (suite *ProjectHandlerTestSuite) TearDownSuite() {
	// delete tag
	err := tag.Mgr.Delete(suite.ctx, 1)
	suite.Nil(err)
	// delete artifact
	err = pkg.ArtifactMgr.Delete(suite.ctx, 1)
	suite.Nil(err)
	// delete repository
	err = pkg.RepositoryMgr.Delete(suite.ctx, 1)
	suite.Nil(err)

}

func (suite *ProjectHandlerTestSuite) TestOnProjectDelete() {
	// create project
	projID, err := project.New().Create(suite.ctx, &models.Project{Name: "test-project", OwnerID: 1})
	suite.Nil(err)

	defer project.New().Delete(suite.ctx, projID)
	immutableRule := &immutableModel.Metadata{
		ProjectID: projID,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*immutableModel.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-**",
			},
		},
		ScopeSelectors: map[string][]*immutableModel.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "repoMatches",
					Pattern:    "redis",
				},
			},
		},
	}
	// create immutable rule
	immutableID, err := immutable.Ctr.CreateImmutableRule(suite.ctx, immutableRule)
	suite.Nil(err)

	// emit delete project event
	event := &event.DeleteProjectEvent{ProjectID: projID}
	err = suite.handler.onProjectDelete(suite.ctx, event)
	suite.Nil(err)

	// check if immutable rule is deleted
	_, err = immutable.Ctr.GetImmutableRule(suite.ctx, immutableID)
	suite.NotNil(err)
}

// TestArtifactHandler tests ArtifactHandler.
func TestProjectEventHandler(t *testing.T) {
	suite.Run(t, &ProjectHandlerTestSuite{})
}
