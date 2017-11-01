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
