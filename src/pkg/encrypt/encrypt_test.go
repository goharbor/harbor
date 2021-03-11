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
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	secret := []byte("9TXCcHgNAAp1aSHh")
	filename, err := ioutil.TempFile(os.TempDir(), "keyfile")
	err = ioutil.WriteFile(filename.Name(), secret, 0644)
	if err != nil {
		fmt.Printf("failed to create temp key file\n")
	}

	defer os.Remove(filename.Name())

	os.Setenv("KEY_PATH", filename.Name())

	ret := m.Run()
	os.Exit(ret)
}

func TestEncryptDecrypt(t *testing.T) {
	password := "zhu888jie"
	encrypted, err := Instance().Encrypt(password)
	if err != nil {
		t.Errorf("Failed to decrypt password, error %v", err)
	}
	decrypted, err := Instance().Decrypt(encrypted)
	if err != nil {
		t.Errorf("Failed to decrypt password, error %v", err)
	}
	assert.NotEqual(t, password, encrypted)
	assert.Equal(t, password, decrypted)
}
