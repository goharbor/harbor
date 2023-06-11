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

package contenttrust

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/accessory"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
)

type CosignMiddlewareTestSuite struct {
	suite.Suite

	originalArtifactController artifact.Controller
	artifactController         *artifacttesting.Controller

	originalProjectController project.Controller
	projectController         *projecttesting.Controller

	originalAccessMgr accessory.Manager
	accessMgr         *accessorytesting.Manager

	originalCosignCAPath  string
	originalCosignKeyPath string

	artifact       *artifact.Artifact
	signedArtifact *artifact.Artifact
	project        *proModels.Project

	next http.Handler

	registry *httptest.Server
}

func (suite *CosignMiddlewareTestSuite) SetupTest() {
	suite.originalArtifactController = artifact.Ctl
	suite.artifactController = &artifacttesting.Controller{}
	artifact.Ctl = suite.artifactController

	suite.originalProjectController = project.Ctl
	suite.projectController = &projecttesting.Controller{}
	project.Ctl = suite.projectController

	suite.originalAccessMgr = accessory.Mgr
	suite.accessMgr = &accessorytesting.Manager{}
	accessory.Mgr = suite.accessMgr

	suite.artifact = &artifact.Artifact{}
	suite.artifact.Type = image.ArtifactTypeImage
	suite.artifact.ProjectID = 1
	suite.artifact.RepositoryName = "library/photon"
	suite.artifact.Digest = "digest"

	suite.signedArtifact = &artifact.Artifact{}
	suite.signedArtifact.Type = image.ArtifactTypeImage
	suite.signedArtifact.ProjectID = 1
	suite.signedArtifact.RepositoryName = "library/hello-world"
	suite.signedArtifact.Digest = "sha256:ad7c8b818cf3016b2b64372e166b339a8e8f0f722a812a867ae20d9a46171dff"
	suite.signedArtifact.Accessories = []accessorymodel.Accessory{
		&basemodel.Default{
			Data: accessorymodel.AccessoryData{
				ID:            1,
				ArtifactID:    1,
				SubArtifactID: 2,
				Type:          accessorymodel.TypeCosignSignature,
				Digest:        "sig",
			},
		},
	}

	suite.project = &proModels.Project{
		ProjectID: suite.artifact.ProjectID,
		Name:      "library",
		Metadata: map[string]string{
			proModels.ProMetaEnableContentTrustCosign: "true",
		},
	}

	suite.next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Set tmp dir to cosign CA and public key for testing, and restore original values.
	suite.originalCosignCAPath = cosignCAPath
	cosignCAPath = fmt.Sprintf("%s/ca.crt", os.TempDir())
	suite.originalCosignKeyPath = cosignKeyPath
	cosignKeyPath = fmt.Sprintf("%s/key.pub", os.TempDir())

	// Set registry url env var.
	suite.registry = httptest.NewServer(registry.New())
	os.Setenv("REGISTRY_URL", suite.registry.URL)
}

func (suite *CosignMiddlewareTestSuite) TearDownTest() {
	artifact.Ctl = suite.originalArtifactController
	project.Ctl = suite.originalProjectController
	accessory.Mgr = suite.originalAccessMgr

	// Clean the created cosign ca file.
	os.RemoveAll(cosignCAPath)
	cosignCAPath = suite.originalCosignCAPath

	// Clean the created cosign key file.
	os.RemoveAll(cosignKeyPath)
	cosignKeyPath = suite.originalCosignKeyPath

	// Tear down registry server.
	suite.registry.Close()

	// Unset env varibles.
	os.Unsetenv("REGISTRY_URL")
	os.Unsetenv(cosignStrictVerificationEnabled)
}

func (suite *CosignMiddlewareTestSuite) makeRequest(setHeader ...bool) *http.Request {
	req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil)
	info := lib.ArtifactInfo{
		Repository: "library/photon",
		Reference:  "2.0",
		Tag:        "2.0",
		Digest:     "",
	}
	if len(setHeader) > 0 {
		if setHeader[0] {
			req.Header.Add("User-Agent", "cosign test")
		}
	}
	return req.WithContext(lib.WithArtifactInfo(req.Context(), info))
}

func (suite *CosignMiddlewareTestSuite) TestGetArtifactFailed() {
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.artifactController, "GetByReference").Return(nil, fmt.Errorf("error"))

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusInternalServerError)
}

func (suite *CosignMiddlewareTestSuite) TestGetProjectFailed() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(nil, fmt.Errorf("err"))

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusInternalServerError)
}

func (suite *CosignMiddlewareTestSuite) TestContentTrustDisabled() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	suite.project.Metadata[proModels.ProMetaEnableContentTrustCosign] = "false"
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func (suite *CosignMiddlewareTestSuite) TestNoneArtifact() {
	req := httptest.NewRequest("GET", "/v1/library/photon/manifests/nonexist", nil)
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusNotFound)
}

func (suite *CosignMiddlewareTestSuite) TestAuthenticatedUserPulling() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	securityCtx := &securitytesting.Context{}
	mock.OnAnything(securityCtx, "Name").Return("local")
	mock.OnAnything(securityCtx, "Can").Return(true, nil)
	mock.OnAnything(securityCtx, "IsAuthenticated").Return(true)

	req := suite.makeRequest()
	req = req.WithContext(security.NewContext(req.Context(), securityCtx))
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusPreconditionFailed)
}

func (suite *CosignMiddlewareTestSuite) TestScannerPulling() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	securityCtx := &securitytesting.Context{}
	mock.OnAnything(securityCtx, "Name").Return("v2token")
	mock.OnAnything(securityCtx, "Can").Return(true, nil)
	mock.OnAnything(securityCtx, "IsAuthenticated").Return(true)

	req := suite.makeRequest()
	req = req.WithContext(security.NewContext(req.Context(), securityCtx))
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

func (suite *CosignMiddlewareTestSuite) TestCosignPulling() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	securityCtx := &securitytesting.Context{}
	mock.OnAnything(securityCtx, "Name").Return("v2token")
	mock.OnAnything(securityCtx, "Can").Return(true, nil)
	mock.OnAnything(securityCtx, "IsAuthenticated").Return(true)

	req := suite.makeRequest(true)
	req = req.WithContext(security.NewContext(req.Context(), securityCtx))
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

// pull a public project a un-signed image when policy checker is enabled.
func (suite *CosignMiddlewareTestSuite) TestUnAuthenticatedUserPulling() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	securityCtx := &securitytesting.Context{}
	mock.OnAnything(securityCtx, "Name").Return("local")
	mock.OnAnything(securityCtx, "Can").Return(true, nil)
	mock.OnAnything(securityCtx, "IsAuthenticated").Return(false)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusPreconditionFailed)
}

// pull cosign signature when policy checker is enabled.
func (suite *CosignMiddlewareTestSuite) TestSignaturePulling() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.artifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	acc := &basemodel.Default{
		Data: accessorymodel.AccessoryData{
			ID:                1,
			ArtifactID:        2,
			SubArtifactDigest: suite.artifact.Digest,
			Type:              accessorymodel.TypeCosignSignature,
		},
	}
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{
		acc,
	}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()

	Cosign()(suite.next).ServeHTTP(rr, req)
	suite.Equal(rr.Code, http.StatusOK)
}

// pull artifact with strict verification not set.
func (suite *CosignMiddlewareTestSuite) TestStrictSigVerificationNotSet() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.signedArtifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)

	req := suite.makeRequest()
	rr := httptest.NewRecorder()
	Cosign()(suite.next).ServeHTTP(rr, req)

	suite.Equal(rr.Code, http.StatusOK)
}

// pull artifact with strict verification disabled.
func (suite *CosignMiddlewareTestSuite) TestStrictSigVerificationDisabled() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.signedArtifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	os.Setenv(cosignStrictVerificationEnabled, "false")

	req := suite.makeRequest()
	rr := httptest.NewRecorder()
	Cosign()(suite.next).ServeHTTP(rr, req)

	suite.Equal(rr.Code, http.StatusOK)
}

// pull artifact with strict signature verification enabled and pass verification by CA cert.
func (suite *CosignMiddlewareTestSuite) TestStrictSigVerificationByCertPass() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.signedArtifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	// Read test image index and its signature.
	host := strings.TrimPrefix(suite.registry.URL, "http://")
	// Push to the regsitry the hello-world image.
	suite.Require().Nil(writeImageIndex("hello-world", fmt.Sprintf("%s/library/hello-world:latest", host)))
	// Push to the regsitry the signature of hello-world image that includes cert for verification.
	suite.Require().Nil(writeImageIndex("hello-world-sig-with-cert", fmt.Sprintf("%s/library/hello-world:sha256-ad7c8b818cf3016b2b64372e166b339a8e8f0f722a812a867ae20d9a46171dff.sig", host)))
	// Write CA cert to the cosignCAPath.
	suite.Require().Nil(writeFileToDstPath("test.crt", cosignCAPath))
	os.Setenv(cosignStrictVerificationEnabled, "true")

	req := suite.makeRequest()
	rr := httptest.NewRecorder()
	Cosign()(suite.next).ServeHTTP(rr, req)

	suite.Equal(rr.Code, http.StatusOK)
}

// pull artifact with strict signature verification enabled and pass verification with public key.
func (suite *CosignMiddlewareTestSuite) TestStrictSigVerificationByKeyPass() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.signedArtifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	// Read test image index and its signature.
	host := strings.TrimPrefix(suite.registry.URL, "http://")
	// Push to the regsitry the hello-world image.
	suite.Require().Nil(writeImageIndex("hello-world", fmt.Sprintf("%s/library/hello-world:latest", host)))
	// Push to the regsitry the signature of hello-world image.
	suite.Require().Nil(writeImageIndex("hello-world-sig", fmt.Sprintf("%s/library/hello-world:sha256-ad7c8b818cf3016b2b64372e166b339a8e8f0f722a812a867ae20d9a46171dff.sig", host)))
	// Write public key to the cosignKeyDir.
	suite.Require().Nil(writeFileToDstPath("test.pub", cosignKeyPath))
	os.Setenv(cosignStrictVerificationEnabled, "true")

	req := suite.makeRequest()
	rr := httptest.NewRecorder()
	Cosign()(suite.next).ServeHTTP(rr, req)

	suite.Equal(rr.Code, http.StatusOK)
}

// pull artifact with signature verification fail due to signature signed by untrusted key.
func (suite *CosignMiddlewareTestSuite) TestStrictSigVerificationFailWithSignedByUntrustedKey() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.signedArtifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)

	// Read test image index, but with a signature signed by untrusted key.
	host := strings.TrimPrefix(suite.registry.URL, "http://")
	// Push to the regsitry the hello-world image.
	suite.Require().Nil(writeImageIndex("hello-world", fmt.Sprintf("%s/library/hello-world:latest", host)))
	// Push to the regsitry the signature of hello-world image that is signed by untrusted key.
	suite.Require().Nil(writeImageIndex("fake-hello-world-sig", fmt.Sprintf("%s/library/hello-world:sha256-ad7c8b818cf3016b2b64372e166b339a8e8f0f722a812a867ae20d9a46171dff.sig", host)))
	// Write public key to the cosignKeyPath.
	suite.Require().Nil(writeFileToDstPath("test.pub", cosignKeyPath))
	os.Setenv(cosignStrictVerificationEnabled, "true")

	req := suite.makeRequest()
	rr := httptest.NewRecorder()
	Cosign()(suite.next).ServeHTTP(rr, req)

	suite.Equal(rr.Code, http.StatusPreconditionFailed)
}

// pull artifact with signature verification fail due to missing signature.
func (suite *CosignMiddlewareTestSuite) TestStrictSigVerificationFailWithMissingSig() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.signedArtifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	// Read test image index, but no signature.
	host := strings.TrimPrefix(suite.registry.URL, "http://")
	// Push to the regsitry the hello-world image.
	suite.Require().Nil(writeImageIndex("hello-world", fmt.Sprintf("%s/library/hello-world:latest", host)))
	// Write CA cert to the cosignCAPath.
	suite.Require().Nil(writeFileToDstPath("test.crt", cosignCAPath))
	os.Setenv(cosignStrictVerificationEnabled, "true")

	req := suite.makeRequest()
	rr := httptest.NewRecorder()
	Cosign()(suite.next).ServeHTTP(rr, req)

	suite.Equal(rr.Code, http.StatusPreconditionFailed)
}

// pull artifact with signature verification fail due to missing cert in signature when the verifier is CA cert.
func (suite *CosignMiddlewareTestSuite) TestStrictSigVerificationFailWithMissingCertInSig() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.signedArtifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	// Read test image index, but with a signature including no cert.
	host := strings.TrimPrefix(suite.registry.URL, "http://")
	// Push to the regsitry the hello-world image.
	suite.Require().Nil(writeImageIndex("hello-world", fmt.Sprintf("%s/library/hello-world:latest", host)))
	// Push to the regsitry the signature of hello-world image that doesn't include cert for verification.
	suite.Require().Nil(writeImageIndex("hello-world-sig", fmt.Sprintf("%s/library/hello-world:sha256-ad7c8b818cf3016b2b64372e166b339a8e8f0f722a812a867ae20d9a46171dff.sig", host)))
	// Write CA cert to the cosignCAPath.
	suite.Require().Nil(writeFileToDstPath("test.crt", cosignCAPath))
	os.Setenv(cosignStrictVerificationEnabled, "true")

	req := suite.makeRequest()
	rr := httptest.NewRecorder()
	Cosign()(suite.next).ServeHTTP(rr, req)

	suite.Equal(rr.Code, http.StatusPreconditionFailed)
}

// pull artifact with signature verification fail due to neither CA certs nor public key existing.
func (suite *CosignMiddlewareTestSuite) TestStrictSigVerificationFailWithMissingVerifier() {
	mock.OnAnything(suite.artifactController, "GetByReference").Return(suite.signedArtifact, nil)
	mock.OnAnything(suite.projectController, "GetByName").Return(suite.project, nil)
	mock.OnAnything(suite.accessMgr, "List").Return([]accessorymodel.Accessory{}, nil)
	// Read test image index, but with a signature including no cert.
	host := strings.TrimPrefix(suite.registry.URL, "http://")
	// Push to the regsitry the hello-world image.
	suite.Require().Nil(writeImageIndex("hello-world", fmt.Sprintf("%s/library/hello-world:latest", host)))
	// Push to the registry the signature of hello-world that includes cert for verification.
	suite.Require().Nil(writeImageIndex("hello-world-sig-with-cert", fmt.Sprintf("%s/library/hello-world:sha256-ad7c8b818cf3016b2b64372e166b339a8e8f0f722a812a867ae20d9a46171dff.sig", host)))
	os.Setenv(cosignStrictVerificationEnabled, "true")

	req := suite.makeRequest()
	rr := httptest.NewRecorder()
	Cosign()(suite.next).ServeHTTP(rr, req)

	suite.Equal(rr.Code, http.StatusPreconditionFailed)
}

func writeImageIndex(indexPath, remoteRef string) error {
	index, err := layout.ImageIndexFromPath(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read image index from path: %w", err)
	}
	reference, err := name.ParseReference(remoteRef)
	if err != nil {
		return fmt.Errorf("failed to parse reference %s: %w", remoteRef, err)
	}
	if err := remote.WriteIndex(reference, index); err != nil {
		return fmt.Errorf("failed to push to registry: %w", err)
	}
	return nil
}

func writeFileToDstPath(testFile, dstPath string) error {
	data, err := ioutil.ReadFile(testFile)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(dstPath, data, 0644); err != nil {
		return err
	}
	return nil
}

func TestCosignMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &CosignMiddlewareTestSuite{})
}
