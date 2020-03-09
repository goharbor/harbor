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

package q

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestParseFuzzyMatchValue(t *testing.T) {
	// empty string
	value := ""
	v, err := parseFuzzyMatchValue(value)
	require.NotNil(t, err)

	// contain no other characters except "~"
	value = "~"
	v, err = parseFuzzyMatchValue(value)
	require.NotNil(t, err)

	// valid fuzzy match value
	value = "~a"
	v, err = parseFuzzyMatchValue(value)
	require.Nil(t, err)
	assert.Equal(t, "a", v.Value)
}

func TestParseRange(t *testing.T) {
	// contains only "["
	value := "["
	v, err := parseRange(value)
	require.NotNil(t, err)

	// contains no ","
	value = "[]"
	v, err = parseRange(value)
	require.NotNil(t, err)

	// contains no other character
	value = "[~]"
	v, err = parseRange(value)
	require.NotNil(t, err)

	// contains multiple "~"
	value = "[~~]"
	v, err = parseRange(value)
	require.NotNil(t, err)

	// contains multiple char
	value = "[1~2~3]"
	v, err = parseRange(value)
	require.NotNil(t, err)

	// valid value
	value = "[1~]"
	v, err = parseRange(value)
	require.Nil(t, err)
	assert.Equal(t, int64(1), v.Min.(int64))
	assert.Nil(t, v.Max)

	// valid value
	value = "[~2]"
	v, err = parseRange(value)
	assert.Equal(t, int64(2), v.Max.(int64))
	assert.Nil(t, v.Min)

	// valid value
	value = "[1~2]"
	v, err = parseRange(value)
	require.Nil(t, err)
	assert.Equal(t, int64(1), v.Min.(int64))
	assert.Equal(t, int64(2), v.Max.(int64))
}

func TestParseOrList(t *testing.T) {
	// invalid
	value := "}{"
	v, err := parseOrList(value)
	require.NotNil(t, err)

	// valid value, contains no element
	value = "{}"
	v, err = parseOrList(value)
	require.Nil(t, err)
	assert.Len(t, v.Values, 0)

	// valid value, contains only one element
	value = "{1}"
	v, err = parseOrList(value)
	require.Nil(t, err)
	require.Len(t, v.Values, 1)
	assert.Equal(t, int64(1), v.Values[0].(int64))

	// valid value, contains multiple elements
	value = "{1 2 3}"
	v, err = parseOrList(value)
	require.Nil(t, err)
	require.Len(t, v.Values, 3)
	assert.Equal(t, int64(1), v.Values[0].(int64))
	assert.Equal(t, int64(2), v.Values[1].(int64))
	assert.Equal(t, int64(3), v.Values[2].(int64))
}

func TestParseAndList(t *testing.T) {
	// invalid
	value := ")("
	v, err := parseAndList(value)
	require.NotNil(t, err)

	// valid value, contains no element
	value = "()"
	v, err = parseAndList(value)
	require.Nil(t, err)
	assert.Len(t, v.Values, 0)

	// valid value, contains only one element
	value = "(1)"
	v, err = parseAndList(value)
	require.Nil(t, err)
	require.Len(t, v.Values, 1)
	assert.Equal(t, int64(1), v.Values[0].(int64))

	// valid value, contains multiple elements
	value = "(1 2 3)"
	v, err = parseAndList(value)
	require.Nil(t, err)
	require.Len(t, v.Values, 3)
	assert.Equal(t, int64(1), v.Values[0].(int64))
	assert.Equal(t, int64(2), v.Values[1].(int64))
	assert.Equal(t, int64(3), v.Values[2].(int64))
}

func TestParseValue(t *testing.T) {
	// time
	value := "2020-03-04T17:08:23"
	v := parseValue(value)
	_, ok := v.(time.Time)
	require.True(t, ok)

	// integer
	value = "1"
	v = parseValue(value)
	i, ok := v.(int64)
	require.True(t, ok)
	assert.Equal(t, int64(1), i)

	// empty string
	value = ""
	v = parseValue(value)
	str, ok := v.(string)
	require.True(t, ok)
	assert.Equal(t, "", str)

	// not empty string
	value = "abc"
	v = parseValue(value)
	str, ok = v.(string)
	require.True(t, ok)
	assert.Equal(t, "abc", str)

	// not empty string
	value = `"abc"`
	v = parseValue(value)
	str, ok = v.(string)
	require.True(t, ok)
	assert.Equal(t, "abc", str)
}

func TestEscapeValue(t *testing.T) {
	// empty string
	value := ""
	v := escapeValue(value)
	assert.Equal(t, "", v)

	// string contains no special character
	value = "abc"
	v = escapeValue(value)
	assert.Equal(t, "abc", v)

	// string starts with special character
	value = `\~abc`
	v = escapeValue(value)
	assert.Equal(t, "~abc", v)
}

func TestParsePattern(t *testing.T) {
	// empty string
	value := ""
	v, err := parsePattern(value)
	require.Nil(t, err)
	_, ok := v.(string)
	assert.True(t, ok)

	// fuzzy match
	value = "~a"
	v, err = parsePattern(value)
	require.Nil(t, err)
	_, ok = v.(*FuzzyMatchValue)
	assert.True(t, ok)

	// range
	value = "[1~3]"
	v, err = parsePattern(value)
	require.Nil(t, err)
	_, ok = v.(*Range)
	assert.True(t, ok)

	// or list
	value = "{1 2}"
	v, err = parsePattern(value)
	require.Nil(t, err)
	_, ok = v.(*OrList)
	assert.True(t, ok)

	// and list
	value = "(1 3)"
	v, err = parsePattern(value)
	require.Nil(t, err)
	_, ok = v.(*AndList)
	assert.True(t, ok)

	// exact match
	value = "a"
	v, err = parsePattern(value)
	require.Nil(t, err)
	_, ok = v.(string)
	assert.True(t, ok)
}

func TestBuild(t *testing.T) {
	// empty string
	q := ``
	query, err := Build(q, 1, 10)
	require.Nil(t, err)
	require.NotNil(t, query)
	assert.Equal(t, int64(1), query.PageNumber)
	assert.Equal(t, int64(10), query.PageSize)

	// contains only ","
	q = `,`
	query, err = Build(q, 1, 10)
	require.NotNil(t, err)

	// valid query string
	q = `k=v`
	query, err = Build(q, 1, 10)
	require.Nil(t, err)
	assert.Equal(t, int64(1), query.PageNumber)
	assert.Equal(t, int64(10), query.PageSize)
	assert.Equal(t, "v", query.Keywords["k"].(string))

	// contains escaped characters
	q = `k%3Dv`
	query, err = Build(q, 1, 10)
	require.Nil(t, err)
	assert.Equal(t, int64(1), query.PageNumber)
	assert.Equal(t, int64(10), query.PageSize)
	assert.Equal(t, "v", query.Keywords["k"].(string))
}
