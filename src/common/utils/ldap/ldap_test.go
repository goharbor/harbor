/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package ldap

import (
	//"fmt"
	//"strings"
	"os"
	"testing"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	ldapConfig "github.com/vmware/harbor/src/ui/config"
)

func TestMain(t *testing.T) {
	os.Setenv("AUTH_MODE", "ldap_auth")
	os.Setenv("LDAP_URL", "ldap://127.0.0.1")
	os.Setenv("LDAP_BASE_DN", "dc=example,dc=com")
	os.Setenv("LDAP_SEARCH_DN", "cn=admin,dc=example,dc=com")
	os.Setenv("LDAP_SEARCH_PWD", "admin")
	os.Setenv("LDAP_UID", "uid")
	os.Setenv("LDAP_SCOPE", "3")
	os.Setenv("LDAP_FILTER", "")
	os.Setenv("LDAP_CONNECT_TIMEOUT", "1")

	db := os.Getenv("DATABASE")
	defer os.Setenv("DATABASE", db)

	os.Setenv("DATABASE", "mysql")

	dbHost := os.Getenv("DB_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable DB_HOST is not set")
	}
	dbUser := os.Getenv("DB_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable DB_USR is not set")
	}
	dbPort := os.Getenv("DB_PORT")
	if len(dbPort) == 0 {
		log.Fatalf("environment variable DB_PORT is not set")
	}
	dbPassword := os.Getenv("DB_PWD")

	log.Infof("DB_HOST: %s, DB_USR: %s, DB_PORT: %s, DB_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	os.Setenv("MYSQL_HOST", dbHost)
	os.Setenv("MYSQL_PORT", dbPort)
	os.Setenv("MYSQL_USR", dbUser)
	os.Setenv("MYSQL_PWD", dbPassword)

	err := ldapConfig.Reload()
	if err != nil {
		panic(err)
	}

	dao.InitDatabase()
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
