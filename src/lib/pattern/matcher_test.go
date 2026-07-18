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
	"github.com/stretchr/testify/require"
)

func TestNewRepositoryFilter(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		kind    string
		want    *RepositoryFilter
	}{
		{
			name:    "empty pattern and kind",
			pattern: "",
			kind:    "",
			want:    &RepositoryFilter{Filter: "", Kind: ""},
		},
		{
			name:    "regex filter",
			pattern: "^library/.*",
			kind:    KindRegex,
			want:    &RepositoryFilter{Filter: "^library/.*", Kind: KindRegex},
		},
		{
			name:    "doublestar filter",
			pattern: "library/**",
			kind:    KindDoublestar,
			want:    &RepositoryFilter{Filter: "library/**", Kind: KindDoublestar},
		},
		{
			name:    "pattern with whitespace is trimmed",
			pattern: "  nginx  ",
			kind:    "",
			want:    &RepositoryFilter{Filter: "nginx", Kind: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRepositoryFilter(tt.pattern, tt.kind)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRepositoryFilter_Match(t *testing.T) {
	tests := []struct {
		name    string
		rf      *RepositoryFilter
		value   string
		want    bool
		wantErr bool
	}{
		{
			name:  "nil filter matches all",
			rf:    nil,
			value: "library/nginx",
			want:  true,
		},
		{
			name:  "empty filter matches all",
			rf:    &RepositoryFilter{},
			value: "library/nginx",
			want:  true,
		},
		{
			name:  "regex prefix is anchored",
			rf:    &RepositoryFilter{Filter: "^library/", Kind: KindRegex},
			value: "library/nginx",
			want:  false,
		},
		{
			name:  "regex wildcard match",
			rf:    &RepositoryFilter{Filter: "^library/.*", Kind: KindRegex},
			value: "library/nginx",
			want:  true,
		},
		{
			name:  "regex no match",
			rf:    &RepositoryFilter{Filter: "^other/", Kind: KindRegex},
			value: "library/nginx",
			want:  false,
		},
		{
			name:  "doublestar match",
			rf:    &RepositoryFilter{Filter: "library/**", Kind: KindDoublestar},
			value: "library/nginx",
			want:  true,
		},
		{
			name:  "doublestar no match",
			rf:    &RepositoryFilter{Filter: "other/**", Kind: KindDoublestar},
			value: "library/nginx",
			want:  false,
		},
		{
			name:  "empty kind defaults to doublestar",
			rf:    &RepositoryFilter{Filter: "nginx"},
			value: "nginx",
			want:  true,
		},
		{
			name:    "invalid regex",
			rf:      &RepositoryFilter{Filter: "[invalid", Kind: KindRegex},
			value:   "library/nginx",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rf.Match(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegexMatcher(t *testing.T) {
	m := &RegexMatcher{}

	tests := []struct {
		name    string
		value   string
		pattern string
		want    bool
		wantErr bool
	}{
		{
			name:    "empty pattern matches all",
			value:   "library/nginx",
			pattern: "",
			want:    true,
		},
		{
			name:    "whitespace pattern matches all",
			value:   "library/nginx",
			pattern: "   ",
			want:    true,
		},
		{
			name:    "exact match",
			value:   "library/nginx",
			pattern: "^library/nginx$",
			want:    true,
		},
		{
			name:    "partial match is anchored",
			value:   "library/nginx",
			pattern: "nginx",
			want:    false,
		},
		{
			name:    "prefix match is anchored",
			value:   "library/nginx",
			pattern: "^library/",
			want:    false,
		},
		{
			name:    "no match",
			value:   "library/nginx",
			pattern: "^other/",
			want:    false,
		},
		{
			name:    "wildcard match",
			value:   "library/nginx",
			pattern: "library/.*",
			want:    true,
		},
		{
			name:    "invalid regex returns error",
			value:   "library/nginx",
			pattern: "[invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.Match(tt.value, tt.pattern)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDoublestarMatcher(t *testing.T) {
	m := &DoublestarMatcher{}

	tests := []struct {
		name    string
		value   string
		pattern string
		want    bool
		wantErr bool
	}{
		{
			name:    "empty pattern matches all",
			value:   "library/nginx",
			pattern: "",
			want:    true,
		},
		{
			name:    "whitespace pattern matches all",
			value:   "library/nginx",
			pattern: "   ",
			want:    true,
		},
		{
			name:    "exact match",
			value:   "library/nginx",
			pattern: "library/nginx",
			want:    true,
		},
		{
			name:    "single star matches within path segment",
			value:   "library/nginx",
			pattern: "library/*",
			want:    true,
		},
		{
			name:    "double star matches across path segments",
			value:   "org/team/repo",
			pattern: "org/**",
			want:    true,
		},
		{
			name:    "double star matches everything",
			value:   "library/nginx",
			pattern: "**",
			want:    true,
		},
		{
			name:    "double star in middle",
			value:   "a/b/c/d",
			pattern: "a/**/d",
			want:    true,
		},
		{
			name:    "no match",
			value:   "library/nginx",
			pattern: "other/*",
			want:    false,
		},
		{
			name:    "question mark wildcard",
			value:   "library/nginx1",
			pattern: "library/nginx?",
			want:    true,
		},
		{
			name:    "character class",
			value:   "library/nginx1",
			pattern: "library/nginx[0-9]",
			want:    true,
		},
		{
			name:    "alternation",
			value:   "library/nginx",
			pattern: "library/{nginx,alpine}",
			want:    true,
		},
		{
			name:    "invalid pattern returns error",
			value:   "library/nginx",
			pattern: "[invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.Match(tt.value, tt.pattern)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMatcher(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want Matcher
	}{
		{
			name: "regex kind",
			kind: KindRegex,
			want: &RegexMatcher{},
		},
		{
			name: "doublestar kind",
			kind: KindDoublestar,
			want: &DoublestarMatcher{},
		},
		{
			name: "empty kind defaults to doublestar",
			kind: "",
			want: &DoublestarMatcher{},
		},
		{
			name: "unknown kind defaults to doublestar",
			kind: "unknown",
			want: &DoublestarMatcher{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMatcher(tt.kind)
			assert.IsType(t, tt.want, got)
		})
	}
}
