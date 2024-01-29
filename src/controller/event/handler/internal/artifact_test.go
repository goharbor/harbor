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

	beegoorm "github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/suite"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/artifact"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/pkg/tag"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	scannerCtlMock "github.com/goharbor/harbor/src/testing/controller/scanner"
	"github.com/goharbor/harbor/src/testing/mock"
	artMock "github.com/goharbor/harbor/src/testing/pkg/artifact"
	projectMock "github.com/goharbor/harbor/src/testing/pkg/project"
	reportMock "github.com/goharbor/harbor/src/testing/pkg/scan/report"
	taskMock "github.com/goharbor/harbor/src/testing/pkg/task"
)

// ArtifactHandlerTestSuite is test suite for artifact handler.
type ArtifactHandlerTestSuite struct {
	suite.Suite

	ctx            context.Context
	handler        *ArtifactEventHandler
	projectManager project.Manager
	scannerCtl     scanner.Controller
	reportMgr      *reportMock.Manager
	execMgr        *taskMock.ExecutionManager
	artMgr         *artMock.Manager
}

// TestArtifactHandler tests ArtifactHandler.
func TestArtifactHandler(t *testing.T) {
	suite.Run(t, &ArtifactHandlerTestSuite{})
}

// SetupSuite prepares for running ArtifactHandlerTestSuite.
func (suite *ArtifactHandlerTestSuite) SetupSuite() {
	common_dao.PrepareTestForPostgresSQL()
	config.Init()
	suite.ctx = orm.NewContext(context.TODO(), beegoorm.NewOrm())
	suite.projectManager = &projectMock.Manager{}
	suite.scannerCtl = &scannerCtlMock.Controller{}
	suite.execMgr = &taskMock.ExecutionManager{}
	suite.reportMgr = &reportMock.Manager{}
	suite.artMgr = &artMock.Manager{}
	suite.handler = &ArtifactEventHandler{execMgr: suite.execMgr, reportMgr: suite.reportMgr, artMgr: suite.artMgr}

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

func (suite *ArtifactHandlerTestSuite) TestOnDelete() {
	evt := &event.ArtifactEvent{Artifact: &artifact.Artifact{ID: 1, RepositoryID: 1, Digest: "mock-digest", References: []*artifact.Reference{{ChildDigest: "ref-1", ChildID: 2}, {ChildDigest: "ref-2", ChildID: 3}}}}
	suite.execMgr.On("DeleteByVendor", suite.ctx, "IMAGE_SCAN", int64(1)).Return(nil).Times(1)
	suite.execMgr.On("DeleteByVendor", suite.ctx, "IMAGE_SCAN", int64(2)).Return(nil).Times(1)
	suite.execMgr.On("DeleteByVendor", suite.ctx, "IMAGE_SCAN", int64(3)).Return(nil).Times(1)
	suite.artMgr.On("Count", suite.ctx, mock.Anything).Return(int64(0), nil).Times(3)
	suite.reportMgr.On("DeleteByDigests", suite.ctx, "mock-digest", "ref-1", "ref-2").Return(nil).Times(1)
	err := suite.handler.onDelete(suite.ctx, evt)
	suite.Nil(err, "onDelete should return nil")
}

func (suite *ArtifactHandlerTestSuite) TestIsScannerUser() {
	type args struct {
		prefix string
		event  *event.ArtifactEvent
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"normal_true", args{"robot$", &event.ArtifactEvent{Operator: "robot$library+scanner+Trivy-2e6240a1-f3be-11ec-8fba-0242ac1e0009", Repository: "library/nginx"}}, true},
		{"no_scanner_prefix_false", args{"robot$", &event.ArtifactEvent{Operator: "robot$library+Trivy-2e6240a1-f3be-11ec-8fba-0242ac1e0009", Repository: "library/nginx"}}, false},
		{"operator_empty", args{"robot$", &event.ArtifactEvent{Operator: "", Repository: "library/nginx"}}, false},
		{"normal_user", args{"robot$", &event.ArtifactEvent{Operator: "Trivy_sample", Repository: "library/nginx"}}, false},
		{"normal_user_with_robotname", args{"robot$", &event.ArtifactEvent{Operator: "robot_Trivy", Repository: "library/nginx"}}, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if got := isScannerUser(suite.ctx, tt.args.event); got != tt.want {
				suite.Errorf(nil, "isScannerUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseProjectName(t *testing.T) {
	type args struct {
		repoName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal repo name", args{"library/nginx"}, "library"},
		{"three levels of repository", args{"library/nginx/nginx"}, "library"},
		{"repo name without project name", args{"nginx"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseProjectName(tt.args.repoName); got != tt.want {
				t.Errorf("parseProjectName() = %v, want %v", got, tt.want)
			}
		})
	}
}
