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

	"github.com/vmware/harbor/src/common/models"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
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

func TestMain(m *testing.M) {
	server, err := test.NewAdminserver(adminServerLdapTestConfig)
	if err != nil {
		log.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

	if err := os.Setenv("ADMINSERVER_URL", server.URL); err != nil {
		log.Fatalf("failed to set env %s: %v", "ADMINSERVER_URL", err)
	}

	secretKeyPath := "/tmp/secretkey"
	_, err = test.GenerateKey(secretKeyPath)
	if err != nil {
		log.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		log.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}

	if err := uiConfig.Init(); err != nil {
		log.Fatalf("failed to initialize configurations: %v", err)
	}

	database, err := uiConfig.Database()
	if err != nil {
		log.Fatalf("failed to get database configuration: %v", err)
	}

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	os.Exit(m.Run())

}

func TestLoadSystemLdapConfig(t *testing.T) {
	session, err := LoadSystemLdapConfig()
	if err != nil {
		t.Fatalf("failed to get system ldap config %v", err)
	}

	if session.ldapConfig.LdapURL != "ldap://127.0.0.1:389" {
		t.Errorf("unexpected LdapURL: %s != %s", session.ldapConfig.LdapURL, "ldap://127.0.0.1:389")
	}

}

func TestConnectTest(t *testing.T) {
	session, err := LoadSystemLdapConfig()
	if err != nil {
		t.Errorf("failed to load system ldap config")
	}
	err = session.ConnectionTest()
	if err != nil {
		t.Errorf("Unexpected ldap connect fail: %v", err)
	}

}

func TestCreateWithConfig(t *testing.T) {
	var testConfigs = []struct {
		config        models.LdapConf
		internalValue int
	}{
		{
			models.LdapConf{
				LdapScope: 3,
				LdapURL:   "ldaps://127.0.0.1",
			}, 2},
		{
			models.LdapConf{
				LdapScope: 2,
				LdapURL:   "ldaps://127.0.0.1",
			}, 1},
		{
			models.LdapConf{
				LdapScope: 1,
				LdapURL:   "ldaps://127.0.0.1",
			}, 0},
		{
			models.LdapConf{
				LdapScope: 1,
				LdapURL:   "ldaps://127.0.0.1:abc",
			}, -1},
	}

	for _, val := range testConfigs {
		_, err := CreateWithConfig(val.config)
		if val.internalValue < 0 {
			if err == nil {
				t.Fatalf("Should have error with url :%v", val.config)
			}
			continue
		}
		if err != nil {
			t.Fatalf("Can not create with ui config, err:%v", err)
		}
	}

}

func TestSearchUser(t *testing.T) {

	session, err := LoadSystemLdapConfig()
	if err != nil {
		t.Fatalf("Can not load system ldap config")
	}
	err = session.Open()
	if err != nil {
		t.Fatalf("failed to create ldap session %v", err)
	}

	err = session.Bind(session.ldapConfig.LdapSearchDn, session.ldapConfig.LdapSearchPassword)
	if err != nil {
		t.Fatalf("failed to bind search dn")
	}

	defer session.Close()

	result, err := session.SearchUser("test")
	if err != nil || len(result) == 0 {
		t.Fatalf("failed to search user test!")
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
		{"ldap:\\wrong url", ""},
	}

	for _, u := range urls {
		goodURL, err := formatURL(u.rawURL)
		if u.goodURL == "" {
			if err == nil {
				t.Fatalf("Should failed on wrong url, %v", u.rawURL)
			}
			continue
		}
		if err != nil || goodURL != u.goodURL {
			t.Fatalf("Faild on URL: raw=%v, expected:%v, actual:%v", u.rawURL, u.goodURL, goodURL)
		}
	}

}
