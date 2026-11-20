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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadataValidate(t *testing.T) {
	tag := func(kind, decoration, pattern string) []*Selector {
		return []*Selector{{Kind: kind, Decoration: decoration, Pattern: pattern}}
	}
	repo := func(kind, decoration, pattern string) map[string][]*Selector {
		return map[string][]*Selector{"repository": {{Kind: kind, Decoration: decoration, Pattern: pattern}}}
	}

	cases := []struct {
		name    string
		rule    *Metadata
		wantErr string
	}{
		{
			name: "valid rule",
			rule: &Metadata{
				TagSelectors:   tag("doublestar", "matches", "release-**"),
				ScopeSelectors: repo("doublestar", "repoMatches", "{redis,mysql}"),
			},
		},
		{
			name: "rule without selectors",
			rule: &Metadata{},
		},
		{
			name: "valid pattern with a character class",
			rule: &Metadata{
				TagSelectors:   tag("doublestar", "excludes", "v[0-9]*"),
				ScopeSelectors: repo("doublestar", "repoExcludes", "library/**"),
			},
		},
		{
			name: "alternatives may contain a slash",
			rule: &Metadata{
				TagSelectors:   tag("doublestar", "matches", "{v1/a,v2}"),
				ScopeSelectors: repo("doublestar", "repoMatches", "{redis,team/app}"),
			},
		},
		{
			name: "empty pattern matches everything",
			rule: &Metadata{
				TagSelectors:   tag("doublestar", "matches", ""),
				ScopeSelectors: repo("doublestar", "repoMatches", ""),
			},
		},
		{
			name: "nil selector is skipped",
			rule: &Metadata{
				TagSelectors:   []*Selector{nil},
				ScopeSelectors: map[string][]*Selector{"repository": {nil}},
			},
		},
		{
			name:    "malformed tag pattern",
			rule:    &Metadata{TagSelectors: tag("doublestar", "matches", "[release")},
			wantErr: "invalid tag selector: pattern \"[release\": syntax error in pattern",
		},
		{
			name:    "malformed repository pattern",
			rule:    &Metadata{ScopeSelectors: repo("doublestar", "repoMatches", "{a,b")},
			wantErr: "invalid repository scope selector: pattern \"{a,b\": syntax error in pattern",
		},
		{
			name:    "trailing backslash",
			rule:    &Metadata{TagSelectors: tag("doublestar", "matches", "release\\")},
			wantErr: "syntax error in pattern",
		},
		{
			name:    "unregistered kind",
			rule:    &Metadata{TagSelectors: tag("regex", "matches", "release-.*")},
			wantErr: "invalid tag selector: selector regex is not registered",
		},
		{
			name:    "unsupported decoration",
			rule:    &Metadata{ScopeSelectors: repo("doublestar", "equals", "redis")},
			wantErr: "invalid repository scope selector: decoration equals of selector doublestar is not supported",
		},
		{
			name:    "empty kind and decoration",
			rule:    &Metadata{TagSelectors: tag("", "", "release-*")},
			wantErr: "empty selector kind or decoration",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.rule.Validate()
			if c.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), c.wantErr)
			}
		})
	}
}
