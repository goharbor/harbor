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

func TestAuthenticate(t *testing.T) {
	var person models.AuthModel
	var auth *Auth
	person.Principal = "test"
	person.Password = "123456"
	user, err := auth.Authenticate(person)
	if err != nil {
		t.Errorf("unexpected ldap authenticate fail: %v", err)
	}
	if user.Username != "test" {
		t.Errorf("unexpected ldap user authenticate fail: %s = %s", "user.Username", user.Username)
	}
	person.Principal = "test"
	person.Password = "1"
	user, err = auth.Authenticate(person)
	if err != nil {
		t.Errorf("unexpected ldap error: %v", err)
	}
	if user != nil {
		t.Errorf("Nil user expected for wrong password")
	}
	person.Principal = ""
	person.Password = ""
	user, err = auth.Authenticate(person)
	if err != nil {
		t.Errorf("unexpected ldap error: %v", err)
	}
	if user != nil {
		t.Errorf("Nil user for empty credentials")
	}
}
