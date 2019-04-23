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
			pattern: "a[",
			str:     "aaa",
			match:   false,
		},
	}
	for _, c := range cases {
		match, err := Match(c.pattern, c.str)
		require.Nil(t, err)
		assert.Equal(t, c.match, match)
	}
}

func TestGetHTTPTransport(t *testing.T) {
	transport := GetHTTPTransport(true)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	transport = GetHTTPTransport(false)
	assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestParseRepository(t *testing.T) {
	// empty repository
	repository := ""
	namespace, rest := ParseRepository(repository)
	assert.Equal(t, "", namespace)
	assert.Equal(t, "", rest)
	// repository contains no "/"
	repository = "c"
	namespace, rest = ParseRepository(repository)
	assert.Equal(t, "", namespace)
	assert.Equal(t, "c", rest)
	// repository contains only one "/"
	repository = "b/c"
	namespace, rest = ParseRepository(repository)
	assert.Equal(t, "b", namespace)
	assert.Equal(t, "c", rest)
	// repository contains more than one "/"
	repository = "a/b/c"
	namespace, rest = ParseRepository(repository)
	assert.Equal(t, "a/b", namespace)
	assert.Equal(t, "c", rest)
}
func TestIsSpecificRepositoryName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Is Specific", "a", true},
		{"Is Specific", "abc", true},
		{"Is Specific", "a/b", true},
		{"Is Specific", "a/b/c", true},
		{"Not Specific", "*", false},
		{"Not Specific", "?", false},
		{"Not Specific", "*c", false},
		{"Not Specific", "a*", false},
		{"Not Specific", "a*b*c*d*e*", false},
		{"Not Specific", "a*b?c*x", false},
		{"Not Specific", "ab[c]", false},
		{"Not Specific", "ab[b-d]", false},
		{"Not Specific", "ab[e-g]", false},
		{"Not Specific", "ab[^c]", false},
		{"Not Specific", "ab[^b-d]", false},
		{"Not Specific", "ab[^e-g]", false},
		{"Not Specific", "a\\*b", false},
		{"Not Specific", "a?b", false},
		{"Not Specific", "a[^a]b", false},
		{"Not Specific", "a???b", false},
		{"Not Specific", "a[^a][^a][^a]b", false},
		{"Not Specific", "[a-ζ]*", false},
		{"Not Specific", "*[a-ζ]", false},
		{"Not Specific", "a?b", false},
		{"Not Specific", "a*b", false},
		{"Not Specific", "[\\-]", false},
		{"Not Specific", "[x\\-]", false},
		{"Not Specific", "[x\\-]", false},
		{"Not Specific", "[x\\-]", false},
		{"Not Specific", "[\\-x]", false},
		{"Not Specific", "[\\-x]", false},
		{"Not Specific", "[\\-x]", false},
		{"Not Specific", "[a-b-c]", false},
		{"Not Specific", "*x", false},
		{"Not Specific", "[abc]", false},
		{"Not Specific", "**", false},
		{"Not Specific", "ab{c,d}", false},
		{"Not Specific", "ab{c,d,*}", false},
		{"Not Specific", "abc**", false},
		{"Not Specific", "[]a]", false},
		{"Not Specific", "[-]", false},
		{"Not Specific", "[x-]", false},
		{"Not Specific", "[-x]", false},
		{"Not Specific", "\\", false},
		{"Not Specific", "[a-b-c]", false},
		{"Not Specific", "[]", false},
		{"Not Specific", "[", false},
		{"Not Specific", "[^", false},
		{"Not Specific", "^", false},
		{"Not Specific", "]", false},
		{"Not Specific", "[^bc", false},
		{"Not Specific", "a[", false},
		{"Not Specific", "ab{c,d}[", false},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			var got = IsSpecificRepositoryName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
