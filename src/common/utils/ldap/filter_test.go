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

package ldap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilterBuilder(t *testing.T) {
	fb, err := NewFilterBuilder("(objectclass=groupOfNames)")
	if err != nil {
		t.Error(err)
	}
	fb2, err := NewFilterBuilder("(cn=admin)")
	if err != nil {
		t.Error(err)
	}

	fb3 := fb.And(fb2)
	result, err := fb3.String()
	assert.Equal(t, result, "(&(objectclass=groupOfNames)(cn=admin))")
}

func TestFilterBuilder_And(t *testing.T) {
	type args struct {
		filterA string
		filterB string
	}
	cases := []struct {
		name      string
		in        args
		want      string
		wantError error
	}{
		{name: `normal`, in: args{"(objectclass=groupOfNames)", "(cn=admin)"}, want: "(&(objectclass=groupOfNames)(cn=admin))", wantError: nil},
		{name: `empty left`, in: args{"", "(cn=admin)"}, want: "(cn=admin))", wantError: nil},
		{name: `empty right`, in: args{"(cn=admin)", ""}, want: "(cn=admin))", wantError: nil},
		{name: `both empty`, in: args{"", ""}, want: "", wantError: nil},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			fbA, err := NewFilterBuilder(tt.in.filterA)
			if err != nil {
				t.Error(err)
			}
			fbB, err := NewFilterBuilder(tt.in.filterB)
			if err != nil {
				t.Error(fbB)
			}
			got, err := fbA.And(fbB).String()
			if got != tt.want && err != tt.wantError {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}

		})
	}
}

func TestFilterBuilder_Or(t *testing.T) {
	type args struct {
		filterA string
		filterB string
	}
	cases := []struct {
		name      string
		in        args
		want      string
		wantError error
	}{
		{name: `normal`, in: args{"(objectclass=groupOfNames)", "(cn=admin)"}, want: "(|(objectclass=groupOfNames)(cn=admin))", wantError: nil},
		{name: `empty left`, in: args{"", "(cn=admin)"}, want: "(cn=admin))", wantError: nil},
		{name: `empty right`, in: args{"(cn=admin)", ""}, want: "(cn=admin))", wantError: nil},
		{name: `both empty`, in: args{"", ""}, want: "", wantError: nil},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			fbA, err := NewFilterBuilder(tt.in.filterA)
			if err != nil {
				t.Error(err)
			}
			fbB, err := NewFilterBuilder(tt.in.filterB)
			if err != nil {
				t.Error(fbB)
			}
			got, err := fbA.Or(fbB).String()
			if got != tt.want && err != tt.wantError {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}

		})
	}
}
