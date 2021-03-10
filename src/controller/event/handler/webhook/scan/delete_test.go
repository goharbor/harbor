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

package scan

import (
	"context"
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib/q"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	scantesting "github.com/goharbor/harbor/src/testing/controller/scan"
	"github.com/stretchr/testify/suite"
)

type DelArtHandlerTestSuite struct {
	suite.Suite

	artifactCtl         *artifacttesting.Controller
	originalArtifactCtl artifact.Controller

	scanCtl         *scantesting.Controller
	originalScanCtl scan.Controller
}

func (suite *DelArtHandlerTestSuite) SetupSuite() {
	suite.artifactCtl = &artifacttesting.Controller{}
	suite.originalArtifactCtl = artifact.Ctl
	artifact.Ctl = suite.artifactCtl

	suite.scanCtl = &scantesting.Controller{}
	suite.originalScanCtl = scan.DefaultController
	scan.DefaultController = suite.scanCtl
}

func (suite *DelArtHandlerTestSuite) TeardownSuite() {
	artifact.Ctl = suite.originalArtifactCtl
	scan.DefaultController = suite.originalScanCtl
}

func (suite *DelArtHandlerTestSuite) TestHandle() {
	o := DelArtHandler{}

	suite.Error(o.Handle(context.TODO(), nil))

	suite.Error(o.Handle(context.TODO(), "string"))

	art := &artifact.Artifact{}
	art.Digest = "digest"
	ev := &event.ArtifactEvent{Artifact: &art.Artifact}

	value := &event.DeleteArtifactEvent{ArtifactEvent: ev}

	suite.artifactCtl.On("Count", context.TODO(), q.New(q.KeyWords{"digest": "digest"})).Return(int64(0), fmt.Errorf("failed")).Once()
	suite.Require().NoError(o.Handle(context.TODO(), value))

	suite.artifactCtl.On("Count", context.TODO(), q.New(q.KeyWords{"digest": "digest"})).Return(int64(1), nil).Once()
	suite.Require().NoError(o.Handle(context.TODO(), value))

	suite.artifactCtl.On("Count", context.TODO(), q.New(q.KeyWords{"digest": "digest"})).Return(int64(0), nil).Once()
	suite.scanCtl.On("DeleteReports", context.TODO(), "digest").Return(fmt.Errorf("failed")).Once()
	suite.Require().Error(o.Handle(context.TODO(), value))

	suite.artifactCtl.On("Count", context.TODO(), q.New(q.KeyWords{"digest": "digest"})).Return(int64(0), nil).Once()
	suite.scanCtl.On("DeleteReports", context.TODO(), "digest").Return(nil).Once()
	suite.Require().NoError(o.Handle(context.TODO(), value))
}

func TestDelArtHandlerTestSuite(t *testing.T) {
	suite.Run(t, &DelArtHandlerTestSuite{})
}
