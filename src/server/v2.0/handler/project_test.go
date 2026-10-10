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
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/pattern"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	apiModels "github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	repositorytesting "github.com/goharbor/harbor/src/testing/controller/repository"
	scannertesting "github.com/goharbor/harbor/src/testing/controller/scanner"
	usertesting "github.com/goharbor/harbor/src/testing/controller/user"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

func TestValidateProxyCacheRepositoryFilterUpdate(t *testing.T) {
	tests := []struct {
		name    string
		stored  map[string]string
		update  *apiModels.ProjectMetadata
		wantErr bool
	}{
		{
			name: "reject pattern-only update incompatible with stored kind",
			stored: map[string]string{
				models.ProMetaProxyCacheFilterPattern: "^foo/.*$",
				models.ProMetaProxyCacheFilterKind:    pattern.KindRegex,
			},
			update:  &apiModels.ProjectMetadata{ProxyCacheFilterPattern: stringPtr("*")},
			wantErr: true,
		},
		{
			name: "reject kind-only update incompatible with stored pattern",
			stored: map[string]string{
				models.ProMetaProxyCacheFilterPattern: "*",
				models.ProMetaProxyCacheFilterKind:    pattern.KindDoublestar,
			},
			update:  &apiModels.ProjectMetadata{ProxyCacheFilterKind: stringPtr(pattern.KindRegex)},
			wantErr: true,
		},
		{
			name: "accept compatible partial update",
			stored: map[string]string{
				models.ProMetaProxyCacheFilterPattern: "^foo/.*$",
				models.ProMetaProxyCacheFilterKind:    pattern.KindRegex,
			},
			update: &apiModels.ProjectMetadata{ProxyCacheFilterPattern: stringPtr("^bar/.*$")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProxyCacheRepositoryFilterUpdate(tt.update, tt.stored)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}

type ProjectTestSuite struct {
	htesting.Suite

	projectCtl *projecttesting.Controller
	scannerCtl *scannertesting.Controller
	project    *models.Project
	reg        *scanner.Registration
	metadata   *v1.ScannerAdapterMetadata
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
	suite.metadata = &v1.ScannerAdapterMetadata{
		Capabilities: []*v1.ScannerCapability{
			{Type: "vulnerability", ProducesMimeTypes: []string{v1.MimeTypeScanResponse}},
			{Type: "sbom", ProducesMimeTypes: []string{v1.MimeTypeSBOMReport}},
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
		mock.OnAnything(suite.scannerCtl, "GetMetadata").Return(suite.metadata, nil).Once()
		mock.OnAnything(suite.scannerCtl, "RetrieveCap").Return(nil).Once()
		res, err := suite.Get("/projects/1/scanner")
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "GetRegistrationByProject").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.scannerCtl, "GetMetadata").Return(suite.metadata, nil).Once()
		mock.OnAnything(suite.scannerCtl, "RetrieveCap").Return(nil).Once()
		var scanner scanner.Registration
		res, err := suite.GetJSON("/projects/1/scanner", &scanner)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal(suite.reg.UUID, scanner.UUID)
	}

	{
		mock.OnAnything(projectCtlMock, "GetByName").Return(suite.project, nil).Once()
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "GetMetadata").Return(suite.metadata, nil).Once()
		mock.OnAnything(suite.scannerCtl, "GetRegistrationByProject").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.scannerCtl, "RetrieveCap").Return(nil).Once()

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

		var scanners []any
		res, err := suite.GetJSON("/projects/1/scanner/candidates", &scanners)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(scanners, 0)
	}

	{
		// scanners found
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(3), nil).Once()
		mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{suite.reg}, nil).Once()

		var scanners []any
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

		res, err := suite.PutJSON("/projects/1/scanner", map[string]any{"uuid": "uuid"})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "SetRegistrationByProject").Return(nil).Once()

		res, err := suite.PutJSON("/projects/1/scanner", map[string]any{"uuid": "uuid"})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		mock.OnAnything(projectCtlMock, "GetByName").Return(suite.project, nil).Once()
		mock.OnAnything(suite.projectCtl, "Get").Return(suite.project, nil).Once()
		mock.OnAnything(suite.scannerCtl, "SetRegistrationByProject").Return(nil).Once()

		res, err := suite.PutJSON("/projects/library/scanner", map[string]any{"uuid": "uuid"})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func (suite *ProjectTestSuite) TestHeadProjectAuthorization() {
	api, ok := suite.Config.ProjectAPI.(*projectAPI)
	suite.Require().True(ok)

	originalProjectCtl := api.projectCtl
	originalBaseProjectCtl := baseProjectCtl
	originalSecurity := suite.Security
	suite.T().Cleanup(func() {
		api.projectCtl = originalProjectCtl
		baseProjectCtl = originalBaseProjectCtl
		suite.Security = originalSecurity
	})

	testCases := []struct {
		name        string
		projectName string
		project     *models.Project
		lookupErr   error
		projectID   int64
	}{
		{
			name:        "existing project",
			projectName: suite.project.Name,
			project:     suite.project,
			projectID:   suite.project.ProjectID,
		},
		{
			name:        "missing project",
			projectName: "missing",
			lookupErr:   errors.NotFoundError(nil),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			baseCtl := projecttesting.NewController(suite.T())
			handlerCtl := projecttesting.NewController(suite.T())
			securityCtx := securitytesting.NewContext(suite.T())
			baseProjectCtl = baseCtl
			api.projectCtl = handlerCtl
			suite.Security = securityCtx

			baseCtl.On("GetByName", mock.Anything, tc.projectName).
				Return(tc.project, tc.lookupErr).Twice()
			handlerCtl.On("GetByName", mock.Anything, tc.projectName).
				Return(tc.project, tc.lookupErr).Maybe()
			securityCtx.On(
				"Can",
				mock.Anything,
				rbac.ActionRead,
				rbac_project.NewNamespace(tc.projectID).Resource(),
			).Return(false).Twice()
			securityCtx.On("IsAuthenticated").Return(true).Twice()

			headResponse, err := suite.DoReq(
				http.MethodHead,
				"/projects?project_name="+tc.projectName,
				nil,
			)
			suite.Require().NoError(err)
			suite.Equal(http.StatusForbidden, headResponse.StatusCode)
			suite.NoError(headResponse.Body.Close())

			getResponse, err := suite.Get("/projects/" + tc.projectName)
			suite.Require().NoError(err)
			suite.Equal(http.StatusForbidden, getResponse.StatusCode)
			suite.NoError(getResponse.Body.Close())
		})
	}
}

func (suite *ProjectTestSuite) TestHeadProjectPublicAccess() {
	api, ok := suite.Config.ProjectAPI.(*projectAPI)
	suite.Require().True(ok)

	originalProjectCtl := api.projectCtl
	originalRepositoryCtl := api.repositoryCtl
	originalBaseProjectCtl := baseProjectCtl
	originalSecurity := suite.Security
	suite.T().Cleanup(func() {
		api.projectCtl = originalProjectCtl
		api.repositoryCtl = originalRepositoryCtl
		baseProjectCtl = originalBaseProjectCtl
		suite.Security = originalSecurity
	})

	baseCtl := projecttesting.NewController(suite.T())
	handlerCtl := projecttesting.NewController(suite.T())
	repositoryCtl := repositorytesting.NewController(suite.T())
	securityCtx := securitytesting.NewContext(suite.T())
	baseProjectCtl = baseCtl
	api.projectCtl = handlerCtl
	api.repositoryCtl = repositoryCtl
	suite.Security = securityCtx

	baseCtl.On("GetByName", mock.Anything, suite.project.Name).
		Return(suite.project, nil).Twice()
	handlerCtl.On("GetByName", mock.Anything, suite.project.Name).
		Return(suite.project, nil).Once()
	handlerCtl.On("Get", mock.Anything, suite.project.Name, mock.Anything, mock.Anything).
		Return(suite.project, nil).Once()
	repositoryCtl.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil).Once()
	securityCtx.On(
		"Can",
		mock.Anything,
		rbac.ActionRead,
		rbac_project.NewNamespace(suite.project.ProjectID).Resource(),
	).Return(true).Twice()
	securityCtx.On("IsAuthenticated").Return(false).Maybe()

	headResponse, err := suite.DoReq(
		http.MethodHead,
		"/projects?project_name="+suite.project.Name,
		nil,
	)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, headResponse.StatusCode)
	suite.NoError(headResponse.Body.Close())

	getResponse, err := suite.Get("/projects/" + suite.project.Name)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, getResponse.StatusCode)
	suite.NoError(getResponse.Body.Close())
}

func (suite *ProjectTestSuite) TestRetentionContextForSolutionUser() {
	secCtx := &securitytesting.Context{}
	secCtx.On("IsSolutionUser").Return(true).Once()
	owner := &commonmodels.User{UserID: 7, Username: "proxy-owner", SysAdminFlag: true}
	userCtl := &usertesting.Controller{}
	userCtl.On("Get", mock.Anything, 7, mock.Anything).Return(owner, nil).Once()
	api := &projectAPI{userCtl: userCtl}

	ctx, err := api.retentionContext(context.Background(), secCtx, owner.UserID)
	suite.Require().NoError(err)
	resolved, ok := security.FromContext(ctx)
	suite.Require().True(ok)
	suite.Equal(owner.Username, resolved.GetUsername())
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, &ProjectTestSuite{})
}
