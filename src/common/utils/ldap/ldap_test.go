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
	common.LDAPURL:              "ldap://127.0.0.1",
	common.LDAPSearchDN:         "cn=admin,dc=example,dc=com",
	common.LDAPSearchPwd:        "admin",
	common.LDAPBaseDN:           "dc=example,dc=com",
	common.LDAPUID:              "uid",
	common.LDAPFilter:           "",
	common.LDAPScope:            3,
	common.LDAPTimeout:          30,
	common.CfgExpiration:        5,
	common.AdminInitialPassword: "password",
}

var adminServerDefaultConfigWithVerifyCert = map[string]interface{}{
	common.ExtEndpoint:                "https://host01.com",
	common.AUTHMode:                   common.LDAPAuth,
	common.DatabaseType:               "mysql",
	common.MySQLHost:                  "127.0.0.1",
	common.MySQLPort:                  3306,
	common.MySQLUsername:              "root",
	common.MySQLPassword:              "root123",
	common.MySQLDatabase:              "registry",
	common.SQLiteFile:                 "/tmp/registry.db",
	common.SelfRegistration:           true,
	common.LDAPURL:                    "ldap://127.0.0.1:389",
	common.LDAPSearchDN:               "cn=admin,dc=example,dc=com",
	common.LDAPSearchPwd:              "admin",
	common.LDAPBaseDN:                 "dc=example,dc=com",
	common.LDAPUID:                    "uid",
	common.LDAPFilter:                 "",
	common.LDAPScope:                  3,
	common.LDAPTimeout:                30,
	common.LDAPVerifyCert:             true,
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
	common.CfgExpiration:              5,
	common.AdminInitialPassword:       "password",
	common.AdmiralEndpoint:            "http://www.vmware.com",
	common.WithNotary:                 false,
	common.WithClair:                  false,
}

func TestMain(t *testing.T) {
	server, err := test.NewAdminserver(adminServerLdapTestConfig)
	if err != nil {
		t.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

	if err := os.Setenv("ADMINSERVER_URL", server.URL); err != nil {
		t.Fatalf("failed to set env %s: %v", "ADMINSERVER_URL", err)
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

	database, err := uiConfig.Database()
	if err != nil {
		log.Fatalf("failed to get database configuration: %v", err)
	}

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
}

func TestLoadSystemLdapConfig(t *testing.T) {
	var session Session
	err := session.LoadSystemLdapConfig()
	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}

	if session.ldapConfig.LdapURL != "ldap://127.0.0.1:389" {
		t.Errorf("unexpected LdapURL: %s != %s", session.ldapConfig.LdapURL, "ldap://127.0.0.1:389")
	}

	if session.ldapConfig.LdapScope != 2 {
		t.Errorf("unexpected LdapScope: %d != %d", session.ldapConfig.LdapScope, 2)
	}

}

func TestConnectTest(t *testing.T) {
	var session Session
	err := session.ConnectionTest()
	if err != nil {
		t.Errorf("Unexpected ldap connect fail: %v", err)
	}

}

func TestSearchUser(t *testing.T) {
	var session Session

	err := session.Create()
	if err != nil {
		t.Fatalf("failed to create ldap session %v", err)
	}

	err = session.BindSearchDn()
	if err != nil {
		t.Fatalf("failed to bind search dn")
	}

	defer session.Close()

	result, err := session.SearchUser("test")
	if err != nil || len(result) == 0 {
		t.Fatalf("failed to search user test!")
	}

}

func InitTest(ldapTestConfig map[string]interface{}, t *testing.T) {
	server, err := test.NewAdminserver(ldapTestConfig)
	if err != nil {
		t.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

	if err := os.Setenv("ADMIN_SERVER_URL", server.URL); err != nil {
		t.Fatalf("failed to set env %s:%v", "ADMIN_SERVER_URL", err)
	}

	if err := uiConfig.Init(); err != nil {
		t.Fatalf("failed to initialize configurations: %v ", err)
	}
}

func TestSearchAndImportUser(t *testing.T) {
	var session Session

	err := session.Create()

	if err != nil {
		t.Fatalf("failed to create ldap session: %v", err)
	}
	err = session.BindSearchDn()
	if err != nil {
		t.Fatalf("failed to bind search dn")
	}
	defer session.Close()

	userID, err := session.SearchAndImport("test")

	if err != nil {
		t.Fatalf("Failed on error : %v ", err)
	}

	if userID <= 0 {
		t.Fatalf("userID= %v", userID)
	}
}
func TestImportUser(t *testing.T) {
	var u models.LdapUser
	var user models.User
	var err error
	u.Username = "ldapUser0001"
	u.Realname = "ldapUser"

	var session Session
	err = session.Create()
	if err != nil {
		t.Fatalf("failed to create ldap session: %v", err)
	}
	err = session.BindSearchDn()
	if err != nil {
		t.Fatalf("failed to bind search dn")
	}
	defer session.Close()

	_, err = session.ImportUser(u)
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

	_, err = session.ImportUser(u)
	if err.Error() != "duplicate_username" {
		t.Fatalf("failed to checking duplicate user: %v", err)
	}
}

func TestImportUserExit(t *testing.T) {
	var users = []models.LdapUser{
		{Username: "ldapUser0003",
			Realname: "ldapUser0003",
			Email:    "anonymous@example.com"},
		{Username: "admin",
			Realname: "admin",
			Email:    "admin@example.com"},
	}
	var err error

	var session Session
	err = session.Create()
	if err != nil {
		t.Fatalf("failed to create ldap session: %v", err)
	}
	err = session.BindSearchDn()
	if err != nil {
		t.Fatalf("failed to bind search dn")
	}
	defer session.Close()
	for _, u := range users {
		_, err = session.ImportUser(u)
		if err == nil {
			t.Fatalf("Should fail on import duplicate email")
			t.Fail()
		}
	}
}

func TestSearchAndImportUserNotExist(t *testing.T) {
	var session Session

	err := session.Create()

	if err != nil {
		t.Fatalf("failed to create ldap session: %v", err)
	}
	err = session.BindSearchDn()
	if err != nil {
		t.Fatalf("failed to bind search dn")
	}
	defer session.Close()

	userID, _ := session.SearchAndImport("notexist")

	if userID > 0 {
		t.Fatal("Can not import a non exist ldap user!")
		t.Fail()
	}
}

func TestFormatURL(t *testing.T) {

	var invalidURL = "http://localhost:389"
	_, err := formatURL(invalidURL)
	if err == nil {
		t.Fatalf("Should failed on invalid URL %v", invalidURL)
		t.Fail()
	}

	var urls = []struct {
		rawURL  string
		goodURL string
	}{
		{"ldaps://127.0.0.1", "ldaps://127.0.0.1:636"},
		{"ldap://9.123.102.33", "ldap://9.123.102.33:389"},
		{"ldaps://127.0.0.1:389", "ldaps://127.0.0.1:389"},
		{"ldap://127.0.0.1:636", "ldaps://127.0.0.1:636"},
		{"112.122.122.122", "ldap://112.122.122.122:389"},
	}

	for _, u := range urls {
		goodURL, err := formatURL(u.rawURL)
		if err != nil || goodURL != u.goodURL {
			t.Fatalf("Faild on URL: raw=%v, expected:%v, actual:%v", u.rawURL, u.goodURL, goodURL)
		}
	}

}
