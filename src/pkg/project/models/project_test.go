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

package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeProxyCacheBasePath(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "empty", in: "", want: ""},
		{name: "only slashes", in: "///", want: ""},
		{name: "single segment", in: "dev", want: "dev"},
		{name: "multi segment", in: "dev/team", want: "dev/team"},
		{name: "leading slash", in: "/dev", want: "dev"},
		{name: "trailing slash", in: "dev/", want: "dev"},
		{name: "surrounding slashes", in: "/dev/team/", want: "dev/team"},
		{name: "double underscore separator", in: "a__b", want: "a__b"},
		{name: "dash separators", in: "a---b", want: "a---b"},
		{name: "dot and digits", in: "dev.1/team_2", want: "dev.1/team_2"},
		{name: "uppercase", in: "Dev", wantErr: true},
		{name: "empty segment", in: "dev//team", wantErr: true},
		{name: "relative segment", in: "dev/../team", wantErr: true},
		{name: "parent segment", in: "../dev", wantErr: true},
		{name: "leading separator", in: "-dev", wantErr: true},
		{name: "trailing separator", in: "dev-", wantErr: true},
		// a dotted first segment is a legal repository path component, it is never
		// interpreted as a registry host because the base path only namespaces the
		// repository inside the upstream registry
		{name: "dotted first segment", in: "docker.io/dev", want: "docker.io/dev"},
		{name: "registry domain with port", in: "localhost:5000/dev", wantErr: true},
		{name: "whitespace", in: "de v", wantErr: true},
		{name: "too long", in: strings.Repeat("a", 256), wantErr: true},
		{name: "at the length limit", in: strings.Repeat("a", 255), want: strings.Repeat("a", 255)},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeProxyCacheBasePath(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProxyCacheBasePath(t *testing.T) {
	cases := []struct {
		name string
		pro  *Project
		want string
	}{
		{name: "nil metadata", pro: &Project{}, want: ""},
		{name: "key not set", pro: &Project{Metadata: map[string]string{ProMetaPublic: "true"}}, want: ""},
		{name: "empty value", pro: &Project{Metadata: map[string]string{ProMetaProxyCacheBasePath: ""}}, want: ""},
		{name: "single segment", pro: &Project{Metadata: map[string]string{ProMetaProxyCacheBasePath: "dev"}}, want: "dev"},
		{name: "multi segment", pro: &Project{Metadata: map[string]string{ProMetaProxyCacheBasePath: "dev/team"}}, want: "dev/team"},
		{name: "trimmed on read", pro: &Project{Metadata: map[string]string{ProMetaProxyCacheBasePath: "/dev/"}}, want: "dev"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.pro.ProxyCacheBasePath())
		})
	}
}
