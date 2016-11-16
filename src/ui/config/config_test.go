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
package config

import (
	"os"
	"testing"
)

var (
	auth = "ldap_auth"
	ldap = LDAPSetting{
		"ldap://test.ldap.com",
		"ou=people",
		"dc=whatever,dc=org",
		"1234567",
		"cn",
		"uid",
		"2",
	}
	tokenExp                   = "3"
	tokenExpRes                = 3
	adminPassword              = "password"
	externalRegURL             = "127.0.0.1"
	uiSecret                   = "ffadsdfsdf"
	secretKey                  = "keykey"
	selfRegistration           = "off"
	projectCreationRestriction = "adminonly"
	internalRegistryURL        = "http://registry:5000"
	jobServiceURL              = "http://jobservice"
)

func TestMain(m *testing.M) {

	os.Setenv("AUTH_MODE", auth)
	os.Setenv("LDAP_URL", ldap.URL)
	os.Setenv("LDAP_BASE_DN", ldap.BaseDn)
	os.Setenv("LDAP_SEARCH_DN", ldap.SearchDn)
	os.Setenv("LDAP_SEARCH_PWD", ldap.SearchPwd)
	os.Setenv("LDAP_UID", ldap.UID)
	os.Setenv("LDAP_SCOPE", ldap.Scope)
	os.Setenv("LDAP_FILTER", ldap.Filter)
	os.Setenv("TOKEN_EXPIRATION", tokenExp)
	os.Setenv("HARBOR_ADMIN_PASSWORD", adminPassword)
	os.Setenv("EXT_REG_URL", externalRegURL)
	os.Setenv("UI_SECRET", uiSecret)
	os.Setenv("SECRET_KEY", secretKey)
	os.Setenv("SELF_REGISTRATION", selfRegistration)
	os.Setenv("PROJECT_CREATION_RESTRICTION", projectCreationRestriction)
	os.Setenv("REGISTRY_URL", internalRegistryURL)
	os.Setenv("JOB_SERVICE_URL", jobServiceURL)

	err := Reload()
	if err != nil {
		panic(err)
	}
	rc := m.Run()

	os.Unsetenv("AUTH_MODE")
	os.Unsetenv("LDAP_URL")
	os.Unsetenv("LDAP_BASE_DN")
	os.Unsetenv("LDAP_SEARCH_DN")
	os.Unsetenv("LDAP_SEARCH_PWD")
	os.Unsetenv("LDAP_UID")
	os.Unsetenv("LDAP_SCOPE")
	os.Unsetenv("LDAP_FILTER")
	os.Unsetenv("TOKEN_EXPIRATION")
	os.Unsetenv("HARBOR_ADMIN_PASSWORD")
	os.Unsetenv("EXT_REG_URL")
	os.Unsetenv("UI_SECRET")
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("SELF_REGISTRATION")
	os.Unsetenv("CREATE_PROJECT_RESTRICTION")
	os.Unsetenv("REGISTRY_URL")
	os.Unsetenv("JOB_SERVICE_URL")

	os.Exit(rc)
}

func TestAuth(t *testing.T) {
	if AuthMode() != auth {
		t.Errorf("Expected auth mode:%s, in fact: %s", auth, AuthMode())
	}
	if LDAP() != ldap {
		t.Errorf("Expected ldap setting: %+v, in fact: %+v", ldap, LDAP())
	}
}

func TestTokenExpiration(t *testing.T) {
	if TokenExpiration() != tokenExpRes {
		t.Errorf("Expected token expiration: %d, in fact: %d", tokenExpRes, TokenExpiration())
	}
}

func TestURLs(t *testing.T) {
	if InternalRegistryURL() != internalRegistryURL {
		t.Errorf("Expected internal Registry URL: %s, in fact: %s", internalRegistryURL, InternalRegistryURL())
	}
	if InternalJobServiceURL() != jobServiceURL {
		t.Errorf("Expected internal jobservice URL: %s, in fact: %s", jobServiceURL, InternalJobServiceURL())
	}
	if ExtRegistryURL() != externalRegURL {
		t.Errorf("Expected External Registry URL: %s, in fact: %s", externalRegURL, ExtRegistryURL())
	}
}

func TestSelfRegistration(t *testing.T) {
	if SelfRegistration() {
		t.Errorf("Expected Self Registration to be false")
	}
}

func TestSecrets(t *testing.T) {
	if SecretKey() != secretKey {
		t.Errorf("Expected Secrect Key :%s, in fact: %s", secretKey, SecretKey())
	}
	if UISecret() != uiSecret {
		t.Errorf("Expected UI Secret: %s, in fact: %s", uiSecret, UISecret())
	}
}

func TestProjectCreationRestrict(t *testing.T) {
	if !OnlyAdminCreateProject() {
		t.Errorf("Expected OnlyAdminCreateProject to be true")
	}
}

func TestInitAdminPassword(t *testing.T) {
	if InitialAdminPassword() != adminPassword {
		t.Errorf("Expected adminPassword: %s, in fact: %s", adminPassword, InitialAdminPassword())
	}
}
