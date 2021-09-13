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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	scantesting "github.com/goharbor/harbor/src/testing/controller/scan"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/suite"
)

type ScanTestSuite struct {
	htesting.Suite

	artifactCtl *artifacttesting.Controller
	scanCtl     *scantesting.Controller

	execution      *task.Execution
	projectCtlMock *projecttesting.Controller
}

func (suite *ScanTestSuite) SetupSuite() {
	suite.execution = &task.Execution{
		Status: "Running",
	}

	suite.scanCtl = &scantesting.Controller{}
	suite.artifactCtl = &artifacttesting.Controller{}

	suite.Config = &restapi.Config{
		ScanAPI: &scanAPI{
			artCtl:  suite.artifactCtl,
			scanCtl: suite.scanCtl,
		},
	}

	suite.Suite.SetupSuite()

	mock.OnAnything(projectCtlMock, "GetByName").Return(&project.Project{ProjectID: 1}, nil)
}

func (suite *ScanTestSuite) TestStopScan() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	url := "/projects/library/repositories/nginx/artifacts/sha256:e4f0474a75c510f40b37b6b7dc2516241ffa8bde5a442bde3d372c9519c84d90/scan/stop"

	{
		// failed to get artifact by reference
		mock.OnAnything(suite.artifactCtl, "GetByReference").Return(&artifact.Artifact{}, fmt.Errorf("failed to get artifact by reference")).Once()

		res, err := suite.Post(url, nil)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// get nil artifact by reference
		mock.OnAnything(suite.artifactCtl, "GetByReference").Return(nil, nil).Once()
		mock.OnAnything(suite.scanCtl, "Stop").Return(fmt.Errorf("nil artifact to stop scan")).Once()

		res, err := suite.Post(url, nil)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// successfully stop scan artifact
		mock.OnAnything(suite.artifactCtl, "GetByReference").Return(&artifact.Artifact{}, nil).Once()
		mock.OnAnything(suite.scanCtl, "Stop").Return(nil).Once()

		res, err := suite.Post(url, nil)
		suite.NoError(err)
		suite.Equal(202, res.StatusCode)
	}
}

func TestScanTestSuite(t *testing.T) {
	suite.Run(t, &ScanTestSuite{})
}
