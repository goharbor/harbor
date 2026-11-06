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
	"encoding/json"
	"testing"
	"time"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	registrytesting "github.com/goharbor/harbor/src/testing/controller/registry"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

type RegistryTestSuite struct {
	htesting.Suite
	regCtl *registrytesting.Controller
}

func (suite *RegistryTestSuite) SetupSuite() {
	suite.regCtl = &registrytesting.Controller{}
	suite.Config = &restapi.Config{
		RegistryAPI: &registryAPI{ctl: suite.regCtl},
	}
	suite.Suite.SetupSuite()
}

func (suite *RegistryTestSuite) SetupTest() {
	suite.regCtl.ExpectedCalls = nil
	suite.regCtl.Calls = nil
	suite.Security.ExpectedCalls = nil
	suite.Security.Calls = nil
}

func (suite *RegistryTestSuite) ptrStr(s string) *string { return &s }

// TestPingRegistryByIDIgnoresOverrides guards against CVE-class credential
// exfiltration: a caller referencing an existing registry by id must not be able
// to override its saved connection settings (url, insecure, ca certificate) and so
// redirect the health check (and the saved credentials) to an untrusted endpoint.
func (suite *RegistryTestSuite) TestPingRegistryByIDIgnoresOverrides() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()

	saved := &model.Registry{ID: 1, Type: "harbor", URL: "https://registry.example.com", Insecure: false}
	mock.OnAnything(suite.regCtl, "Get").Return(saved, nil).Once()

	var pinged *model.Registry
	suite.regCtl.On("IsHealthy", mock.Anything, mock.Anything).Return(true, nil).Once().
		Run(func(args testifymock.Arguments) { pinged = args.Get(1).(*model.Registry) })

	id := int64(1)
	insecure := true
	res, err := suite.PostJSON("/registries/ping", &models.RegistryPing{
		ID:            &id,
		URL:           suite.ptrStr("https://attacker.example.com"),
		Insecure:      &insecure,
		CaCertificate: suite.ptrStr("-----BEGIN CERTIFICATE-----\nattacker\n-----END CERTIFICATE-----"),
	})
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Require().NotNil(pinged)
	// every supplied override is ignored; the saved settings are used
	suite.Equal("https://registry.example.com", pinged.URL)
	suite.False(pinged.Insecure)
	suite.Empty(pinged.CACertificate)
}

// TestPingRegistryInlineUsesSuppliedURL confirms inline pings (no id) still honor
// the supplied URL, so the fix does not regress the normal ad-hoc ping flow.
func (suite *RegistryTestSuite) TestPingRegistryInlineUsesSuppliedURL() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()

	var pinged *model.Registry
	suite.regCtl.On("IsHealthy", mock.Anything, mock.Anything).Return(true, nil).Once().
		Run(func(args testifymock.Arguments) { pinged = args.Get(1).(*model.Registry) })

	res, err := suite.PostJSON("/registries/ping", &models.RegistryPing{
		Type: suite.ptrStr("harbor"),
		URL:  suite.ptrStr("https://inline.example.com"),
	})
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Require().NotNil(pinged)
	suite.Equal("https://inline.example.com", pinged.URL)
}

// TestUpdateRegistryID0ReturnsNotFound guards against updating synthetic local registry ID 0,
// asserting the NotFound response and that controller's Get and Update methods are not called.
func (suite *RegistryTestSuite) TestUpdateRegistryID0ReturnsNotFound() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()

	res, err := suite.PutJSON("/registries/0", &models.RegistryUpdate{
		URL: suite.ptrStr("https://attacker.example.com"),
	})
	suite.NoError(err)
	suite.Equal(404, res.StatusCode)

	suite.regCtl.AssertNotCalled(suite.T(), "Get", testifymock.Anything, testifymock.Anything)
	suite.regCtl.AssertNotCalled(suite.T(), "Update", testifymock.Anything, testifymock.Anything)
}

// TestUpdateRegistryClearsAccessSecretOnURLChange tests that updating an existing registry's
// URL without providing a new AccessSecret clears the stored AccessSecret.
func (suite *RegistryTestSuite) TestUpdateRegistryClearsAccessSecretOnURLChange() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()

	saved := &model.Registry{
		ID:         1,
		Type:       "harbor",
		URL:        "https://registry.example.com",
		Credential: &model.Credential{Type: "basic", AccessKey: "admin", AccessSecret: "secret123"},
	}
	mock.OnAnything(suite.regCtl, "Get").Return(saved, nil).Once()

	var updated *model.Registry
	suite.regCtl.On("Update", mock.Anything, mock.Anything).Return(nil).Once().
		Run(func(args testifymock.Arguments) { updated = args.Get(1).(*model.Registry) })

	res, err := suite.PutJSON("/registries/1", &models.RegistryUpdate{
		URL: suite.ptrStr("https://new.example.com"),
	})
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Require().NotNil(updated)
	suite.Equal("https://new.example.com", updated.URL)
	suite.Empty(updated.Credential.AccessSecret)
}

// TestUpdateRegistryInvalidURLReturnsError tests that updating a registry with an invalid
// URL returns BadRequest error and does not update the registry.
func (suite *RegistryTestSuite) TestUpdateRegistryInvalidURLReturnsError() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()

	saved := &model.Registry{
		ID:         1,
		Type:       "harbor",
		URL:        "https://registry.example.com",
		Credential: &model.Credential{Type: "basic", AccessKey: "admin", AccessSecret: "secret123"},
	}
	mock.OnAnything(suite.regCtl, "Get").Return(saved, nil).Once()

	res, err := suite.PutJSON("/registries/1", &models.RegistryUpdate{
		URL: suite.ptrStr("ftp://invalid.example.com"),
	})
	suite.NoError(err)
	suite.Equal(400, res.StatusCode)
	suite.regCtl.AssertNotCalled(suite.T(), "Update", testifymock.Anything, testifymock.Anything)
}

func (suite *RegistryTestSuite) TestPreviewRegistryArtifact() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()

	modified := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	var gotFilter *model.FileFilter
	suite.regCtl.On("Preview", mock.Anything, int64(3), "Org/Model", "main", mock.Anything).
		Return(&model.ResolveResult{
			Repository: "org/model", SourceID: "Org/Model", Reference: "main",
			Revision: "93efa2f097d58c2a74874c7e644dbc9b0cee75a2", RevisionTag: "sha-93efa2f097d5",
			Digest: "sha256:2dc8", Licenses: []string{"apache-2.0"}, Task: "text-generation",
			LastModified: &modified, TotalFiles: 2, MatchedFiles: 1, TotalSize: 30, MatchedSize: 10,
			Files: []*model.ResolvedFile{
				{Path: "config.json", Size: 10, Class: model.FileClassConfig, Matched: true},
				{Path: "model.safetensors", Size: 20, Class: model.FileClassWeight, SHA256: "ab"},
			},
		}, nil).Once().
		Run(func(args testifymock.Arguments) { gotFilter = args.Get(4).(*model.FileFilter) })

	res, err := suite.PostJSON("/registries/3/preview", &models.RegistryPreviewRequest{
		Repository: suite.ptrStr("Org/Model"),
		Reference:  "main",
		FileFilter: &models.FileFilter{Include: []string{"*.json"}, Exclude: []string{"onnx/**"}},
	})
	suite.Require().NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal(&model.FileFilter{Include: []string{"*.json"}, Exclude: []string{"onnx/**"}}, gotFilter)

	body := &models.RegistryPreview{}
	suite.Require().NoError(json.NewDecoder(res.Body).Decode(body))
	suite.Equal("sha-93efa2f097d5", body.RevisionTag)
	suite.Equal([]string{}, body.Tags)
	suite.EqualValues(1, body.MatchedFiles)
	suite.Require().Len(body.Files, 2)
	suite.Equal("weight", body.Files[1].Class)
	suite.False(body.Files[1].Matched)
	suite.Equal("2026-01-02T03:04:05.000Z", body.LastModified.String())
}

func (suite *RegistryTestSuite) TestPreviewRegistryArtifactUnsupported() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	suite.regCtl.On("Preview", mock.Anything, int64(1), "library/nginx", "", (*model.FileFilter)(nil)).
		Return(nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("the registry hub does not support previews")).Once()

	res, err := suite.PostJSON("/registries/1/preview", &models.RegistryPreviewRequest{Repository: suite.ptrStr("library/nginx")})
	suite.Require().NoError(err)
	suite.Equal(400, res.StatusCode)
}

func (suite *RegistryTestSuite) TestPreviewRegistryArtifactRequiresRepository() {
	suite.Security.On("IsAuthenticated").Return(true).Maybe()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Maybe()

	res, err := suite.PostJSON("/registries/1/preview", map[string]string{"reference": "main"})
	suite.Require().NoError(err)
	suite.Equal(422, res.StatusCode)
	suite.regCtl.AssertNotCalled(suite.T(), "Preview", testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything, testifymock.Anything)
}

func (suite *RegistryTestSuite) TestGetRegistryInfoCapabilities() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	mock.OnAnything(suite.regCtl, "GetInfo").Return(&model.RegistryInfo{
		Type: "x",
		SupportedResourceFilters: []*model.FilterStyle{
			{Type: model.FilterTypeName, Style: model.FilterStyleTypeText, Label: "Model", Placeholder: "org/model"},
		},
		SourceOnly: true, SupportedPreview: true,
	}, nil).Once()

	body := map[string]any{}
	res, err := suite.GetJSON("/registries/1/info", &body)
	suite.Require().NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal(true, body["supported_preview"])
	suite.Equal(true, body["source_only"])
	filter := body["supported_resource_filters"].([]any)[0].(map[string]any)
	suite.Equal("Model", filter["label"])
	suite.Equal("org/model", filter["placeholder"])
	suite.NotContains(filter, "description")
}

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, &RegistryTestSuite{})
}
