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

package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	cases := []struct {
		pattern string
		str     string
		match   bool
	}{
		{
			pattern: "",
			str:     "library",
			match:   true,
		},
		{
			pattern: "",
			str:     "",
			match:   true,
		},
		{
			pattern: "*",
			str:     "library",
			match:   true,
		},
		{
			pattern: "*",
			str:     "library/hello-world",
			match:   false,
		},
		{
			pattern: "**",
			str:     "library/hello-world",
			match:   true,
		},
		{
			pattern: "{library,harbor}/**",
			str:     "library/hello-world",
			match:   true,
		},
		{
			pattern: "{library,harbor}/**",
			str:     "harbor/hello-world",
			match:   true,
		},
		{
			pattern: "1.?",
			str:     "1.0",
			match:   true,
		},
		{
			pattern: "1.?",
			str:     "1.01",
			match:   false,
		},
		{
			pattern: "v2.[4-6].*", // match v2.4.*~v2.7.* version
			str:     "v2.4.0",
			match:   true,
		},
		{
			pattern: "v2.[4-7].*", // match v2.4.*~v2.7.* version
			str:     "v2.7.0",
			match:   true,
		},
		// New regex-based tests
		{
			pattern: "regex:^v1\\.[0-9]+$",
			str:     "v1.0",
			match:   true,
		},
		{
			pattern: "regex:^v1\\.[0-9]+$",
			str:     "v1.10",
			match:   true,
		},
		{
			pattern: "regex:^v1\\.[0-9]+$",
			str:     "v2.0",
			match:   false,
		},
		{
			pattern: "regex:^feature/.+$",
			str:     "feature/abc",
			match:   true,
		},
		{
			pattern: "regex:^feature/.+$",
			str:     "bugfix/abc",
			match:   false,
		},
		{
			pattern: "regex:^(v|release)-\\d+\\.\\d+(\\.\\d+)?$",
			str:     "v-1.2.3",
			match:   true,
		},
		{
			pattern: "regex:^(v|release)-\\d+\\.\\d+(\\.\\d+)?$",
			str:     "release-2.0",
			match:   true,
		},
		{
			pattern: "regex:^(v|release)-\\d+\\.\\d+(\\.\\d+)?$",
			str:     "v-2",
			match:   false,
		},
		{
			pattern: "regex:^hotfix-(issue|bug)-[0-9]{4,}$",
			str:     "hotfix-bug-1234",
			match:   true,
		},
		{
			pattern: "regex:^hotfix-(issue|bug)-[0-9]{4,}$",
			str:     "hotfix-feature-1234",
			match:   false,
		},
	}
	for i, c := range cases {
		fmt.Printf("running case %d ...\n", i)
		match, err := Match(c.pattern, c.str)
		require.Nil(t, err, "unexpected error for pattern: %s", c.pattern)
		assert.Equal(t, c.match, match, "pattern: %s, str: %s", c.pattern, c.str)
	}
}

func TestIsSpecificPathComponent(t *testing.T) {
	cases := []struct {
		component        string
		isSpecific       bool
		resultComponents []string
	}{
		{
			component:        "",
			isSpecific:       false,
			resultComponents: []string{},
		},
		{
			component:        "library/hello-world",
			isSpecific:       false,
			resultComponents: []string{},
		},
		{
			component:        "library",
			isSpecific:       true,
			resultComponents: []string{"library"},
		},
		{
			component:        "lib*",
			isSpecific:       false,
			resultComponents: []string{},
		},
		{
			component:        "{library}",
			isSpecific:       true,
			resultComponents: []string{"library"},
		},
		{
			component:        "{library,test}",
			isSpecific:       true,
			resultComponents: []string{"library", "test"},
		},
		{
			component:        "{library{a}c}",
			isSpecific:       false,
			resultComponents: []string{},
		},
	}
	for i, c := range cases {
		fmt.Printf("running case %d ...\n", i)
		components, ok := IsSpecificPathComponent(c.component)
		require.Equal(t, c.isSpecific, ok)
		require.Equal(t, len(c.resultComponents), len(components))
		for i := range components {
			assert.Equal(t, c.resultComponents[i], components[i])
		}
	}
}

func TestMatch_InvalidRegex(t *testing.T) {
	invalidPatterns := []string{
		"regex:^v[0-9+$",     // missing closing bracket
		"regex:(abc",         // unclosed group
		"regex:*abc",         // invalid quantifier at start
		"regex:[a-z",         // incomplete character class
	}

	for i, pattern := range invalidPatterns {
		fmt.Printf("[TestMatch_InvalidRegex case %d] pattern=%q\n", i, pattern)
		_, err := Match(pattern, "test-tag")
		assert.Error(t, err, "expected error for invalid pattern: %s", pattern)
	}
}

func TestIsSpecificPath(t *testing.T) {
	cases := []struct {
		path        string
		isSpecific  bool
		resultPaths []string
	}{
		{
			path:        "",
			isSpecific:  false,
			resultPaths: []string{},
		},
		{
			path:        "library",
			isSpecific:  true,
			resultPaths: []string{"library"},
		},
		{
			path:        "library/hello-world",
			isSpecific:  true,
			resultPaths: []string{"library/hello-world"},
		},
		{
			path:        "library/**",
			isSpecific:  false,
			resultPaths: []string{},
		},
		{
			path:        "{library}",
			isSpecific:  true,
			resultPaths: []string{"library"},
		},
		{
			path:        "library/{hello-world,busybox}",
			isSpecific:  true,
			resultPaths: []string{"library/hello-world", "library/busybox"},
		},
	}
	for i, c := range cases {
		fmt.Printf("running case %d ...\n", i)
		paths, ok := IsSpecificPath(c.path)
		require.Equal(t, c.isSpecific, ok)
		require.Equal(t, len(c.resultPaths), len(paths))
		for i := range paths {
			assert.Equal(t, c.resultPaths[i], paths[i])
		}
	}
}
