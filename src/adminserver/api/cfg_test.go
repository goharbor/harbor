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

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/adminserver/systemcfg"
	"github.com/vmware/harbor/src/common"
)

type fakeCfgStore struct {
	cfgs map[string]interface{}
	err  error
}

func (f *fakeCfgStore) Name() string {
	return "fake"
}

func (f *fakeCfgStore) Read() (map[string]interface{}, error) {
	return f.cfgs, f.err
}

func (f *fakeCfgStore) Write(cfgs map[string]interface{}) error {
	f.cfgs = cfgs
	return f.err
}

func TestListCfgs(t *testing.T) {
	// 500
	systemcfg.CfgStore = &fakeCfgStore{
		cfgs: nil,
		err:  errors.New("error"),
	}

	w := httptest.NewRecorder()
	ListCfgs(w, nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// 200
	key := "key"
	value := "value"
	cfgs := map[string]interface{}{
		key: value,
	}
	systemcfg.CfgStore = &fakeCfgStore{
		cfgs: cfgs,
		err:  nil,
	}
	w = httptest.NewRecorder()
	ListCfgs(w, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	result, err := parse(w.Body)
	if err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	assert.Equal(t, value, result[key])
}

func TestUpdateCfgs(t *testing.T) {
	// 400
	w := httptest.NewRecorder()
	r, err := http.NewRequest("", "", bytes.NewReader([]byte{'a'}))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	UpdateCfgs(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 500
	systemcfg.CfgStore = &fakeCfgStore{
		cfgs: nil,
		err:  errors.New("error"),
	}
	w = httptest.NewRecorder()
	r, err = http.NewRequest("", "", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	UpdateCfgs(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// 200
	key := "key"
	value := "value"
	systemcfg.CfgStore = &fakeCfgStore{
		cfgs: nil,
		err:  nil,
	}
	w = httptest.NewRecorder()
	r, err = http.NewRequest("", "",
		bytes.NewBufferString(fmt.Sprintf(`{"%s":"%s"}`, key, value)))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	UpdateCfgs(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestResetCfgs(t *testing.T) {
	// 500
	systemcfg.CfgStore = &fakeCfgStore{
		cfgs: nil,
		err:  errors.New("error"),
	}
	w := httptest.NewRecorder()

	ResetCfgs(w, nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// 200
	os.Clearenv()
	key := "LDAP_URL"
	value := "ldap://ldap.com"
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	store := &fakeCfgStore{
		cfgs: nil,
		err:  nil,
	}
	systemcfg.CfgStore = store
	w = httptest.NewRecorder()

	ResetCfgs(w, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, value, store.cfgs[common.LDAPURL])
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
