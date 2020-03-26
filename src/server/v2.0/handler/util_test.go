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

package handler

import (
	"testing"
)

func Test_unescapePathParams(t *testing.T) {
	type Params struct {
		ProjectName    string
		RepositoryName string
	}

	str := "params"

	type args struct {
		params     interface{}
		fieldNames []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"non ptr", args{str, []string{"RepositoryName"}}, true},
		{"non struct", args{&str, []string{"RepositoryName"}}, true},
		{"ptr of struct", args{&Params{}, []string{"RepositoryName"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unescapePathParams(tt.args.params, tt.args.fieldNames...); (err != nil) != tt.wantErr {
				t.Errorf("unescapePathParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("ok", func(t *testing.T) {
		params := Params{ProjectName: "library", RepositoryName: "hello%2Fworld"}
		unescapePathParams(&params, "RepositoryName")
		if params.RepositoryName != "hello/world" {
			t.Errorf("unescapePathParams() not unescape RepositoryName field")
		}
	})
}
