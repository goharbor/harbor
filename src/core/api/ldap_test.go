// Copyright 2018 Project Harbor Authors
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
package api

import (
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/stretchr/testify/assert"
)

var ldapTestConfig = map[string]interface{}{
	common.ExtEndpoint:            "host01.com",
	common.AUTHMode:               "ldap_auth",
	common.DatabaseType:           "postgresql",
	common.PostGreSQLHOST:         "127.0.0.1",
	common.PostGreSQLPort:         5432,
	common.PostGreSQLUsername:     "postgres",
	common.PostGreSQLPassword:     "root123",
	common.PostGreSQLDatabase:     "registry",
	common.LDAPURL:                "ldap://127.0.0.1",
	common.LDAPSearchDN:           "cn=admin,dc=example,dc=com",
	common.LDAPSearchPwd:          "admin",
	common.LDAPBaseDN:             "dc=example,dc=com",
	common.LDAPUID:                "uid",
	common.LDAPFilter:             "",
	common.LDAPScope:              2,
	common.LDAPTimeout:            30,
	common.AdminInitialPassword:   "password",
	common.LDAPGroupSearchFilter:  "objectclass=groupOfNames",
	common.LDAPGroupBaseDN:        "dc=example,dc=com",
	common.LDAPGroupAttributeName: "cn",
	common.LDAPGroupSearchScope:   2,
	common.LDAPGroupAdminDn:       "cn=harbor_users,ou=groups,dc=example,dc=com",
}

func TestLdapGroupsSearch(t *testing.T) {

	fmt.Println("Testing Ldap Groups Search")
	assert := assert.New(t)
	config.InitWithSettings(ldapTestConfig)
	apiTest := newHarborAPI()

	ldapGroup := apilib.LdapGroupsSearch{
		GroupName: "harbor_users",
		GroupDN:   "cn=harbor_users,ou=groups,dc=example,dc=com",
	}

	// case 1: search group by name
	code, groups, err := apiTest.LdapGroupsSearch(ldapGroup.GroupName, "", *admin)
	if err != nil {
		t.Error("Error occurred while search ldap groups", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Search ldap group status should be 200")
		assert.Equal(1, len(groups), "Search ldap groups record should be 1")
		assert.Equal(ldapGroup.GroupDN, groups[0].GroupDN, "Group DNs should be equal")
		assert.Equal(ldapGroup.GroupName, groups[0].GroupName, "Group names should be equal")
	}

	// case 2: search group by DN
	code, groups, err = apiTest.LdapGroupsSearch("", ldapGroup.GroupDN, *admin)
	if err != nil {
		t.Error("Error occurred while search ldap groups", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Search ldap groups status should be 200")
		assert.Equal(1, len(groups), "Search ldap groups record should be 1 ")
		assert.Equal(ldapGroup.GroupDN, groups[0].GroupDN, "Group DNs should be equal")
		assert.Equal(ldapGroup.GroupName, groups[0].GroupName, "Group names should be equal")
	}

	config.InitWithSettings(test.GetDefaultConfigMap())
}
