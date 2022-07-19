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

package vulnerable

import (
	"fmt"
	"github.com/goharbor/harbor/src/pkg/accessory"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	scantesting "github.com/goharbor/harbor/src/testing/controller/scan"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type MiddlewareTestSuite struct {
	suite.Suite

	originalArtifactController artifact.Controller
	artifactController         *artifacttesting.Controller

	originalProjectController project.Controller
	projectController         *projecttesting.Controller

	originalScanController scan.Controller
	scanController         *scantesting.Controller

	originalAccessMgr accessory.Manager
	accessMgr         *accessorytesting.Manager

	checker     *scantesting.Checker
	scanChecker func() scan.Checker

	artifact *artifact.Artifact
	project  *proModels.Project

	next http.Handler
}

func (suite *MiddlewareTestSuite) SetupTest() {
	suite.originalArtifactController = artifactController
	suite.artifactController = &artifacttesting.Controller{}
	artifactController = suite.artifactController

	suite.originalProjectController = projectController
	suite.projectController = &projecttesting.Controller{}
	projectController = suite.projectController

	suite.originalAccessMgr = accessory.Mgr
	suite.accessMgr = &accessorytesting.Manager{}
	accessory.Mgr = suite.accessMgr

	suite.originalScanController = scanController
	suite.scanController = &scantesting.Controller{}
	scanController = suite.scanController

	suite.checker = &scantesting.Checker{}
	suite.scanChecker = scanChecker

	scanChecker = func() scan.Checker {
		return suite.checker
	}

	suite.artifact = &artifact.Artifact{}
	suite.artifact.Type = image.ArtifactTypeImage
	suite.artifact.ProjectID = 1
	suite.artifact.RepositoryName = "library/photon"
	suite.artifact.Digest = "digest"

	suite.project = &proModels.Project{
		ProjectID: suite.artifact.ProjectID,
		Name:      "library",
		Metadata: map[string]string{
			proModels.ProMetaPreventVul: "true",
			proModels.ProMetaSeverity:   vuln.High.String(),
		},
	}

	suite.next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (suite *MiddlewareTestSuite) TearDownTest() {
	artifactController = suite.originalArtifactController
	projectController = suite.originalProjectController
	scanController = suite.originalScanController
	accessory.Mgr = suite.originalAccessMgr
	scanChecker = suite.scanChecker
}

func (suite *MiddlewareTestSuite) makeRequest() *http.Request {
	req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil)

	info := lib.ArtifactInfo{
		Repository: "library/photon",
		Reference:  "2.0",
		Tag:        "2.0",
		Digest:     "",
	}

	return req.WithContext(lib.WithArtifactInfo(req.Context(), info))
}

func (suite *MiddlewareTestSuite) TestNoArtifactInfo() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(nil, fmt.Errorf("error"))

	req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil)
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusNotFound)
}

func (suite *MiddlewareTestSuite) TestGetArtifactFailed() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(nil, fmt.Errorf("error"))
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusInternalServerError)
}

func (suite *MiddlewareTestSuite) TestGetProjectFailed() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(nil, fmt.Errorf("err"))

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusInternalServerError)
}

func (suite *MiddlewareTestSuite) TestPreventionDisabled() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	suite.project.Metadata[proModels.ProMetaPreventVul] = "false"
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func (suite *MiddlewareTestSuite) TestNonScannerPulling() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	securityCtx := &securitytesting.Context{}
	mock.OnAnything(securityCtx, "Name").Return("local")
	mock.OnAnything(securityCtx, "Can").Return(true, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(false, nil)

	req := suite.makeRequest()
	req = req.WithContext(security.NewContext(req.Context(), securityCtx))
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func (suite *MiddlewareTestSuite) TestScannerPulling() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	securityCtx := &securitytesting.Context{}
	mock.OnAnything(securityCtx, "Name").Return("v2token")
	mock.OnAnything(securityCtx, "Can").Return(true, nil)

	req := suite.makeRequest()
	req = req.WithContext(security.NewContext(req.Context(), securityCtx))
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func (suite *MiddlewareTestSuite) TestCheckIsScannableFailed() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(false, fmt.Errorf("error"))
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusInternalServerError)
}

func (suite *MiddlewareTestSuite) TestArtifactIsNotScannable() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(false, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func (suite *MiddlewareTestSuite) TestArtifactNotScanned() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(true, nil)
	mock.OnAnything(suite.scanController, "GetVulnerable").Return(nil, errors.NotFoundError(nil))
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusPreconditionFailed)
}

func (suite *MiddlewareTestSuite) TestArtifactScanFailed() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(true, nil)
	mock.OnAnything(suite.scanController, "GetVulnerable").Return(&scan.Vulnerable{ScanStatus: "Error"}, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusPreconditionFailed)
}

func (suite *MiddlewareTestSuite) TestGetVulnerableFailed() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(true, nil)
	mock.OnAnything(suite.scanController, "GetVulnerable").Return(nil, fmt.Errorf("error"))
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusInternalServerError)
}

func (suite *MiddlewareTestSuite) TestNoVulnerabilities() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(true, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	mock.OnAnything(suite.scanController, "GetVulnerable").Return(&scan.Vulnerable{
		ScanStatus:  "Success",
		CVEBypassed: []string{"cve-2020"},
	}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func (suite *MiddlewareTestSuite) TestAllowed() {
	low := vuln.Low
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(true, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	mock.OnAnything(suite.scanController, "GetVulnerable").Return(&scan.Vulnerable{
		ScanStatus:           "Success",
		Severity:             &low,
		VulnerabilitiesCount: 1,
		CVEBypassed:          []string{"cve-2020"},
	}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func (suite *MiddlewareTestSuite) TestPrevented() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(true, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)

	critical := vuln.Critical

	{
		// only one vulnerability
		mock.OnAnything(suite.scanController, "GetVulnerable").Return(&scan.Vulnerable{
			ScanStatus:           "Success",
			Severity:             &critical,
			VulnerabilitiesCount: 1,
		}, nil).Once()

		req := suite.makeRequest()
		rr := httptest.NewRecorder()

		Middleware()(suite.next).ServeHTTP(rr, req)
		suite.Equal(rr.Code, http.StatusPreconditionFailed)

		suite.Contains(rr.Body.String(), "current image with 1 vulnerability cannot be pulled")
	}

	{
		// multiple vulnerabilities
		mock.OnAnything(suite.scanController, "GetVulnerable").Return(&scan.Vulnerable{
			ScanStatus:           "Success",
			Severity:             &critical,
			VulnerabilitiesCount: 2,
		}, nil).Once()

		req := suite.makeRequest()
		rr := httptest.NewRecorder()

		Middleware()(suite.next).ServeHTTP(rr, req)
		suite.Equal(rr.Code, http.StatusPreconditionFailed)

		suite.Contains(rr.Body.String(), "current image with 2 vulnerabilities cannot be pulled")
	}
}

func (suite *MiddlewareTestSuite) TestArtifactIsImageIndex() {
	critical := vuln.Critical

	suite.artifact.ManifestMediaType = manifestlist.MediaTypeManifestList
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	mock.OnAnything(suite.checker, "IsScannable").Return(true, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	mock.OnAnything(suite.scanController, "GetVulnerable").Return(&scan.Vulnerable{
		ScanStatus: "Success",
		Severity:   &critical,
	}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

// pull cosign signature when policy checker is enabled.
func (suite *MiddlewareTestSuite) TestSignaturePulling() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "Get").Return(suite.project, nil)
	acc := &basemodel.Default{
		Data: accessorymodel.AccessoryData{
			ID:            1,
			ArtifactID:    2,
			SubArtifactID: 1,
			Type:          accessorymodel.TypeCosignSignature,
		},
	}
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{
		acc,
	}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Middleware()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &MiddlewareTestSuite{})
}
