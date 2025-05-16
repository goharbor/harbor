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

package utils

import (
	"encoding/base64"
	"net/http/httptest"
	"reflect"
	"sort"
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
	tests := map[string]struct {
		content string
		salt    string
		alg     string
		want    string
	}{
		"sha1 test":   {content: "content", salt: "salt", alg: SHA1, want: "dc79e76c88415c97eb089d9cc80b4ab0"},
		"sha256 test": {content: "content", salt: "salt", alg: SHA256, want: "83d3d6f3e7cacb040423adf7ced63d21"},
	}

	for name, tc := range tests {
		got := Encrypt(tc.content, tc.salt, tc.alg)
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("%s: expected: %v, got: %v", name, tc.want, got)
		}
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
	// Test b64 for backward compatibility
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

func TestGenerateRandomStringWithLen(t *testing.T) {
	str := GenerateRandomStringWithLen(16)
	if len(str) != 16 {
		t.Errorf("Failed to generate ramdom string with fixed length.")
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
	dataMap := make(map[string]any)
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

func TestSafeCastString(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nil value", args{nil}, ""},
		{"normal string", args{"sample"}, "sample"},
		{"wrong type", args{12}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeCastString(tt.args.value); got != tt.want {
				t.Errorf("SafeCastString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeCastBool(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nil value", args{nil}, false},
		{"normal bool", args{true}, true},
		{"wrong type", args{"true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeCastBool(tt.args.value); got != tt.want {
				t.Errorf("SafeCastBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeCastInt(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"nil value", args{nil}, 0},
		{"normal int", args{1234}, 1234},
		{"wrong type", args{"sample"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeCastInt(tt.args.value); got != tt.want {
				t.Errorf("SafeCastInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeCastFloat64(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"nil value", args{nil}, 0},
		{"normal float64", args{12.34}, 12.34},
		{"wrong type", args{false}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeCastFloat64(tt.args.value); got != tt.want {
				t.Errorf("SafeCastFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimLower(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{" CN=example,DC=test,DC=com "}, "cn=example,dc=test,dc=com"},
		{"empty", args{" "}, ""},
		{"empty2", args{""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimLower(tt.args.str); got != tt.want {
				t.Errorf("TrimLower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStrValueOfAnyType(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"float", args{float32(1048576.1)}, "1048576.1"},
		{"float", args{float64(1048576.12)}, "1048576.12"},
		{"float", args{1048576.000}, "1048576"},
		{"int", args{1048576}, "1048576"},
		{"int", args{9223372036854775807}, "9223372036854775807"},
		{"string", args{"hello world"}, "hello world"},
		{"bool", args{true}, "true"},
		{"bool", args{false}, "false"},
		{"map", args{map[string]any{"key1": "value1"}}, "{\"key1\":\"value1\"}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStrValueOfAnyType(tt.args.value); got != tt.want {
				t.Errorf("GetStrValueOfAnyType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextSchedule(t *testing.T) {
	curTime := time.Date(2009, time.November, 10, 20, 0, 0, 0, time.UTC)
	expectTime1 := time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC)
	expectTime2 := time.Date(2009, time.November, 10, 21, 0, 0, 0, time.UTC)
	expectTime4 := time.Date(2009, time.November, 10, 20, 8, 0, 0, time.UTC)

	type args struct {
		cron    string
		curTime time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"daily_in_six", args{"0 0 0 * * *", curTime}, expectTime1},
		{"hourly_in_six", args{"0 0 * * * *", curTime}, expectTime2},
		{"custom", args{"0 8 20 * * *", curTime}, expectTime4},
		{"zero time", args{"", curTime}, time.Time{}},
		{"invalid cron", args{"2", curTime}, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NextSchedule(tt.args.cron, tt.args.curTime), "NextSchedule(%v, %v)", tt.args.cron, tt.args.curTime)
		})
	}
}

type UserGroupSearchItem struct {
	GroupName string
}

func Test_sortMostMatch(t *testing.T) {
	type args struct {
		input     []*UserGroupSearchItem
		matchWord string
		expected  []*UserGroupSearchItem
	}
	tests := []struct {
		name string
		args args
	}{
		{"normal", args{[]*UserGroupSearchItem{
			{GroupName: "user"}, {GroupName: "harbor_user"}, {GroupName: "admin_user"}, {GroupName: "users"},
		}, "user", []*UserGroupSearchItem{
			{GroupName: "user"}, {GroupName: "users"}, {GroupName: "admin_user"}, {GroupName: "harbor_user"},
		}}},
		{"duplicate_item", args{[]*UserGroupSearchItem{
			{GroupName: "user"}, {GroupName: "user"}, {GroupName: "harbor_user"}, {GroupName: "admin_user"}, {GroupName: "users"},
		}, "user", []*UserGroupSearchItem{
			{GroupName: "user"}, {GroupName: "user"}, {GroupName: "users"}, {GroupName: "admin_user"}, {GroupName: "harbor_user"},
		}}},
		{"miss_exact_match", args{[]*UserGroupSearchItem{
			{GroupName: "harbor_user"}, {GroupName: "admin_user"}, {GroupName: "users"},
		}, "user", []*UserGroupSearchItem{
			{GroupName: "users"}, {GroupName: "admin_user"}, {GroupName: "harbor_user"},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sort.Slice(tt.args.input, func(i, j int) bool {
				return MostMatchSorter(tt.args.input[i].GroupName, tt.args.input[j].GroupName, tt.args.matchWord)
			})
			assert.True(t, reflect.DeepEqual(tt.args.input, tt.args.expected))
		})
	}
}

func TestValidateCronString(t *testing.T) {
	testCases := []struct {
		description string
		input       string
		hasErr      bool
	}{
		// empty cron string
		{
			description: "test case 1",
			input:       "",
			hasErr:      true,
		},

		// invalid cron format
		{
			description: "test case 2",
			input:       "0 2 3",
			hasErr:      true,
		},

		// the 1st field (indicating Seconds of time) of the cron setting must be 0
		{
			description: "test case 3",
			input:       "1 0 0 1 1 0",
			hasErr:      true,
		},

		// valid cron string
		{
			description: "test case 4",
			input:       "0 1 2 1 1 *",
			hasErr:      false,
		},
	}

	for _, tc := range testCases {
		err := ValidateCronString(tc.input)
		if tc.hasErr {
			if err == nil {
				t.Errorf("%s, expect having error, while actual error returned is nil", tc.description)
			}
		} else {
			// tc.hasErr == false
			if err != nil {
				t.Errorf("%s, expect having no error, while actual error returned is not nil, err=%v", tc.description, err)
			}
		}
	}
}

func TestIsLocalPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"normal test", args{"/harbor/project"}, true},
		{"failed", args{"www.myexample.com"}, false},
		{"other_site1", args{"//www.myexample.com"}, false},
		{"other_site2", args{"https://www.myexample.com"}, false},
		{"other_site", args{"http://www.myexample.com"}, false},
		{"empty_path", args{""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsLocalPath(tt.args.path), "IsLocalPath(%v)", tt.args.path)
		})
	}
}
