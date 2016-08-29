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

package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestParseEndpoint(t *testing.T) {
	endpoint := "example.com"
	u, err := ParseEndpoint(endpoint)
	if err != nil {
		t.Fatalf("failed to parse endpoint %s: %v", endpoint, err)
	}

	if u.String() != "http://example.com" {
		t.Errorf("unexpected endpoint: %s != %s", endpoint, "http://example.com")
	}

	endpoint = "https://example.com"
	u, err = ParseEndpoint(endpoint)
	if err != nil {
		t.Fatalf("failed to parse endpoint %s: %v", endpoint, err)
	}

	if u.String() != "https://example.com" {
		t.Errorf("unexpected endpoint: %s != %s", endpoint, "https://example.com")
	}

	endpoint = "  example.com/ "
	u, err = ParseEndpoint(endpoint)
	if err != nil {
		t.Fatalf("failed to parse endpoint %s: %v", endpoint, err)
	}

	if u.String() != "http://example.com" {
		t.Errorf("unexpected endpoint: %s != %s", endpoint, "http://example.com")
	}
}

func TestParseRepository(t *testing.T) {
	repository := "library/ubuntu"
	project, rest := ParseRepository(repository)
	if project != "library" {
		t.Errorf("unexpected project: %s != %s", project, "library")
	}
	if rest != "ubuntu" {
		t.Errorf("unexpected rest: %s != %s", rest, "ubuntu")
	}

	repository = "library/test/ubuntu"
	project, rest = ParseRepository(repository)
	if project != "library/test" {
		t.Errorf("unexpected project: %s != %s", project, "library/test")
	}
	if rest != "ubuntu" {
		t.Errorf("unexpected rest: %s != %s", rest, "ubuntu")
	}

	repository = "ubuntu"
	project, rest = ParseRepository(repository)
	if project != "" {
		t.Errorf("unexpected project: [%s] != [%s]", project, "")
	}

	if rest != "ubuntu" {
		t.Errorf("unexpected rest: %s != %s", rest, "ubuntu")
	}

	repository = ""
	project, rest = ParseRepository(repository)
	if project != "" {
		t.Errorf("unexpected project: [%s] != [%s]", project, "")
	}

	if rest != "" {
		t.Errorf("unexpected rest: [%s] != [%s]", rest, "")
	}
}

func TestEncrypt(t *testing.T) {
	content := "content"
	salt := "salt"
	result := Encrypt(content, salt)

	if result != "dc79e76c88415c97eb089d9cc80b4ab0" {
		t.Errorf("unexpected result: %s != %s", result, "dc79e76c88415c97eb089d9cc80b4ab0")
	}
}

func TestReversibleEncrypt(t *testing.T) {
	password := "password"
	key := "1234567890123456"
	encrypted, err := ReversibleEncrypt(password, key)
	if err != nil {
		t.Errorf("Failed to encrypt: %v", err)
	}
	t.Logf("Encrypted password: %s", encrypted)
	if encrypted == password {
		t.Errorf("Encrypted password is identical to the original")
	}
	if !strings.HasPrefix(encrypted, EncryptHeaderV1) {
		t.Errorf("Encrypted password does not have v1 header")
	}
	decrypted, err := ReversibleDecrypt(encrypted, key)
	if err != nil {
		t.Errorf("Failed to decrypt: %v", err)
	}
	if decrypted != password {
		t.Errorf("decrypted password: %s, is not identical to original", decrypted)
	}
	//Test b64 for backward compatibility
	b64password := base64.StdEncoding.EncodeToString([]byte(password))
	decrypted, err = ReversibleDecrypt(b64password, key)
	if err != nil {
		t.Errorf("Failed to decrypt: %v", err)
	}
	if decrypted != password {
		t.Errorf("decrypted password: %s, is not identical to original", decrypted)
	}
}
