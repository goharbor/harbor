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
	"os"
	"testing"

	"github.com/vmware/harbor/src/common/utils/test"
)

func TestConfig(t *testing.T) {
	secretKeyPath := "/tmp/secretkey"

	secretKey, err := test.GenerateKey(secretKeyPath)
	if err != nil {
		t.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)

	secret := "secret"
	/*
		secretPlaintext := "secret"
		secretCiphertext, err := utils.ReversibleEncrypt(secretPlaintext, string(data))
		if err != nil {
			t.Errorf("failed to encrypt secret: %v", err)
			return
		}
	*/
	envs := map[string]string{
		"KEY_PATH":  secretKeyPath,
		"UI_SECRET": secret,
	}

	for k, v := range envs {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("failed to set env %s: %v", k, err)
		}
	}

	if err := Init(); err != nil {
		t.Errorf("failed to load configurations of adminserver: %v", err)
		return
	}

	if SecretKey() != secretKey {
		t.Errorf("unexpected secret key: %s != %s", SecretKey(), secretKey)
	}

	if Secret() != secret {
		t.Errorf("unexpected secret: %s != %s", Secret(), secret)
	}
}
