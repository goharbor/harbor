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
package config

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config/client/db"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
)

// test functions under package core/config
var adminServerDefaultConfig = map[string]interface{}{
	common.ExtEndpoint:                "https://host01.com",
	common.AUTHMode:                   common.DBAuth,
	common.DatabaseType:               "postgresql",
	common.PostGreSQLHOST:             "127.0.0.1",
	common.PostGreSQLPort:             5432,
	common.PostGreSQLUsername:         "postgres",
	common.PostGreSQLPassword:         "root123",
	common.PostGreSQLDatabase:         "registry",
	common.SelfRegistration:           true,
	common.LDAPURL:                    "ldap://127.0.0.1",
	common.LDAPSearchDN:               "uid=searchuser,ou=people,dc=mydomain,dc=com",
	common.LDAPSearchPwd:              "password",
	common.LDAPBaseDN:                 "ou=people,dc=mydomain,dc=com",
	common.LDAPUID:                    "uid",
	common.LDAPFilter:                 "",
	common.LDAPScope:                  3,
	common.LDAPTimeout:                30,
	common.LDAPGroupBaseDN:            "dc=example,dc=com",
	common.LDAPGroupSearchFilter:      "objectClass=groupOfNames",
	common.LDAPGroupSearchScope:       2,
	common.LDAPGroupAttributeName:     "cn",
	common.TokenServiceURL:            "http://token_service",
	common.RegistryURL:                "http://registry",
	common.EmailHost:                  "127.0.0.1",
	common.EmailPort:                  25,
	common.EmailUsername:              "user01",
	common.EmailPassword:              "password",
	common.EmailFrom:                  "from",
	common.EmailSSL:                   true,
	common.EmailInsecure:              false,
	common.EmailIdentity:              "",
	common.ProjectCreationRestriction: common.ProCrtRestrAdmOnly,
	common.MaxJobWorkers:              3,
	common.TokenExpiration:            30,
	common.CfgExpiration:              5,
	common.AdminInitialPassword:       "password",
	common.AdmiralEndpoint:            "http://www.vmware.com",
	common.WithNotary:                 false,
	common.WithClair:                  false,
	common.ClairDBUsername:            "postgres",
	common.ClairDBHost:                "postgresql",
	common.ClairDB:                    "postgres",
	common.ClairDBPort:                5432,
	common.ClairDBPassword:            "root123",
	common.UAAClientID:                "testid",
	common.UAAClientSecret:            "testsecret",
	common.UAAEndpoint:                "10.192.168.5",
	common.UAAVerifyCert:              false,
	common.CoreURL:                    "http://myui:8888/",
	common.JobServiceURL:              "http://myjob:8888/",
	common.ReadOnly:                   false,
	common.NotaryURL:                  "http://notary-server:4443",
}

func TestConfig(t *testing.T) {

	defaultCACertPath = path.Join(currPath(), "test", "ca.crt")
	c := map[string]interface{}{
		common.AdmiralEndpoint: "http://www.vmware.com",
	}
	server, err := test.NewAdminserver(c)
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
	assert := assert.New(t)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		t.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}

	db.InitDatabaseAndConfigure()
	cfgManager := db.NewCoreConfigManager()
	cfgManager.Upload(adminServerDefaultConfig)

	if err := Init(); err != nil {
		t.Fatalf("failed to initialize configurations: %v", err)
	}

	if err := Load(); err != nil {
		t.Fatalf("failed to load configurations: %v", err)
	}

	if err := Upload(map[string]interface{}{}); err != nil {
		t.Fatalf("failed to upload configurations: %v", err)
	}

	if _, err := GetSystemCfg(); err != nil {
		t.Fatalf("failed to get system configurations: %v", err)
	}

	mode, err := AuthMode()
	if err != nil {
		t.Fatalf("failed to get auth mode: %v", err)
	}
	if mode != "db_auth" {
		t.Errorf("unexpected mode: %s != %s", mode, "db_auth")
	}

	if _, err := LDAPConf(); err != nil {
		t.Fatalf("failed to get ldap settings: %v", err)
	}

	if _, err := LDAPGroupConf(); err != nil {
		t.Fatalf("failed to get ldap group settings: %v", err)
	}

	if _, err := TokenExpiration(); err != nil {
		t.Fatalf("failed to get token expiration: %v", err)
	}

	if _, err := ExtEndpoint(); err != nil {
		t.Fatalf("failed to get domain name: %v", err)
	}

	if _, err := SecretKey(); err != nil {
		t.Fatalf("failed to get secret key: %v", err)
	}

	if _, err := SelfRegistration(); err != nil {
		t.Fatalf("failed to get self registration: %v", err)
	}

	if _, err := RegistryURL(); err != nil {
		t.Fatalf("failed to get registry URL: %v", err)
	}

	if len(InternalJobServiceURL()) == 0 {
		t.Error("the internal job service url is null")
	}

	if len(InternalTokenServiceEndpoint()) == 0 {
		t.Error("the internal token service endpoint is null")
	}

	if _, err := InitialAdminPassword(); err != nil {
		t.Fatalf("failed to get initial admin password: %v", err)
	}

	if _, err := OnlyAdminCreateProject(); err != nil {
		t.Fatalf("failed to get onldy admin create project: %v", err)
	}

	if _, err := Email(); err != nil {
		t.Fatalf("failed to get email settings: %v", err)
	}

	if _, err := Database(); err != nil {
		t.Fatalf("failed to get database: %v", err)
	}

	clairDB, err := ClairDB()
	if err != nil {
		t.Fatalf("failed to get clair DB %v", err)
	}
	adminServerDefaultConfig := test.GetDefaultConfigMap()
	assert.Equal(adminServerDefaultConfig[common.ClairDB], clairDB.Database)
	assert.Equal(adminServerDefaultConfig[common.ClairDBUsername], clairDB.Username)
	assert.Equal(adminServerDefaultConfig[common.ClairDBPassword], clairDB.Password)
	assert.Equal(adminServerDefaultConfig[common.ClairDBHost], clairDB.Host)
	assert.Equal(adminServerDefaultConfig[common.ClairDBPort], clairDB.Port)

	if InternalNotaryEndpoint() != "http://notary-server:4443" {
		t.Errorf("Unexpected notary endpoint: %s", InternalNotaryEndpoint())
	}
	if WithNotary() {
		t.Errorf("Withnotary should be false")
	}
	if WithClair() {
		t.Errorf("WithClair should be false")
	}
	if !WithAdmiral() {
		t.Errorf("WithAdmiral should be true")
	}
	if ReadOnly() {
		t.Errorf("ReadOnly should be false")
	}
	if AdmiralEndpoint() != "http://www.vmware.com" {
		t.Errorf("Unexpected admiral endpoint: %s", AdmiralEndpoint())
	}

	extURL, err := ExtURL()
	if err != nil {
		t.Errorf("Unexpected error getting external URL: %v", err)
	}
	if extURL != "host01.com" {
		t.Errorf(`extURL should be "host01.com".`)
	}

	// reset configurations
	if err = Reset(); err != nil {
		t.Errorf("failed to reset configurations: %v", err)
		return
	}
	mode, err = AuthMode()
	if err != nil {
		t.Fatalf("failed to get auth mode: %v", err)
	}
	if mode != "db_auth" {
		t.Errorf("unexpected mode: %s != %s", mode, "db_auth")
	}

	if s := ScanAllPolicy(); s.Type != "none" {
		t.Errorf("unexpected scan all policy %v", s)
	}

	if tokenKeyPath := TokenPrivateKeyPath(); tokenKeyPath != "/etc/core/private_key.pem" {
		t.Errorf("Unexpected token private key path: %s, expected: %s", tokenKeyPath, "/etc/core/private_key.pem")
	}

	us, err := UAASettings()
	if err != nil {
		t.Fatalf("failed to get UAA setting, error: %v", err)
	}
	if us.ClientID != "testid" || us.ClientSecret != "testsecret" || us.Endpoint != "10.192.168.5" || us.VerifyCert {
		t.Errorf("Unexpected UAA setting: %+v", *us)
	}
	assert.Equal("http://myjob:8888", InternalJobServiceURL())
	assert.Equal("http://myui:8888/service/token", InternalTokenServiceEndpoint())

}

func currPath() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Dir(f)
}
