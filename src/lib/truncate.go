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

package lib

import (
	"strings"
	"unicode/utf8"
)

// Truncate tries to append the "suffix" to the "str". If the length of the appended string exceeds "n",
// the function truncates the "str" to make sure the "suffix" is appended
func Truncate(str, suffix string, n int) string {
	s := str + suffix
	if len(s) <= n {
		return s
	}
	return s[:len(str)-(len(s)-n)] + suffix
}

// TruncateUTF8 truncates s to at most maxLen bytes, ensuring that the truncation boundary
// does not break a multi-byte UTF-8 character, and appends "..." if truncated.
// If len(s) <= maxLen, s is returned unmodified.
func TruncateUTF8(s string, maxLen int) string {
	s = strings.ToValidUTF8(s, "")
	if len(s) <= maxLen {
		return s
	}
	const suffix = "..."
	if maxLen <= 0 {
		return ""
	}
	if maxLen <= len(suffix) {
		return suffix[:maxLen]
	}
	target := maxLen - len(suffix)
	for target > 0 && !utf8.RuneStart(s[target]) {
		target--
	}
	return s[:target] + suffix
}
