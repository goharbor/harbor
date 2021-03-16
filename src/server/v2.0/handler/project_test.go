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

	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	scannertesting "github.com/goharbor/harbor/src/testing/controller/scanner"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	htesting.Suite

	projectCtl *projecttesting.Controller
	scannerCtl *scannertesting.Controller
	project    *models.Project
	reg        *scanner.Registration
}

func (suite *ProjectTestSuite) SetupSuite() {
	suite.project = &models.Project{
		ProjectID: 1,
		Name:      "library",
	}

	suite.reg = &scanner.Registration{
		Name: "reg",
		URL:  "http://reg:8080",
		UUID: "uuid",
	}

	suite.projectCtl = &projecttesting.Controller{}
	suite.scannerCtl = &scannertesting.Controller{}

	suite.Config = &restapi.Config{
		ProjectAPI: &projectAPI{
			projectCtl: suite.projectCtl,
			scannerCtl: suite.scannerCtl,
		},
	}

	suite.Suite.SetupSuite()
}

func (suite *ProjectTestSuite) TestGetScannerOfProject() {
	times := 4
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get project failed
		mock.OnAnything(suite.projectCtl, "Get").Return(nil, fmt.Errorf("failed to get project")).Once()

		res, err := suite.Get("/projects/1/scanner")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// scanner not found
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "GetRegistrationByProject").Return(nil, nil).Once()

		res, err := suite.Get("/projects/1/scanner")
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "GetRegistrationByProject").Return(suite.reg, nil).Once()

		var scanner scanner.Registration
		res, err := suite.GetJSON("/projects/1/scanner", &scanner)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal(suite.reg.UUID, scanner.UUID)
	}

	{
		mock.OnAnything(projectCtlMock, "GetByName").Return(suite.project, nil).Once()
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "GetRegistrationByProject").Return(suite.reg, nil).Once()

		var scanner scanner.Registration
		res, err := suite.GetJSON("/projects/library/scanner", &scanner)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal(suite.reg.UUID, scanner.UUID)
	}
}

func (suite *ProjectTestSuite) TestListScannerCandidatesOfProject() {
	times := 4
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// list scanners failed
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(0), fmt.Errorf("failed to count scanners")).Once()

		res, err := suite.Get("/projects/1/scanner/candidates")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// list scanners failed
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(1), nil).Once()
		mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return(nil, fmt.Errorf("failed to list scanners")).Once()

		res, err := suite.Get("/projects/1/scanner/candidates")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// scanners not found
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(0), nil).Once()
		mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return(nil, nil).Once()

		var scanners []interface{}
		res, err := suite.GetJSON("/projects/1/scanner/candidates", &scanners)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(scanners, 0)
	}

	{
		// scanners found
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(3), nil).Once()
		mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{suite.reg}, nil).Once()

		var scanners []interface{}
		res, err := suite.GetJSON("/projects/1/scanner/candidates?page_size=1&page=2&name=n&description=d&url=u&ex_name=n&ex_url=u", &scanners)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(scanners, 1)
		suite.Equal("3", res.Header.Get("X-Total-Count"))
		suite.Contains(res.Header, "Link")
		suite.Equal(`</api/v2.0/projects/1/scanner/candidates?description=d&ex_name=n&ex_url=u&name=n&page=1&page_size=1&url=u>; rel="prev" , </api/v2.0/projects/1/scanner/candidates?description=d&ex_name=n&ex_url=u&name=n&page=3&page_size=1&url=u>; rel="next"`, res.Header.Get("Link"))
	}
}

func (suite *ProjectTestSuite) TestSetScannerOfProject() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get project failed
		mock.OnAnything(suite.projectCtl, "Get").Return(nil, fmt.Errorf("failed to get project")).Once()

		res, err := suite.PutJSON("/projects/1/scanner", map[string]interface{}{"uuid": "uuid"})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "SetRegistrationByProject").Return(nil).Once()

		res, err := suite.PutJSON("/projects/1/scanner", map[string]interface{}{"uuid": "uuid"})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		mock.OnAnything(projectCtlMock, "GetByName").Return(suite.project, nil).Once()
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "SetRegistrationByProject").Return(nil).Once()

		res, err := suite.PutJSON("/projects/library/scanner", map[string]interface{}{"uuid": "uuid"})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, &ProjectTestSuite{})
}
