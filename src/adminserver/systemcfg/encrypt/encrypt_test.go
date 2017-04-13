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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	comcfg "github.com/vmware/harbor/src/common/config"
)

type fakeKeyProvider struct {
	key string
	err error
}

func (f *fakeKeyProvider) Get(params map[string]interface{}) (
	string, error) {
	return f.key, f.err
}

func TestEncrypt(t *testing.T) {
	cases := []struct {
		plaintext   string
		keyProvider comcfg.KeyProvider
		err         bool
	}{
		{"", &fakeKeyProvider{"", errors.New("error")}, true},
		{"text", &fakeKeyProvider{"1234567890123456", nil}, false},
	}

	for _, c := range cases {
		encrptor := NewAESEncryptor(c.keyProvider, nil)
		ciphertext, err := encrptor.Encrypt(c.plaintext)
		if c.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			str, err := encrptor.Decrypt(ciphertext)
			assert.Nil(t, err)
			assert.Equal(t, c.plaintext, str)
		}
	}
}

func TestDecrypt(t *testing.T) {
	plaintext := "text"
	key := "1234567890123456"

	encrptor := NewAESEncryptor(&fakeKeyProvider{
		key: key,
		err: nil,
	}, nil)

	ciphertext, err := encrptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrpt %s: %v", plaintext, err)
	}

	cases := []struct {
		ciphertext  string
		keyProvider comcfg.KeyProvider
		err         bool
	}{
		{"", &fakeKeyProvider{"", errors.New("error")}, true},
		{ciphertext, &fakeKeyProvider{key, nil}, false},
	}

	for _, c := range cases {
		encrptor := NewAESEncryptor(c.keyProvider, nil)
		str, err := encrptor.Decrypt(c.ciphertext)
		if c.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, plaintext, str)
		}
	}
}
