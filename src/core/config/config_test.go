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
	"encoding/json"
	"os"
	"path"
	"runtime"
	"testing"

	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
)

// test functions under package core/config
func TestConfig(t *testing.T) {
	test.InitDatabaseFromEnv()
	dao.PrepareTestData([]string{"delete from properties where k='scan_all_policy'"}, []string{})
	defaultCACertPath = path.Join(currPath(), "test", "ca.crt")
	c := map[string]interface{}{
		common.AdmiralEndpoint: "https://www.vmware.com",
		common.WithClair:       false,
		common.WithChartMuseum: false,
		common.WithNotary:      false,
	}
	Init()

	Upload(c)

	secretKeyPath := "/tmp/secretkey"
	_, err := test.GenerateKey(secretKeyPath)
	if err != nil {
		t.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)
	assert := assert.New(t)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		t.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}
	oriKeyPath := os.Getenv("TOKEN_PRIVATE_KEY_PATH")
	if err := os.Setenv("TOKEN_PRIVATE_KEY_PATH", ""); err != nil {
		t.Fatalf("failed to set env %s: %v", "TOKEN_PRIVATE_KEY_PATH", err)
	}
	defer os.Setenv("TOKEN_PRIVATE_KEY_PATH", oriKeyPath)

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

	tkExp := RobotTokenDuration()
	assert.Equal(tkExp, 43200)

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
	defaultConfig := test.GetDefaultConfigMap()
	defaultConfig[common.AdmiralEndpoint] = "http://www.vmware.com"
	Upload(defaultConfig)
	assert.Equal(defaultConfig[common.ClairDB], clairDB.Database)
	assert.Equal(defaultConfig[common.ClairDBUsername], clairDB.Username)
	assert.Equal(defaultConfig[common.ClairDBPassword], clairDB.Password)
	assert.Equal(defaultConfig[common.ClairDBHost], clairDB.Host)
	assert.Equal(defaultConfig[common.ClairDBPort], clairDB.Port)

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
func TestConfigureValue_GetMap(t *testing.T) {
	var policy models.ScanAllPolicy
	value2 := `{"parameter":{"daily_time":0},"type":"daily"}`
	err := json.Unmarshal([]byte(value2), &policy)
	if err != nil {
		t.Errorf("Failed with error %v", err)
	}
	fmt.Printf("%+v\n", policy)
}

func TestHTTPAuthProxySetting(t *testing.T) {
	m := map[string]interface{}{
		common.HTTPAuthProxySkipSearch: "true",
		common.HTTPAuthProxyVerifyCert: "true",
		common.HTTPAuthProxyEndpoint:   "https://auth.proxy/suffix",
	}
	InitWithSettings(m)
	v, e := HTTPAuthProxySetting()
	assert.Nil(t, e)
	assert.Equal(t, *v, models.HTTPAuthProxy{
		Endpoint:   "https://auth.proxy/suffix",
		SkipSearch: true,
		VerifyCert: true,
	})
}

func TestOIDCSetting(t *testing.T) {
	m := map[string]interface{}{
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://oidc.test",
		common.OIDCVerifyCert:   "true",
		common.OIDCScope:        "openid, profile",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "secret",
		common.ExtEndpoint:      "https://harbor.test",
	}
	InitWithSettings(m)
	v, e := OIDCSetting()
	assert.Nil(t, e)
	assert.Equal(t, "test", v.Name)
	assert.Equal(t, "https://oidc.test", v.Endpoint)
	assert.True(t, v.VerifyCert)
	assert.Equal(t, "client", v.ClientID)
	assert.Equal(t, "secret", v.ClientSecret)
	assert.Equal(t, "https://harbor.test/c/oidc/callback", v.RedirectURL)
	assert.ElementsMatch(t, []string{"openid", "profile"}, v.Scope)
}
