// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package ldap

import (
	//"fmt"
	//"strings"
	"os"
	"testing"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/test"
	uiConfig "github.com/vmware/harbor/src/ui/config"
)

var adminServerLdapTestConfig = map[string]interface{}{
	common.ExtEndpoint:   "host01.com",
	common.AUTHMode:      "ldap_auth",
	common.DatabaseType:  "mysql",
	common.MySQLHost:     "127.0.0.1",
	common.MySQLPort:     3306,
	common.MySQLUsername: "root",
	common.MySQLPassword: "root123",
	common.MySQLDatabase: "registry",
	common.SQLiteFile:    "/tmp/registry.db",
	//config.SelfRegistration: true,
	common.LDAPURL:       "ldap://127.0.0.1",
	common.LDAPSearchDN:  "cn=admin,dc=example,dc=com",
	common.LDAPSearchPwd: "admin",
	common.LDAPBaseDN:    "dc=example,dc=com",
	common.LDAPUID:       "uid",
	common.LDAPFilter:    "",
	common.LDAPScope:     3,
	common.LDAPTimeout:   30,
	//	config.TokenServiceURL:            "",
	//	config.RegistryURL:                "",
	//	config.EmailHost:                  "",
	//	config.EmailPort:                  25,
	//	config.EmailUsername:              "",
	//	config.EmailPassword:              "password",
	//	config.EmailFrom:                  "from",
	//	config.EmailSSL:                   true,
	//	config.EmailIdentity:              "",
	//	config.ProjectCreationRestriction: config.ProCrtRestrAdmOnly,
	//	config.VerifyRemoteCert:           false,
	//	config.MaxJobWorkers:              3,
	//	config.TokenExpiration:            30,
	common.CfgExpiration: 5,
	//	config.JobLogDir:                  "/var/log/jobs",
	common.AdminInitialPassword: "password",
}

func TestMain(t *testing.T) {
	server, err := test.NewAdminserver(adminServerLdapTestConfig)
	if err != nil {
		t.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

	if err := os.Setenv("ADMIN_SERVER_URL", server.URL); err != nil {
		t.Fatalf("failed to set env %s: %v", "ADMIN_SERVER_URL", err)
	}

	secretKeyPath := "/tmp/secretkey"
	_, err = test.GenerateKey(secretKeyPath)
	if err != nil {
		t.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		t.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}

	if err := uiConfig.Init(); err != nil {
		t.Fatalf("failed to initialize configurations: %v", err)
	}

	//	if err := uiConfig.Load(); err != nil {
	//		t.Fatalf("failed to load configurations: %v", err)
	//	}

	//	mode, err := uiConfig.AuthMode()
	//	if err != nil {
	//		t.Fatalf("failed to get auth mode: %v", err)
	//	}

	database, err := uiConfig.Database()
	if err != nil {
		log.Fatalf("failed to get database configuration: %v", err)
	}

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
}

func TestGetSystemLdapConf(t *testing.T) {

	testLdapConfig, err := GetSystemLdapConf()

	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}

	if testLdapConfig.LdapURL != "ldap://127.0.0.1" {
		t.Errorf("unexpected LdapURL: %s != %s", testLdapConfig.LdapURL, "ldap://test.ldap.com")
	}
}

func TestValidateLdapConf(t *testing.T) {
	testLdapConfig, err := GetSystemLdapConf()
	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}

	testLdapConfig, err = ValidateLdapConf(testLdapConfig)

	if testLdapConfig.LdapScope != 2 {
		t.Errorf("unexpected LdapScope: %d != %d", testLdapConfig.LdapScope, 2)
	}
}

func TestMakeFilter(t *testing.T) {
	testLdapConfig, err := GetSystemLdapConf()

	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}

	testLdapConfig.LdapFilter = "(ou=people)"
	tempUsername := ""

	tempFilter := MakeFilter(tempUsername, testLdapConfig.LdapFilter, testLdapConfig.LdapUID)
	if tempFilter != "(&(ou=people)(uid=*))" {
		t.Errorf("unexpected tempFilter: %s != %s", tempFilter, "(&(ou=people)(uid=*))")
	}

	tempUsername = "user0001"
	tempFilter = MakeFilter(tempUsername, testLdapConfig.LdapFilter, testLdapConfig.LdapUID)
	if tempFilter != "(&(ou=people)(uid=user0001))" {
		t.Errorf("unexpected tempFilter: %s != %s", tempFilter, "(&(ou=people)(uid=user0001)")
	}
}

func TestFormatLdapURL(t *testing.T) {
	testLdapConfig, err := GetSystemLdapConf()

	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}

	testLdapConfig.LdapURL = "test.ldap.com"
	tempLdapURL, err := formatLdapURL(testLdapConfig.LdapURL)

	if err != nil {
		t.Errorf("failed to format Ldap URL %v", err)
	}

	if tempLdapURL != "ldap://test.ldap.com:389" {
		t.Errorf("unexpected tempLdapURL: %s != %s", tempLdapURL, "ldap://test.ldap.com:389")
	}

	testLdapConfig.LdapURL = "ldaps://test.ldap.com"
	tempLdapURL, err = formatLdapURL(testLdapConfig.LdapURL)

	if err != nil {
		t.Errorf("failed to format Ldap URL %v", err)
	}

	if tempLdapURL != "ldaps://test.ldap.com:636" {
		t.Errorf("unexpected tempLdapURL: %s != %s", tempLdapURL, "ldap://test.ldap.com:636")
	}
}

func TestImportUser(t *testing.T) {
	var u models.LdapUser
	var user models.User
	u.Username = "ldapUser0001"
	u.Realname = "ldapUser"
	_, err := ImportUser(u)
	if err != nil {
		t.Fatalf("failed to add Ldap user: %v", err)
	}

	user.Username = "ldapUser0001"
	user.Email = "ldapUser0001@placeholder.com"

	exist, err := dao.UserExists(user, "username")
	if !exist {
		t.Errorf("failed to add Ldap username: %v", err)
	}

	exist, err = dao.UserExists(user, "email")
	if !exist {
		t.Errorf("failed to add Ldap user email: %v", err)
	}

	_, err = ImportUser(u)
	if err.Error() != "duplicate_username" {
		t.Fatalf("failed to checking duplicate user: %v", err)
	}

}

func TestConnectTest(t *testing.T) {

	testLdapConfig, err := GetSystemLdapConf()

	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}

	testLdapConfig.LdapURL = "ldap://localhost:389"

	err = ConnectTest(testLdapConfig)
	if err != nil {
		t.Errorf("unexpected ldap connect fail: %v", err)
	}
}

func TestSearchUser(t *testing.T) {

	testLdapConfig, err := GetSystemLdapConf()

	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}
	testLdapConfig.LdapURL = "ldap://localhost:389"
	testLdapConfig.LdapFilter = MakeFilter("", testLdapConfig.LdapFilter, testLdapConfig.LdapUID)

	ldapUsers, err := SearchUser(testLdapConfig)
	if err != nil {
		t.Errorf("unexpected ldap search fail: %v", err)
	}

	if ldapUsers[0].Username != "test" {
		t.Errorf("unexpected ldap user search result: %s = %s", "ldapUsers[0].Username", ldapUsers[0].Username)
	}
}
