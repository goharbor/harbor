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
	"time"

	beegoorm "github.com/beego/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/artifact"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/pkg/tag"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/stretchr/testify/suite"
)

// ArtifactHandlerTestSuite is test suite for artifact handler.
type ArtifactHandlerTestSuite struct {
	suite.Suite

	ctx     context.Context
	handler *Handler
}

// TestArtifactHandler tests ArtifactHandler.
func TestArtifactHandler(t *testing.T) {
	suite.Run(t, &ArtifactHandlerTestSuite{})
}

// SetupSuite prepares for running ArtifactHandlerTestSuite.
func (suite *ArtifactHandlerTestSuite) SetupSuite() {
	common_dao.PrepareTestForPostgresSQL()
	config.Init()
	suite.handler = &Handler{}
	suite.ctx = orm.NewContext(context.TODO(), beegoorm.NewOrm())

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
func (suite *ArtifactHandlerTestSuite) TearDownSuite() {
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

// TestName tests method Name.
func (suite *ArtifactHandlerTestSuite) TestName() {
	suite.Equal("InternalArtifact", suite.handler.Name())
}

// TestIsStateful tests method IsStateful.
func (suite *ArtifactHandlerTestSuite) TestIsStateful() {
	suite.False(suite.handler.IsStateful(), "artifact handler is not stateful")
}

// TestDefaultAsyncDuration tests default value of async flush duration.
func (suite *ArtifactHandlerTestSuite) TestDefaultAsyncFlushDuration() {
	suite.Equal(defaultAsyncFlushDuration, asyncFlushDuration, "default async flush duration")
}

// TestOnPush tests handle push events.
func (suite *ArtifactHandlerTestSuite) TestOnPush() {
	err := suite.handler.onPush(context.TODO(), &event.ArtifactEvent{Artifact: &artifact.Artifact{}})
	suite.Nil(err, "onPush should return nil")
}

// TestOnPull tests handler pull events.
func (suite *ArtifactHandlerTestSuite) TestOnPull() {
	// test sync mode
	asyncFlushDuration = 0
	err := suite.handler.onPull(suite.ctx, &event.ArtifactEvent{Artifact: &artifact.Artifact{ID: 1, RepositoryID: 1}, Tags: []string{"latest"}})
	suite.Nil(err, "onPull should return nil")
	// sync mode should update db immediately
	// pull_time
	art, err := pkg.ArtifactMgr.Get(suite.ctx, 1)
	suite.Nil(err)
	suite.False(art.PullTime.IsZero(), "sync update pull_time")
	lastPullTime := art.PullTime
	// pull_count
	repository, err := pkg.RepositoryMgr.Get(suite.ctx, 1)
	suite.Nil(err)
	suite.Equal(int64(1), repository.PullCount, "sync update pull_count")

	// test async mode
	asyncFlushDuration = 200 * time.Millisecond
	err = suite.handler.onPull(suite.ctx, &event.ArtifactEvent{Artifact: &artifact.Artifact{ID: 1, RepositoryID: 1}, Tags: []string{"latest"}})
	suite.Nil(err, "onPull should return nil")
	// async mode should not update db immediately
	// pull_time
	art, err = pkg.ArtifactMgr.Get(suite.ctx, 1)
	suite.Nil(err)
	suite.Equal(lastPullTime, art.PullTime, "pull_time should not be updated immediately")
	// pull_count
	repository, err = pkg.RepositoryMgr.Get(suite.ctx, 1)
	suite.Nil(err)
	suite.Equal(int64(1), repository.PullCount, "pull_count should not be updated immediately")
	// wait for db update
	suite.Eventually(func() bool {
		art, err = pkg.ArtifactMgr.Get(suite.ctx, 1)
		suite.Nil(err)
		return art.PullTime.After(lastPullTime)
	}, 3*asyncFlushDuration, asyncFlushDuration/2, "wait for pull_time async update")

	suite.Eventually(func() bool {
		repository, err = pkg.RepositoryMgr.Get(suite.ctx, 1)
		suite.Nil(err)
		return int64(2) == repository.PullCount
	}, 3*asyncFlushDuration, asyncFlushDuration/2, "wait for pull_count async update")
}
