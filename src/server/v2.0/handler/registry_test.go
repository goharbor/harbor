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
	"testing"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

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

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, &RegistryTestSuite{})
}
