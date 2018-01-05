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

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/test"
	"github.com/vmware/harbor/src/ui/auth"
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

	retCode := m.Run()
	os.Exit(retCode)
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

func TestSearchUser(t *testing.T) {
	var username = "test"
	var auth *Auth
	user, err := auth.SearchUser(username)
	if err != nil {
		t.Errorf("Search user failed %v", err)
	}
	if user == nil {
		t.Errorf("Search user failed %v", user)
	}
}
func TestSearchUser_02(t *testing.T) {
	var username = "nonexist"
	var auth *Auth
	user, _ := auth.SearchUser(username)
	if user != nil {
		t.Errorf("Should failed to search nonexist user")
	}

}

func TestOnboardUser(t *testing.T) {
	user := &models.User{
		Username: "sample",
		Email:    "sample@example.com",
		Realname: "Sample",
	}

	var auth *Auth
	err := auth.OnBoardUser(user)
	if err != nil {
		t.Errorf("Failed to onboard user")
	}
	if user.UserID <= 0 {
		t.Errorf("Failed to onboard user")
	}
}

func TestAuthenticateHelperOnboardUser(t *testing.T) {
	user := models.User{
		Username: "test01",
		Realname: "test01",
		Email:    "test01@example.com",
	}

	err := auth.OnBoardUser(&user)
	if err != nil {
		t.Errorf("Failed to onboard user error: %v", err)
	}

	if user.UserID <= 0 {
		t.Errorf("Failed to onboard user, userid: %v", user.UserID)
	}

}

func TestAuthenticateHelperSearchUser(t *testing.T) {

	user, err := auth.SearchUser("test")
	if err != nil {
		t.Error("Failed to search user, test")
	}

	if user == nil {
		t.Error("Failed to search user test")
	}
}

func TestPostAuthentication(t *testing.T) {

	assert := assert.New(t)
	user1 := &models.User{
		Username: "test003",
		Email:    "test003@vmware.com",
		Realname: "test003",
	}

	queryCondition := models.User{
		Username: "test003",
		Realname: "test003",
	}

	err := auth.OnBoardUser(user1)
	assert.Nil(err)

	user2 := &models.User{
		Username: "test003",
		Email:    "234invalidmail@@@@@",
	}

	auth.PostAuthenticate(user2)

	dbUser, err := dao.GetUser(queryCondition)
	if err != nil {
		t.Fatalf("Failed to get user, error %v", err)
	}
	assert.EqualValues("test003@vmware.com", dbUser.Email)

	user3 := &models.User{
		Username: "test003",
	}

	auth.PostAuthenticate(user3)
	dbUser, err = dao.GetUser(queryCondition)
	if err != nil {
		t.Fatalf("Failed to get user, error %v", err)
	}
	assert.EqualValues("test003@vmware.com", dbUser.Email)

	user4 := &models.User{
		Username: "test003",
		Email:    "test003@example.com",
	}

	auth.PostAuthenticate(user4)

	dbUser, err = dao.GetUser(queryCondition)
	if err != nil {
		t.Fatalf("Failed to get user, error %v", err)
	}
	assert.EqualValues("test003@example.com", dbUser.Email)
	dao.CleanUser(int64(dbUser.UserID))
}
