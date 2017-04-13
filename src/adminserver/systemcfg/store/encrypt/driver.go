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
	"github.com/vmware/harbor/src/adminserver/systemcfg/encrypt"
	"github.com/vmware/harbor/src/adminserver/systemcfg/store"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	name = "encrypt"
)

// cfgStore wraps a store.Driver with an encryptor
type cfgStore struct {
	// attrs need to be encrypted and decrypted
	keys      []string
	encryptor encrypt.Encryptor
	store     store.Driver
}

// NewCfgStore returns an instance of cfgStore
// keys are the attrs need to be encrypted or decrypted
func NewCfgStore(encryptor encrypt.Encryptor,
	keys []string, store store.Driver) store.Driver {
	return &cfgStore{
		keys:      keys,
		encryptor: encryptor,
		store:     store,
	}
}

func (c *cfgStore) Name() string {
	return name
}

func (c *cfgStore) Read() (map[string]interface{}, error) {
	m, err := c.store.Read()
	if err != nil {
		return nil, err
	}

	for _, key := range c.keys {
		v, ok := m[key]
		if !ok {
			continue
		}

		str, ok := v.(string)
		if !ok {
			log.Warningf("the value of %s is not string, skip decrypt", key)
			continue
		}

		text, err := c.encryptor.Decrypt(str)
		if err != nil {
			return nil, err
		}
		m[key] = text
	}
	return m, nil
}

func (c *cfgStore) Write(m map[string]interface{}) error {
	for _, key := range c.keys {
		v, ok := m[key]
		if !ok {
			continue
		}

		str, ok := v.(string)
		if !ok {
			log.Warningf("%v is not string, skip encrypt", v)
			continue
		}

		ciphertext, err := c.encryptor.Encrypt(str)
		if err != nil {
			return err
		}
		m[key] = ciphertext
	}
	return c.store.Write(m)
}
