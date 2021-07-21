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
	}
	for _, c := range cases {
		match, err := Match(c.pattern, c.str)
		require.Nil(t, err)
		assert.Equal(t, c.match, match)
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
