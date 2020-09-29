// Copyright Project Harbor Authors
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

package test

import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
)

// GenerateKey generates aes key
func GenerateKey(path string) (string, error) {
	data := make([]byte, aes.BlockSize)
	n, err := rand.Read(data)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}
	if n != aes.BlockSize {
		return "", fmt.Errorf("the length of random bytes %d != %d", n, aes.BlockSize)
	}

	if err = ioutil.WriteFile(path, data, 0777); err != nil {
		return "", fmt.Errorf("failed write secret key to file %s: %v", path, err)
	}

	return string(data), nil
}
