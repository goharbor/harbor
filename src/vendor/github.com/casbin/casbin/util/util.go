// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"regexp"
	"sort"
	"strings"
)

// EscapeAssertion escapes the dots in the assertion, because the expression evaluation doesn't support such variable names.
func EscapeAssertion(s string) string {
	//Replace the first dot, because it can't be recognized by the regexp.
	if (strings.HasPrefix(s, "r") || strings.HasPrefix(s, "p")) {
		s = strings.Replace(s, ".", "_",1)
	}
	var regex = regexp.MustCompile(`(\|| |=|\)|\(|&|<|>|,|\+|-|!|\*|\/)(r|p)\.`)
	s = regex.ReplaceAllStringFunc(s, func(m string) string {
		return strings.Replace(m, ".", "_", 1)
	})
	return s
}

// RemoveComments removes the comments starting with # in the text.
func RemoveComments(s string) string {
	pos := strings.Index(s, "#")
	if pos == -1 {
		return s
	}
	return strings.TrimSpace(s[0:pos])
}

// ArrayEquals determines whether two string arrays are identical.
func ArrayEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Array2DEquals determines whether two 2-dimensional string arrays are identical.
func Array2DEquals(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if !ArrayEquals(v, b[i]) {
			return false
		}
	}
	return true
}

// ArrayRemoveDuplicates removes any duplicated elements in a string array.
func ArrayRemoveDuplicates(s *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *s {
		if !found[x] {
			found[x] = true
			(*s)[j] = (*s)[i]
			j++
		}
	}
	*s = (*s)[:j]
}

// ArrayToString gets a printable string for a string array.
func ArrayToString(s []string) string {
	return strings.Join(s, ", ")
}

// ParamsToString gets a printable string for variable number of parameters.
func ParamsToString(s ...string) string {
	return strings.Join(s, ", ")
}

// SetEquals determines whether two string sets are identical.
func SetEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// JoinSlice joins a string and a slice into a new slice.
func JoinSlice(a string, b ...string) []string {
	res := make([]string, 0, len(b)+1)

	res = append(res, a)
	for _, s := range b {
		res = append(res, s)
	}

	return res
}

// JoinSliceAny joins a string and a slice into a new interface{} slice.
func JoinSliceAny(a string, b ...string) []interface{} {
	res := make([]interface{}, 0, len(b)+1)

	res = append(res, a)
	for _, s := range b {
		res = append(res, s)
	}

	return res
}

// SetSubtract returns the elements in `a` that aren't in `b`.
func SetSubtract(a []string, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
