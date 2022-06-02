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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

// TrimsLineBreaks trims line breaks in string.
func TrimLineBreaks(s string) string {
	escaped := strings.ReplaceAll(s, "\n", "")
	escaped = strings.ReplaceAll(escaped, "\r", "")
	return escaped
}

// Title uppercase the first character, and lower case the rest, for example covert MANUAL to Manual
func Title(s string) string {
	title := cases.Title(language.Und)
	return title.String(strings.ToLower(s))
}
