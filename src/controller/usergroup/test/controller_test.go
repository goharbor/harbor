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

package test

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/controller/usergroup"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type controllerTestSuite struct {
	htesting.Suite
	controller usergroup.Controller
}

func (c *controllerTestSuite) SetupTest() {
	c.controller = usergroup.Ctl
	c.Suite.ClearTables = []string{"user_group"}
}

var defaultConfigWithVerifyCert = map[string]interface{}{
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
	common.WithNotary:                 false,
}

func (c *controllerTestSuite) TestCRUDUserGroup() {
	config.InitWithSettings(defaultConfigWithVerifyCert)
	ctx := c.Context()
	ug := model.UserGroup{
		GroupName:   "harbor_dev",
		GroupType:   1,
		LdapGroupDN: "cn=harbor_dev,ou=groups,dc=example,dc=com",
	}
	id, err := c.controller.Create(ctx, ug)
	c.Nil(err)
	c.True(id > 0)

	ug2, err2 := c.controller.Get(ctx, id)
	c.Nil(err2)
	c.Equal(ug2.GroupName, "harbor_dev")
	c.Equal(ug2.GroupType, 1)
	c.Equal(ug2.LdapGroupDN, "cn=harbor_dev,ou=groups,dc=example,dc=com")

	err3 := c.controller.Update(ctx, id, "my_harbor_dev")
	c.Nil(err3)

	ug4, err4 := c.controller.Get(ctx, id)
	c.Nil(err4)
	c.Equal(ug4.GroupName, "my_harbor_dev")

	err5 := c.controller.Delete(ctx, id)
	c.Nil(err5)
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}
