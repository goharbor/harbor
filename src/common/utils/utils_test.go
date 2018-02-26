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

package utils

import (
	"encoding/base64"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEndpoint(t *testing.T) {
	cases := []struct {
		input    string
		err      bool
		expected string
	}{
		{" example.com/ ", false, "http://example.com"},
		{"ftp://example.com", true, ""},
		{"http://example.com", false, "http://example.com"},
		{"https://example.com", false, "https://example.com"},
		{"http://example!@#!?//#", true, ""},
	}

	for _, c := range cases {
		u, err := ParseEndpoint(c.input)
		if c.err {
			require.NotNil(t, err)
			continue
		}
		require.Nil(t, err)
		assert.Equal(t, c.expected, u.String())
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
	if project != "library" {
		t.Errorf("unexpected project: %s != %s", project, "library/test")
	}
	if rest != "test/ubuntu" {
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

func TestGenerateRandomString(t *testing.T) {
	str := GenerateRandomString()
	if len(str) != 32 {
		t.Errorf("unexpected length: %d != %d", len(str), 32)
	}
	str2 := GenerateRandomString()
	if str2 == str {
		t.Errorf("Two identical random strings in a row: %s", str)
	}
}

func TestParseLink(t *testing.T) {
	raw := ""
	links := ParseLink(raw)
	if len(links) != 0 {
		t.Errorf("unexpected length: %d != %d", len(links), 0)
	}
	raw = "a;b,c"
	links = ParseLink(raw)
	if len(links) != 0 {
		t.Errorf("unexpected length: %d != %d", len(links), 0)
	}

	raw = `</api/users?page=1&page_size=100>; rel="prev"`
	links = ParseLink(raw)
	if len(links) != 1 {
		t.Errorf("unexpected length: %d != %d", len(links), 1)
	}
	prev := `/api/users?page=1&page_size=100`
	if links.Prev() != prev {
		t.Errorf("unexpected prev: %s != %s", links.Prev(), prev)
	}

	raw = `</api/users?page=1&page_size=100>; rel="prev", </api/users?page=3&page_size=100>; rel="next"`
	links = ParseLink(raw)
	if len(links) != 2 {
		t.Errorf("unexpected length: %d != %d", len(links), 2)
	}
	prev = `/api/users?page=1&page_size=100`
	if links.Prev() != prev {
		t.Errorf("unexpected prev: %s != %s", links.Prev(), prev)
	}
	next := `/api/users?page=3&page_size=100`
	if links.Next() != next {
		t.Errorf("unexpected prev: %s != %s", links.Next(), next)
	}
}

func TestTestTCPConn(t *testing.T) {
	server := httptest.NewServer(nil)
	defer server.Close()
	addr := strings.TrimPrefix(server.URL, "http://")
	if err := TestTCPConn(addr, 60, 2); err != nil {
		t.Fatalf("failed to test tcp connection of %s: %v", addr, err)
	}
}

func TestParseTimeStamp(t *testing.T) {
	// invalid input
	_, err := ParseTimeStamp("")
	assert.NotNil(t, err)

	// invalid input
	_, err = ParseTimeStamp("invalid")
	assert.NotNil(t, err)

	// valid
	now := time.Now().Unix()
	result, err := ParseTimeStamp(strconv.FormatInt(now, 10))
	assert.Nil(t, err)
	assert.Equal(t, now, result.Unix())
}

func TestParseHarborIDOrName(t *testing.T) {
	// nil input
	id, name, err := ParseProjectIDOrName(nil)
	assert.NotNil(t, err)

	// valid int ID
	id, name, err = ParseProjectIDOrName(1)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), id)
	assert.Equal(t, "", name)

	// valid int64 ID
	id, name, err = ParseProjectIDOrName(int64(1))
	assert.Nil(t, err)
	assert.Equal(t, int64(1), id)
	assert.Equal(t, "", name)

	// valid name
	id, name, err = ParseProjectIDOrName("project")
	assert.Nil(t, err)
	assert.Equal(t, int64(0), id)
	assert.Equal(t, "project", name)
}

type testingStruct struct {
	Name  string
	Count int
}

func TestConvertMapToStruct(t *testing.T) {
	dataMap := make(map[string]interface{})
	dataMap["Name"] = "testing"
	dataMap["Count"] = 100

	obj := &testingStruct{}
	if err := ConvertMapToStruct(obj, dataMap); err != nil {
		t.Fatal(err)
	} else {
		if obj.Name != "testing" || obj.Count != 100 {
			t.Fail()
		}
	}
}
