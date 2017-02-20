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

// Package config provide methods to get the configurations reqruied by code in src/common
package config

import (
	"io/ioutil"
)

// KeyProvider provides the secret key used to encrypt and decrypt attrs
type KeyProvider interface {
	// Get returns the secret key
	Get() (string, error)
}

// KeyFileProvider reads key from file
type KeyFileProvider struct {
	path string
}

// NewKeyFileProvider returns an instance of KeyFileProvider
// path: where the key should be read from
func NewKeyFileProvider(path string) KeyProvider {
	return &KeyFileProvider{
		path: path,
	}
}

// Get ...
func (kfp *KeyFileProvider) Get() (string, error) {
	b, err := ioutil.ReadFile(kfp.path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
