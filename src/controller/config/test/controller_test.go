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
	"context"
	"github.com/goharbor/harbor/src/common"
	. "github.com/goharbor/harbor/src/controller/config"
	"github.com/goharbor/harbor/src/lib/errors"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"

	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

var TestDBConfig = map[string]interface{}{
	common.LDAPBaseDN: "dc=example,dc=com",
	common.LDAPURL:    "ldap.example.com",
	common.EmailHost:  "127.0.0.1",
}

var TestConfigWithScanAll = map[string]interface{}{
	"postgresql_host":     "localhost",
	"postgresql_database": "registry",
	"postgresql_password": "root123",
	"postgresql_username": "postgres",
	"postgresql_sslmode":  "disable",
	"ldap_base_dn":        "dc=example,dc=com",
	"ldap_url":            "ldap.example.com",
	"email_host":          "127.0.0.1",
	"scan_all_policy":     `{"parameter":{"daily_time":0},"type":"daily"}`,
}

var ctx context.Context

type controllerTestSuite struct {
	htesting.Suite
	controller Controller
}

func (c *controllerTestSuite) SetupTest() {
	c.controller = Ctl
	ctx = c.Context()
	c.controller.UpdateUserConfigs(ctx, TestDBConfig)
}

func (c *controllerTestSuite) TestGetUserCfg() {
	resp, err := c.controller.UserConfigs(ctx)
	if err != nil {
		c.Error(err, "failed to get user config")
	}
	c.Equal("dc=example,dc=com", resp["ldap_base_dn"].Val)
	c.Equal("127.0.0.1", resp["email_host"].Val)
	c.Equal("ldap.example.com", resp["ldap_url"].Val)
}

func (c *controllerTestSuite) TestConvertForGet() {
	conf := map[string]interface{}{
		"ldap_url":             "ldaps.myexample,com",
		"ldap_base_dn":         "dc=myexample,dc=com",
		"auth_mode":            "ldap_auth",
		"ldap_search_password": "admin",
	}

	// password type should not sent to external api call
	resp, err := c.controller.ConvertForGet(ctx, conf, false)
	c.Nil(err)
	c.Equal("ldaps.myexample,com", resp["ldap_url"].Val)
	c.Equal("ldap_auth", resp["auth_mode"].Val)
	_, exist := resp["ldap_search_password"]
	c.False(exist)

	// password type should be sent to internal api call
	conf2 := map[string]interface{}{
		"ldap_url":             "ldaps.myexample,com",
		"ldap_base_dn":         "dc=myexample,dc=com",
		"auth_mode":            "ldap_auth",
		"ldap_search_password": "admin",
	}
	resp2, err2 := c.controller.ConvertForGet(ctx, conf2, true)
	c.Nil(err2)
	c.Equal("ldaps.myexample,com", resp2["ldap_url"].Val)
	c.Equal("ldap_auth", resp2["auth_mode"].Val)
	_, exist2 := resp2["ldap_search_password"]
	c.True(exist2)

}

func (c *controllerTestSuite) TestGetAll() {
	resp, err := c.controller.AllConfigs(ctx)
	if err != nil {
		c.Error(err, "failed to get user config")
	}
	c.Equal("dc=example,dc=com", resp["ldap_base_dn"])
	c.Equal("127.0.0.1", resp["email_host"])
	c.Equal("ldap.example.com", resp["ldap_url"])
}

func (c *controllerTestSuite) TestUpdateUserCfg() {

	userConf := map[string]interface{}{
		common.LDAPURL:    "ldaps.myexample,com",
		common.LDAPBaseDN: "dc=myexample,dc=com",
	}
	err := c.controller.UpdateUserConfigs(ctx, userConf)
	c.Nil(err)
	cfgResp, err := c.controller.UserConfigs(ctx)
	if err != nil {
		c.Error(err, "failed to get user config")
	}
	c.Equal("dc=myexample,dc=com", cfgResp["ldap_base_dn"].Val)
	c.Equal("ldaps.myexample,com", cfgResp["ldap_url"].Val)
	badCfg := map[string]interface{}{
		common.LDAPScope: 5,
	}
	err2 := c.controller.UpdateUserConfigs(ctx, badCfg)
	c.NotNil(err2)
	c.True(errors.IsErr(err2, errors.BadRequestCode))
}

/*func (c *controllerTestSuite) TestCheckUnmodifiable() {
	conf := map[string]interface{}{
		"ldap_url":     "ldaps.myexample,com",
		"ldap_base_dn": "dc=myexample,dc=com",
		"auth_mode":    "ldap_auth",
	}
	failed := c.controller.checkUnmodifiable(ctx, conf, "auth_mode")
	c.True(len(failed) > 0)
	c.Equal(failed[0], "auth_mode")
}
*/
func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}
