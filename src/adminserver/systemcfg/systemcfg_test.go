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

package systemcfg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
)

func TestParseStringToInt(t *testing.T) {
	cases := []struct {
		input  string
		result int
	}{
		{"1", 1},
		{"-1", -1},
		{"0", 0},
		{"", 0},
	}

	for _, c := range cases {
		i, err := parseStringToInt(c.input)
		assert.Nil(t, err)
		assert.Equal(t, c.result, i)
	}
}

func TestParseStringToBool(t *testing.T) {
	cases := []struct {
		input  string
		result bool
	}{
		{"true", true},
		{"on", true},
		{"TRUE", true},
		{"ON", true},
		{"other", false},
		{"", false},
	}

	for _, c := range cases {
		b, _ := parseStringToBool(c.input)
		assert.Equal(t, c.result, b)
	}
}

func TestInitCfgStore(t *testing.T) {
	os.Clearenv()
	path := "/tmp/config.json"
	if err := os.Setenv("CFG_DRIVER", "json"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("JSON_CFG_STORE_PATH", path); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer os.RemoveAll(path)
	err := initCfgStore()
	assert.Nil(t, err)
}

func TestLoadFromEnv(t *testing.T) {
	os.Clearenv()
	ldapURL := "ldap://ldap.com"
	extEndpoint := "http://harbor.com"
	if err := os.Setenv("LDAP_URL", ldapURL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	cfgs := map[string]interface{}{}
	err := LoadFromEnv(cfgs, true)
	assert.Nil(t, err)
	assert.Equal(t, ldapURL, cfgs[common.LDAPURL])

	os.Clearenv()
	if err := os.Setenv("LDAP_URL", ldapURL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("EXT_ENDPOINT", extEndpoint); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	if err := os.Setenv("LDAP_VERIFY_CERT", "false"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	cfgs = map[string]interface{}{}
	err = LoadFromEnv(cfgs, false)
	assert.Nil(t, err)
	assert.Equal(t, extEndpoint, cfgs[common.ExtEndpoint])
	assert.Equal(t, ldapURL, cfgs[common.LDAPURL])
	assert.Equal(t, false, cfgs[common.LDAPVerifyCert])

	os.Clearenv()
	if err := os.Setenv("LDAP_URL", ldapURL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("EXT_ENDPOINT", extEndpoint); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	if err := os.Setenv("LDAP_VERIFY_CERT", "true"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	cfgs = map[string]interface{}{
		common.LDAPURL: "ldap_url",
	}
	err = LoadFromEnv(cfgs, false)
	assert.Nil(t, err)
	assert.Equal(t, extEndpoint, cfgs[common.ExtEndpoint])
	assert.Equal(t, "ldap_url", cfgs[common.LDAPURL])
	assert.Equal(t, true, cfgs[common.LDAPVerifyCert])

}

func TestIsLoadAll(t *testing.T) {
	os.Clearenv()
	if err := os.Setenv("RELOAD_KEY", "123456"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("RESET", "True"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	cfg1 := map[string]interface{}{common.ReloadKey: "123456"}
	cfg2 := map[string]interface{}{common.ReloadKey: "654321"}
	assert.False(t, isLoadAll(cfg1))
	assert.True(t, isLoadAll(cfg2))
}

func TestLoadFromEnvWithReloadConfigInvalidSkipPattern(t *testing.T) {
	os.Clearenv()
	ldapURL := "ldap://ldap.com"
	extEndpoint := "http://harbor.com"
	cfgsReload := map[string]interface{}{
		common.LDAPURL: "ldap_url",
	}
	if err := os.Setenv("LDAP_URL", ldapURL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("EXT_ENDPOINT", extEndpoint); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	if err := os.Setenv("LDAP_VERIFY_CERT", "false"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	if err := os.Setenv("SKIP_RELOAD_ENV_PATTERN", "a(b"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	err := LoadFromEnv(cfgsReload, true)
	if err != nil {
		t.Fatalf("failed to load From env: %v", err)
	}
	assert.Equal(t, ldapURL, cfgsReload[common.LDAPURL])

	os.Clearenv()

}

func TestLoadFromEnvWithReloadConfigSkipPattern(t *testing.T) {
	os.Clearenv()
	ldapURL := "ldap://ldap.com"
	extEndpoint := "http://harbor.com"
	cfgsReload := map[string]interface{}{
		common.LDAPURL: "ldap_url",
	}
	if err := os.Setenv("LDAP_URL", ldapURL); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("EXT_ENDPOINT", extEndpoint); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}

	if err := os.Setenv("LDAP_VERIFY_CERT", "false"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("SKIP_RELOAD_ENV_PATTERN", "^LDAP.*"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("RESET", "true"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	err := LoadFromEnv(cfgsReload, false)
	if err != nil {
		t.Fatalf("failed to load From env: %v", err)
	}
	assert.Equal(t, "ldap_url", cfgsReload[common.LDAPURL]) //env value ignored

	os.Clearenv()

}
func TestGetDatabaseFromCfg(t *testing.T) {
	cfg := map[string]interface{}{
		common.DatabaseType:       "postgresql",
		common.PostGreSQLDatabase: "registry",
		common.PostGreSQLHOST:     "127.0.0.1",
		common.PostGreSQLPort:     5432,
		common.PostGreSQLPassword: "root123",
		common.PostGreSQLUsername: "postgres",
	}

	database := GetDatabaseFromCfg(cfg)

	assert.Equal(t, "postgresql", database.Type)
}

func TestValidLdapScope(t *testing.T) {
	var dbValue float64
	dbValue = 2
	ldapScopeKey := "ldap_scope"
	testCfgs := []struct {
		config          map[string]interface{}
		migrate         bool
		ldapScopeResult int
	}{
		{map[string]interface{}{
			ldapScopeKey: 1,
		}, true, 0},
		{map[string]interface{}{
			ldapScopeKey: 2,
		}, true, 1},
		{map[string]interface{}{
			ldapScopeKey: 3,
		}, true, 2},
		{map[string]interface{}{
			ldapScopeKey: -1,
		}, true, 0},
		{map[string]interface{}{
			ldapScopeKey: 100,
		}, false, 2},
		{map[string]interface{}{
			ldapScopeKey: -100,
		}, false, 0},
		{map[string]interface{}{
			ldapScopeKey: dbValue,
		}, false, 2},
	}

	for i, item := range testCfgs {
		validLdapScope(item.config, item.migrate)
		if item.config[ldapScopeKey].(int) != item.ldapScopeResult {
			t.Fatalf("Failed to update ldapScope expected %v, actual %v at index %v", item.ldapScopeResult, item.config[ldapScopeKey], i)
		}

	}

}
func Test_AddMissingKey(t *testing.T) {

	cfg := map[string]interface{}{
		common.LDAPURL:        "sampleurl",
		common.EmailPort:      555,
		common.LDAPVerifyCert: true,
	}

	type args struct {
		cfg map[string]interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{"Add default value", args{cfg}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddMissedKey(tt.args.cfg)
		})
	}

	if _, ok := cfg[common.LDAPBaseDN]; !ok {
		t.Errorf("Can not found default value for %v", common.LDAPBaseDN)
	}

}
