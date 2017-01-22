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

	comcfg "github.com/vmware/harbor/src/common/config"
)

// test functions under adminserver/systemcfg
func TestSystemcfg(t *testing.T) {
	key := "JSON_STORE_PATH"
	path := "/tmp/config.json"
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			t.Fatalf("failed to remove %s: %v", path, err)
		}
	} else if !os.IsNotExist(err) {
		t.Fatalf("failed to check the existence of %s: %v", path, err)
	}

	if err := os.Setenv(key, path); err != nil {
		t.Fatalf("failed to set env %s: %v", key, err)
	}

	m := map[string]string{
		"AUTH_MODE":        comcfg.DBAuth,
		"LDAP_SCOPE":       "1",
		"LDAP_TIMEOUT":     "30",
		"MYSQL_PORT":       "3306",
		"MAX_JOB_WORKERS":  "3",
		"TOKEN_EXPIRATION": "30",
		"CFG_EXPIRATION":   "5",
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
	defer func() {
		if err := os.Remove(path); err != nil {
			t.Fatalf("failed to remove %s: %v", path, err)
		}
	}()

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

	if cfg.Authentication.Mode != comcfg.DBAuth {
		t.Errorf("unexpected auth mode: %s != %s",
			cfg.Authentication.Mode, comcfg.DBAuth)
		return
	}

	cfg.Authentication.Mode = comcfg.LDAPAuth
	if err = UpdateSystemCfg(cfg); err != nil {
		t.Errorf("failed to update system configurations: %v", err)
		return
	}

	cfg, err = GetSystemCfg()
	if err != nil {
		t.Errorf("failed to get system configurations: %v", err)
		return
	}

	if cfg.Authentication.Mode != comcfg.LDAPAuth {
		t.Errorf("unexpected auth mode: %s != %s",
			cfg.Authentication.Mode, comcfg.DBAuth)
		return
	}
}
