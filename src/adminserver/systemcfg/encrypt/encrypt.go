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
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/utils"
)

// Encryptor encrypts or decrypts a strings
type Encryptor interface {
	// Encrypt encrypts plaintext
	Encrypt(string) (string, error)
	// Decrypt decrypts ciphertext
	Decrypt(string) (string, error)
}

// AESEncryptor uses AES to encrypt or decrypt string
type AESEncryptor struct {
	keyProvider comcfg.KeyProvider
	keyParams   map[string]interface{}
}

// NewAESEncryptor returns an instance of an AESEncryptor
func NewAESEncryptor(keyProvider comcfg.KeyProvider,
	keyParams map[string]interface{}) Encryptor {
	return &AESEncryptor{
		keyProvider: keyProvider,
	}
}

// Encrypt ...
func (a *AESEncryptor) Encrypt(plaintext string) (string, error) {
	key, err := a.keyProvider.Get(a.keyParams)
	if err != nil {
		return "", err
	}
	return utils.ReversibleEncrypt(plaintext, key)
}

// Decrypt ...
func (a *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	key, err := a.keyProvider.Get(a.keyParams)
	if err != nil {
		return "", err
	}
	return utils.ReversibleDecrypt(ciphertext, key)
}
