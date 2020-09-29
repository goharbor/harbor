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

package orm

import (
	"reflect"
	"testing"
)

func Test_snakeCase(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"ProjectID", args{"ProjectID"}, "project_id"},
		{"project_id", args{"project_id"}, "project_id"},
		{"RepositoryName", args{"RepositoryName"}, "repository_name"},
		{"repository_name", args{"repository_name"}, "repository_name"},
		{"ProfileURL", args{"ProfileURL"}, "profile_url"},
		{"City", args{"City"}, "city"},
		{"Address1", args{"Address1"}, "address1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := snakeCase(tt.args.str); got != tt.want {
				t.Errorf("snakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

type Bar struct {
	Field1    string `orm:"-"`
	Field2    string `orm:"column(customized_field2)"`
	Field3    string
	FirstName string
}

func (Bar) Foo() {}

func (bar *Bar) FilterBy() {}

func (bar *Bar) FilterByField4() {}

func Test_queriableColumns(t *testing.T) {
	toWant := func(fields ...string) map[string]bool {
		want := map[string]bool{}

		for _, field := range fields {
			want[field] = true
		}

		return want
	}

	type args struct {
		model interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]bool
	}{
		{"bar", args{&Bar{}}, toWant("Field2", "customized_field2", "Field3", "field3", "FirstName", "first_name")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queriableColumns(tt.args.model); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queriableColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_queriableMethods(t *testing.T) {
	type args struct {
		model interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"bar", args{&Bar{}}, map[string]string{"field4": "FilterByField4"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queriableMethods(tt.args.model); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queriableMethods() = %v, want %v", got, tt.want)
			}
		})
	}
}
