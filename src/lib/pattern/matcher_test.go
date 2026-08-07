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

package pattern

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRepositoryFilter(t *testing.T) {
	cases := []struct {
		name          string
		filterPattern string
		kind          string
		expectErr     bool
	}{
		{
			name:          "empty pattern",
			filterPattern: "",
			kind:          KindRegex,
			expectErr:     false,
		},
		{
			name:          "whitespace pattern",
			filterPattern: "   ",
			kind:          KindDoublestar,
			expectErr:     false,
		},
		{
			name:          "valid regex pattern",
			filterPattern: "library/.*",
			kind:          KindRegex,
			expectErr:     false,
		},
		{
			name:          "invalid regex pattern",
			filterPattern: "[a-z",
			kind:          KindRegex,
			expectErr:     true,
		},
		{
			name:          "valid doublestar pattern",
			filterPattern: "library/**",
			kind:          KindDoublestar,
			expectErr:     false,
		},
		{
			name:          "invalid doublestar pattern",
			filterPattern: "library/[a-z",
			kind:          KindDoublestar,
			expectErr:     true,
		},
		{
			name:          "valid doublestar pattern with empty kind",
			filterPattern: "library/**",
			kind:          "",
			expectErr:     false,
		},
		{
			name:          "invalid doublestar pattern with empty kind",
			filterPattern: "library/[a-z",
			kind:          "",
			expectErr:     true,
		},
		{
			name:          "unsupported kind",
			filterPattern: "library/**",
			kind:          "invalid_kind",
			expectErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRepositoryFilter(tc.filterPattern, tc.kind)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	cases := []struct {
		name          string
		value         string
		filterPattern string
		kind          string
		want          bool
		expectErr     bool
	}{
		{
			name:          "empty pattern matches all",
			value:         "library/nginx",
			filterPattern: "",
			kind:          KindRegex,
			want:          true,
		},
		{
			name:          "whitespace pattern matches all",
			value:         "library/nginx",
			filterPattern: "   ",
			kind:          KindDoublestar,
			want:          true,
		},
		{
			name:          "regex exact match",
			value:         "library/nginx",
			filterPattern: "library/nginx",
			kind:          KindRegex,
			want:          true,
		},
		{
			name:          "regex partial match is anchored",
			value:         "library/nginx",
			filterPattern: "nginx",
			kind:          KindRegex,
			want:          false,
		},
		{
			name:          "regex prefix match is anchored",
			value:         "library/nginx",
			filterPattern: "library/",
			kind:          KindRegex,
			want:          false,
		},
		{
			name:          "regex wildcard match",
			value:         "library/nginx",
			filterPattern: "library/.*",
			kind:          KindRegex,
			want:          true,
		},
		{
			name:          "regex invalid pattern returns error",
			value:         "library/nginx",
			filterPattern: "[a-z",
			kind:          KindRegex,
			expectErr:     true,
		},
		{
			name:          "doublestar exact match",
			value:         "library/nginx",
			filterPattern: "library/nginx",
			kind:          KindDoublestar,
			want:          true,
		},
		{
			name:          "doublestar single star",
			value:         "library/nginx",
			filterPattern: "library/*",
			kind:          KindDoublestar,
			want:          true,
		},
		{
			name:          "doublestar double star",
			value:         "org/team/repo",
			filterPattern: "org/**",
			kind:          KindDoublestar,
			want:          true,
		},
		{
			name:          "doublestar match all",
			value:         "library/nginx",
			filterPattern: "**",
			kind:          KindDoublestar,
			want:          true,
		},
		{
			name:          "doublestar no match",
			value:         "library/nginx",
			filterPattern: "other/*",
			kind:          KindDoublestar,
			want:          false,
		},
		{
			name:          "doublestar alternation",
			value:         "library/nginx",
			filterPattern: "library/{nginx,alpine}",
			kind:          KindDoublestar,
			want:          true,
		},
		{
			name:          "empty kind defaults to doublestar",
			value:         "library/nginx",
			filterPattern: "library/**",
			kind:          "",
			want:          true,
		},
		{
			name:          "unsupported kind returns error",
			value:         "library/nginx",
			filterPattern: "library/**",
			kind:          "invalid_kind",
			expectErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Match(tc.value, tc.filterPattern, tc.kind)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
