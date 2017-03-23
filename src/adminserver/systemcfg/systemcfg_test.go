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

package systemcfg

import (
	"os"
	"testing"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/utils/test"
)

// test functions in adminserver/systemcfg/systemcfg.go
func TestSystemcfg(t *testing.T) {
	configPath := "/tmp/config.json"
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			t.Errorf("failed to remove %s: %v", configPath, err)
			return
		}
	} else if !os.IsNotExist(err) {
		t.Errorf("failed to check the existence of %s: %v", configPath, err)
		return
	}

	if err := os.Setenv("JSON_CFG_STORE_PATH", configPath); err != nil {
		t.Errorf("failed to set env: %v", err)
		return
	}

	keyPath := "/tmp/secretkey"
	if _, err := test.GenerateKey(keyPath); err != nil {
		t.Errorf("failed to generate key: %v", err)
		return
	}
	defer os.Remove(keyPath)

	if err := os.Setenv("KEY_PATH", keyPath); err != nil {
		t.Errorf("failed to set env: %v", err)
		return
	}

	m := map[string]string{
		"AUTH_MODE":             common.DBAuth,
		"LDAP_SCOPE":            "1",
		"LDAP_TIMEOUT":          "30",
		"MYSQL_PORT":            "3306",
		"MAX_JOB_WORKERS":       "3",
		"TOKEN_EXPIRATION":      "30",
		"CFG_EXPIRATION":        "5",
		"EMAIL_PORT":            "25",
		"MYSQL_PWD":             "",
		"LDAP_SEARCH_PWD":       "",
		"EMAIL_PWD":             "",
		"HARBOR_ADMIN_PASSWORD": "",
	}

	for k, v := range m {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("failed to set env %s: %v", k, err)
		}
	}

	if err := Init(); err != nil {
		t.Errorf("failed to initialize system configurations: %v", err)
		return
	}
	defer os.Remove(configPath)

	// run Init again to make sure it works well when the configuration file
	// already exists
	if err := Init(); err != nil {
		t.Errorf("failed to initialize system configurations: %v", err)
		return
	}

	cfg, err := GetSystemCfg()
	if err != nil {
		t.Errorf("failed to get system configurations: %v", err)
		return
	}

	if cfg[common.AUTHMode] != common.DBAuth {
		t.Errorf("unexpected auth mode: %s != %s",
			cfg[common.AUTHMode], common.DBAuth)
		return
	}

	cfg[common.AUTHMode] = common.LDAPAuth

	if err = UpdateSystemCfg(cfg); err != nil {
		t.Errorf("failed to update system configurations: %v", err)
		return
	}

	cfg, err = GetSystemCfg()
	if err != nil {
		t.Errorf("failed to get system configurations: %v", err)
		return
	}

	if cfg[common.AUTHMode] != common.LDAPAuth {
		t.Errorf("unexpected auth mode: %s != %s",
			cfg[common.AUTHMode], common.DBAuth)
		return
	}

	if err = Reset(); err != nil {
		t.Errorf("failed to reset system configurations: %v", err)
		return
	}

	cfg, err = GetSystemCfg()
	if err != nil {
		t.Errorf("failed to get system configurations: %v", err)
		return
	}

	if cfg[common.AUTHMode] != common.DBAuth {
		t.Errorf("unexpected auth mode: %s != %s",
			cfg[common.AUTHMode], common.DBAuth)
		return
	}
}
