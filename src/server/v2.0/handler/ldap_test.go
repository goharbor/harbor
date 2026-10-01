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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/pkg/ldap"
	ldapModel "github.com/goharbor/harbor/src/pkg/ldap/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

// fakeLdapCtl records the configuration the handler asks it to ping
type fakeLdapCtl struct {
	pinged []cfgModels.LdapConf
}

func (f *fakeLdapCtl) Ping(_ context.Context, cfg cfgModels.LdapConf) (bool, error) {
	f.pinged = append(f.pinged, cfg)
	return true, nil
}

func (f *fakeLdapCtl) SearchUser(context.Context, string) ([]ldapModel.User, error) {
	return nil, nil
}

func (f *fakeLdapCtl) ImportUser(context.Context, []string) ([]ldapModel.FailedImportUser, error) {
	return nil, nil
}

func (f *fakeLdapCtl) SearchGroup(context.Context, string, string) ([]ldapModel.Group, error) {
	return nil, nil
}

func (f *fakeLdapCtl) Session(context.Context) (*ldap.Session, error) {
	return nil, nil
}

type LdapTestSuite struct {
	htesting.Suite

	ctl *fakeLdapCtl
}

func (suite *LdapTestSuite) SetupSuite() {
	suite.ctl = &fakeLdapCtl{}
	suite.Config = &restapi.Config{
		LdapAPI: &ldapAPI{ctl: suite.ctl},
	}
	suite.Suite.SetupSuite()
}

func (suite *LdapTestSuite) SetupTest() {
	suite.ctl.pinged = nil
	suite.Security.On("IsAuthenticated").Return(true)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.Security.On("GetUsername").Return("admin")
}

func (suite *LdapTestSuite) TearDownTest() {
	suite.Security.ExpectedCalls = nil
}

func (suite *LdapTestSuite) TestPingWithoutBody() {
	// no body at all: the handler must not dereference a nil payload
	res, err := suite.DoReq(http.MethodPost, "/ldap/ping", nil)
	suite.NoError(err)
	suite.Equal(http.StatusOK, res.StatusCode)

	var result models.LdapPingResult
	suite.NoError(json.NewDecoder(res.Body).Decode(&result))
	suite.True(result.Success)
	suite.Require().Len(suite.ctl.pinged, 1)
	suite.Equal(cfgModels.LdapConf{}, suite.ctl.pinged[0])
}

func (suite *LdapTestSuite) TestPingWithEmptyObject() {
	res, err := suite.DoReq(http.MethodPost, "/ldap/ping", bytes.NewBufferString("{}"))
	suite.NoError(err)
	suite.Equal(http.StatusOK, res.StatusCode)
	suite.Require().Len(suite.ctl.pinged, 1)
	suite.Equal(cfgModels.LdapConf{}, suite.ctl.pinged[0])
}

func (suite *LdapTestSuite) TestPingWithConfig() {
	body, err := json.Marshal(&models.LdapConf{
		LdapURL:            "ldap://ldap.example.com:389",
		LdapSearchDn:       "cn=admin,dc=example,dc=com",
		LdapSearchPassword: "secret",
		LdapBaseDn:         "dc=example,dc=com",
		LdapUID:            "uid",
		LdapScope:          2,
		LdapVerifyCert:     true,
	})
	suite.NoError(err)

	res, err := suite.DoReq(http.MethodPost, "/ldap/ping", bytes.NewBuffer(body))
	suite.NoError(err)
	suite.Equal(http.StatusOK, res.StatusCode)
	suite.Require().Len(suite.ctl.pinged, 1)
	suite.Equal(cfgModels.LdapConf{
		URL:            "ldap://ldap.example.com:389",
		SearchDn:       "cn=admin,dc=example,dc=com",
		SearchPassword: "secret",
		BaseDn:         "dc=example,dc=com",
		UID:            "uid",
		Scope:          2,
		VerifyCert:     true,
	}, suite.ctl.pinged[0])
}

func TestLdapTestSuite(t *testing.T) {
	suite.Run(t, &LdapTestSuite{})
}
