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

package encrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeCfgStore struct {
	cfgs map[string]interface{}
}

func (f *fakeCfgStore) Name() string {
	return "fake"
}

func (f *fakeCfgStore) Read() (map[string]interface{}, error) {
	return f.cfgs, nil
}

func (f *fakeCfgStore) Write(cfgs map[string]interface{}) error {
	f.cfgs = cfgs
	return nil
}

type fakeEncryptor struct {
}

func (f *fakeEncryptor) Encrypt(plaintext string) (string, error) {
	return "encrypted" + plaintext, nil
}

func (f *fakeEncryptor) Decrypt(ciphertext string) (string, error) {
	return "decrypted" + ciphertext, nil
}

func TestName(t *testing.T) {
	driver := NewCfgStore(nil, nil, nil)
	assert.Equal(t, name, driver.Name())
}

func TestRead(t *testing.T) {
	keys := []string{"key"}
	driver := NewCfgStore(&fakeEncryptor{}, keys, &fakeCfgStore{
		cfgs: map[string]interface{}{"key": "value"},
	})

	cfgs, err := driver.Read()
	assert.Nil(t, err)
	assert.Equal(t, "decryptedvalue", cfgs["key"])
}

func TestWrite(t *testing.T) {
	keys := []string{"key"}
	store := &fakeCfgStore{
		cfgs: map[string]interface{}{},
	}
	driver := NewCfgStore(&fakeEncryptor{}, keys, store)

	cfgs := map[string]interface{}{
		"key": "value",
	}

	err := driver.Write(cfgs)
	assert.Nil(t, err)
	assert.Equal(t, "encryptedvalue", store.cfgs["key"])
}
