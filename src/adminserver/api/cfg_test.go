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

package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/vmware/harbor/src/adminserver/config"
	"github.com/vmware/harbor/src/adminserver/systemcfg"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/utils/test"
)

func TestConfigAPI(t *testing.T) {
	configPath := "/tmp/config.json"
	secretKeyPath := "/tmp/secretkey"

	_, err := test.GenerateKey(secretKeyPath)
	if err != nil {
		t.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)

	secret := "secret"
	/*
		secretPlaintext := "secret"
		secretCiphertext, err := utils.ReversibleEncrypt(secretPlaintext, string(data))
		if err != nil {
			t.Errorf("failed to encrypt secret: %v", err)
			return
		}
	*/
	envs := map[string]string{

		"JSON_STORE_PATH":       configPath,
		"KEY_PATH":              secretKeyPath,
		"UI_SECRET":             secret,
		"MYSQL_PORT":            "3306",
		"TOKEN_EXPIRATION":      "30",
		"CFG_EXPIRATION":        "5",
		"MAX_JOB_WORKERS":       "3",
		"LDAP_SCOPE":            "3",
		"LDAP_TIMEOUT":          "30",
		"EMAIL_PORT":            "25",
		"MYSQL_PWD":             "",
		"LDAP_SEARCH_PWD":       "",
		"EMAIL_PWD":             "",
		"HARBOR_ADMIN_PASSWORD": "",
	}

	for k, v := range envs {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("failed to set env %s: %v", k, err)
		}
	}
	defer os.Remove(configPath)

	if err := config.Init(); err != nil {
		t.Errorf("failed to load configurations of adminserver: %v", err)
		return
	}

	if err := systemcfg.Init(); err != nil {
		t.Errorf("failed to initialize system configurations: %v", err)
		return
	}

	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Errorf("failed to create request: %v", err)
		return
	}

	w := httptest.NewRecorder()
	ListCfgs(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusUnauthorized)
		return
	}

	r.AddCookie(&http.Cookie{
		Name:  "secret",
		Value: secret,
	})

	w = httptest.NewRecorder()
	ListCfgs(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusOK)
		return
	}

	m, err := parse(w.Body)
	if err != nil {
		t.Errorf("failed to parse response body: %v", err)
		return
	}

	scope := int(m[comcfg.LDAPScope].(float64))
	if scope != 3 {
		t.Errorf("unexpected ldap scope: %d != %d", scope, 3)
		return
	}

	// modify configurations
	c := map[string]interface{}{
		comcfg.AUTHMode: comcfg.LDAPAuth,
	}

	b, err := json.Marshal(c)
	if err != nil {
		t.Errorf("failed to marshal configuartions: %v", err)
		return
	}

	w = httptest.NewRecorder()
	r, err = http.NewRequest("GET", "", bytes.NewReader(b))
	if err != nil {
		t.Errorf("failed to create request: %v", err)
		return
	}
	r.AddCookie(&http.Cookie{
		Name:  "secret",
		Value: secret,
	})

	UpdateCfgs(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusOK)
		return
	}

	// confirm the modification is done
	r, err = http.NewRequest("GET", "", nil)
	if err != nil {
		t.Errorf("failed to create request: %v", err)
		return
	}
	r.AddCookie(&http.Cookie{
		Name:  "secret",
		Value: secret,
	})
	w = httptest.NewRecorder()
	ListCfgs(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("unexpected status code: %d != %d", w.Code, http.StatusOK)
		return
	}

	m, err = parse(w.Body)
	if err != nil {
		t.Errorf("failed to parse response body: %v", err)
		return
	}

	mode := m[comcfg.AUTHMode].(string)
	if mode != comcfg.LDAPAuth {
		t.Errorf("unexpected ldap scope: %s != %s", mode, comcfg.LDAPAuth)
		return
	}
}

func parse(reader io.Reader) (map[string]interface{}, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
