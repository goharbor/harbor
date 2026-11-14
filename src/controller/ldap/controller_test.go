//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ldap

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/config/models"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/ldap"
)

var defaultConfigWithVerifyCert = map[string]any{
	common.ExtEndpoint:                "https://host01.com",
	common.AUTHMode:                   common.LDAPAuth,
	common.DatabaseType:               "postgresql",
	common.PostGreSQLHOST:             "127.0.0.1",
	common.PostGreSQLPort:             5432,
	common.PostGreSQLUsername:         "postgres",
	common.PostGreSQLPassword:         "root123",
	common.PostGreSQLDatabase:         "registry",
	common.SelfRegistration:           true,
	common.LDAPURL:                    "ldap://127.0.0.1:389",
	common.LDAPSearchDN:               "cn=admin,dc=example,dc=com",
	common.LDAPSearchPwd:              "admin",
	common.LDAPBaseDN:                 "dc=example,dc=com",
	common.LDAPUID:                    "uid",
	common.LDAPFilter:                 "",
	common.LDAPScope:                  2,
	common.LDAPTimeout:                30,
	common.LDAPVerifyCert:             false,
	common.LDAPGroupBaseDN:            "ou=groups,dc=example,dc=com",
	common.LDAPGroupSearchScope:       2,
	common.LDAPGroupSearchFilter:      "objectclass=groupOfNames",
	common.LDAPGroupAttributeName:     "cn",
	common.TokenServiceURL:            "http://token_service",
	common.RegistryURL:                "http://registry",
	common.EmailHost:                  "127.0.0.1",
	common.EmailPort:                  25,
	common.EmailUsername:              "user01",
	common.EmailPassword:              "password",
	common.EmailFrom:                  "from",
	common.EmailSSL:                   true,
	common.EmailIdentity:              "",
	common.ProjectCreationRestriction: common.ProCrtRestrAdmOnly,
	common.MaxJobWorkers:              3,
	common.TokenExpiration:            30,
	common.AdminInitialPassword:       "password",
}

var ldapCfg = models.LdapConf{
	URL:               "ldap://127.0.0.1",
	SearchDn:          "cn=admin,dc=example,dc=com",
	SearchPassword:    "admin",
	BaseDn:            "dc=example,dc=com",
	UID:               "cn",
	Scope:             2,
	ConnectionTimeout: 30,
}

var ldapCfgNoPwd = models.LdapConf{
	URL:               "ldap://127.0.0.1",
	SearchDn:          "cn=admin,dc=example,dc=com",
	BaseDn:            "dc=example,dc=com",
	UID:               "cn",
	Scope:             2,
	ConnectionTimeout: 30,
}

var groupCfg = models.GroupConf{
	BaseDN:              "dc=example,dc=com",
	NameAttribute:       "cn",
	SearchScope:         2,
	Filter:              "objectclass=groupOfNames",
	MembershipAttribute: "memberof",
}

type controllerTestSuite struct {
	htesting.Suite
	controller Controller
}

func (c *controllerTestSuite) SetupTest() {
	c.controller = Ctl
	config.Upload(defaultConfigWithVerifyCert)
}

func (c *controllerTestSuite) TestPing() {
	result, err := c.controller.Ping(c.Context(), ldapCfg)
	c.Nil(err)
	c.True(result)
}

func (c *controllerTestSuite) TestPingNoPassword() {
	result, err := c.controller.Ping(c.Context(), ldapCfgNoPwd)
	c.Nil(err)
	c.True(result)
}

func (c *controllerTestSuite) TestPingEmptyRequestUsesSystemConfig() {
	mgr := &ldap.Manager{}
	var pinged models.LdapConf
	mgr.On("Ping", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		pinged = args.Get(1).(models.LdapConf)
	}).Return(true, nil)
	c.controller = &controller{mgr: mgr}

	// an empty request must be tested against the configuration saved in the system
	result, err := c.controller.Ping(c.Context(), models.LdapConf{})
	c.Nil(err)
	c.True(result)
	c.Equal(defaultConfigWithVerifyCert[common.LDAPURL], pinged.URL)
	c.Equal(defaultConfigWithVerifyCert[common.LDAPSearchDN], pinged.SearchDn)
	c.Equal(defaultConfigWithVerifyCert[common.LDAPSearchPwd], pinged.SearchPassword)
	c.Equal(defaultConfigWithVerifyCert[common.LDAPBaseDN], pinged.BaseDn)
	c.Equal(defaultConfigWithVerifyCert[common.LDAPUID], pinged.UID)
	c.Equal(defaultConfigWithVerifyCert[common.LDAPScope], pinged.Scope)
	c.Equal(defaultConfigWithVerifyCert[common.LDAPTimeout], pinged.ConnectionTimeout)
	c.Equal(defaultConfigWithVerifyCert[common.LDAPVerifyCert], pinged.VerifyCert)
	mgr.AssertExpectations(c.T())
}

func (c *controllerTestSuite) TestPingRequestConfigIsUsedAsIs() {
	mgr := &ldap.Manager{}
	var pinged models.LdapConf
	mgr.On("Ping", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		pinged = args.Get(1).(models.LdapConf)
	}).Return(true, nil)
	c.controller = &controller{mgr: mgr}

	// a request that carries its own URL must not be overridden by the saved configuration
	req := models.LdapConf{
		URL:      "ldap://other.example.com:389",
		SearchDn: "cn=other,dc=example,dc=com",
		BaseDn:   "dc=other,dc=com",
		UID:      "sAMAccountName",
		Scope:    2,
	}
	result, err := c.controller.Ping(c.Context(), req)
	c.Nil(err)
	c.True(result)
	c.Equal(req.URL, pinged.URL)
	c.Equal(req.SearchDn, pinged.SearchDn)
	c.Equal(req.BaseDn, pinged.BaseDn)
	c.Equal(req.UID, pinged.UID)
	// only the password falls back to the saved one
	c.Equal(defaultConfigWithVerifyCert[common.LDAPSearchPwd], pinged.SearchPassword)
	mgr.AssertExpectations(c.T())
}

func (c *controllerTestSuite) TestSearchUser() {
	users, err := c.controller.SearchUser(c.Context(), "mike02")
	c.Nil(err)
	c.True(len(users) > 0)
}

func (c *controllerTestSuite) TestSearchGroup() {
	groups, err := c.controller.SearchGroup(c.Context(), "", "cn=harbor_dev,ou=groups,dc=example,dc=com")
	c.Nil(err)
	c.True(len(groups) > 0)
}

func (c *controllerTestSuite) TestImportUser() {
	mgr := &ldap.Manager{}
	mock.OnAnything(mgr, "ImportUser").Return(nil, nil)
	c.controller = &controller{mgr: mgr}
	failedUsers, err := c.controller.ImportUser(c.Context(), []string{"mike02"})
	c.Nil(err)
	c.True(len(failedUsers) == 0)
}

func (c *controllerTestSuite) TestSession() {
	session, err := c.controller.Session(c.Context())
	c.Nil(err)
	c.NotNil(session)
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}
