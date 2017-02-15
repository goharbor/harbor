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

package config

import (
	"io/ioutil"
	"os"

	//"github.com/vmware/harbor/src/common/utils"
)

const defaultKeyPath string = "/harbor/secretkey"

var (
	secret    string
	secretKey string
)

// Init configurations used by adminserver
func Init() error {
	path := os.Getenv("KEY_PATH")
	if len(path) == 0 {
		path = defaultKeyPath
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	secretKey = string(b)

	secret = os.Getenv("UI_SECRET")

	/*
		secretCipherText := os.Getenv("UI_SECRET")

		secret, err = utils.ReversibleDecrypt(secretCipherText, secretKey)
		if err != nil {
			return err
		}
	*/
	return nil
}

// Secret is used by API to authenticate requests
func Secret() string {
	return secret
}

// SecretKey is used to encrypt or decrypt
func SecretKey() string {
	return secretKey
}
