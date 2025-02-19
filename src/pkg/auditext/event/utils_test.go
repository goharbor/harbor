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

package event

import (
	"testing"
)

func TestRedact(t *testing.T) {
	type args struct {
		payload             string
		sensitiveAttributes []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal case", args{`{"password":"123456"}`, []string{"password"}}, "{\n  \"password\": \"***\"\n}"},
		{"no sensitive case", args{`{"ldap_base_dn":"dc=example,dc=com"}`, []string{"password"}}, "{\n  \"ldap_base_dn\": \"dc=example,dc=com\"\n}"},
		{"empty case", args{"", []string{"password"}}, ""},
		{"mixed attribute", args{`{"ldap_base_dn":"dc=example,dc=com", "ldap_search_passwd": "admin"}`, []string{"ldap_search_passwd"}}, "{\n  \"ldap_base_dn\": \"dc=example,dc=com\",\n  \"ldap_search_passwd\": \"***\"\n}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Redact(tt.args.payload, tt.args.sensitiveAttributes); got != tt.want {
				t.Errorf("Redact() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_replacePassword(t *testing.T) {
	type args struct {
		data           map[string]interface{}
		maskAttributes []string
	}
	tests := []struct {
		name string
		args args
	}{
		{"normal case", args{map[string]interface{}{"password": "123456"}, []string{"password"}}},
		{"nested case", args{map[string]interface{}{"master": map[string]interface{}{"password": "123456"}}, []string{"password"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replacePassword(tt.args.data, tt.args.maskAttributes)
			if val, ok := tt.args.data["password"]; ok && val != nil && tt.args.data["password"] != "***" {
				t.Errorf("replacePassword() = %v, want %v", tt.args.data["password"], "***")
			}
			if val, ok := tt.args.data["master"]; ok && val != nil && tt.args.data["master"].(map[string]interface{})["password"] != "***" {
				t.Errorf("replacePassword() = %v, want %v", tt.args.data["master"].(map[string]interface{})["password"], "***")
			}
		})
	}
}
