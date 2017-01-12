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

/*
import (
	"os"
	"testing"
)


func TestSystemcfg(t *testing.T) {
	key := "JSON_STORE_PATH"
	tmpPath := "/tmp/config.json"
	originalPath := os.Getenv(key)
	defer func() {
		if err := os.Remove(tmpPath); err != nil {
			t.Errorf("failed to remove %s: %v", tmpPath, err)
		}

		if len(originalPath) == 0 {
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("failed to unset env %s: %v", key, err)
			}
			return
		}

		if err := os.Setenv(key, originalPath); err != nil {
			t.Fatalf("failed to set env %s: %v", key, err)
		}
	}()

	if err := os.Setenv(key, tmpPath); err != nil {
		t.Fatalf("failed to set env %s: %v", key, err)
	}

	m := map[string]string{
		"LDAP_SCOPE":       "1",
		"LDAP_TIMEOUT":     "30",
		"MYSQL_PORT":       "3306",
		"MAX_JOB_WORKERS":  "3",
		"TOKEN_EXPIRATION": "30",
		"CFG_EXPIRATION":   "5",
	}

	for k, v := range m {
		if err := os.Setenv(k, v); err != nil {
			t.Errorf("failed to set env %s: %v", k, err)
			return
		}
	}

	if err := Init(); err != nil {
		t.Errorf("failed to initialize system configurations: %v", err)
		return
	}
}
*/
