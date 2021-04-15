//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package encrypt

import (
	"io/ioutil"
)

// KeyProvider provides the key used to encrypt and decrypt attrs
type KeyProvider interface {
	// Get returns the key
	// params can be used to pass parameters in different implements
	Get(params map[string]interface{}) (string, error)
}

// FileKeyProvider reads key from file
type FileKeyProvider struct {
	path string
}

// NewFileKeyProvider returns an instance of FileKeyProvider
// path: where the key should be read from
func NewFileKeyProvider(path string) KeyProvider {
	return &FileKeyProvider{
		path: path,
	}
}

// Get returns the key read from file
func (f *FileKeyProvider) Get(params map[string]interface{}) (string, error) {
	b, err := ioutil.ReadFile(f.path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// PresetKeyProvider returns the preset key disregarding the parm, this is for testing only
type PresetKeyProvider struct {
	Key string
}

// Get ...
func (p *PresetKeyProvider) Get(params map[string]interface{}) (string, error) {
	return p.Key, nil
}
