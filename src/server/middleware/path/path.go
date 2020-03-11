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

package path

import (
	"net/http"
	"net/url"
	"regexp"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	defaultRegexps = []*regexp.Regexp{
		regexp.MustCompile(`^/api/` + api.APIVersion + `/projects/.*/repositories/(.*)/artifacts/?$`),
		regexp.MustCompile(`^/api/` + api.APIVersion + `/projects/.*/repositories/(.*)/artifacts/.*$`),
		regexp.MustCompile(`^/api/` + api.APIVersion + `/projects/.*/repositories/(.*)/?$`),
	}
)

// EscapeMiddleware middleware which escape path parameters for swagger APIs
func EscapeMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		for _, re := range defaultRegexps {
			if re.MatchString(r.URL.Path) {
				r.URL.Path = escape(re, r.URL.Path)
				break
			}
		}

		next.ServeHTTP(w, r)
	})
}

func escape(re *regexp.Regexp, path string) string {
	return replaceAllSubmatchFunc(re, path, func(groups []string) []string {
		var results []string
		for _, group := range groups {
			results = append(results, url.PathEscape(group))
		}
		return results
	}, -1)
}

func replaceAllSubmatchFunc(re *regexp.Regexp, src string, repl func([]string) []string, n int) string {
	var result string

	last := 0
	for _, match := range re.FindAllSubmatchIndex([]byte(src), n) {
		// Append string between our last match and this one (i.e. non-matched string).
		matchStart := match[0]
		matchEnd := match[1]
		result = result + src[last:matchStart]
		last = matchEnd

		// Determine the groups / submatch string and indices.
		groups := []string{}
		indices := [][2]int{}
		for i := 2; i < len(match); i += 2 {
			start := match[i]
			end := match[i+1]
			groups = append(groups, src[start:end])
			indices = append(indices, [2]int{start, end})
		}

		// Replace the groups
		groups = repl(groups)

		// Append match data.
		lastGroup := matchStart
		for i, newValue := range groups {
			// Append string between our last group match and this one (i.e. non-group-matched string)
			groupStart := indices[i][0]
			groupEnd := indices[i][1]
			result = result + src[lastGroup:groupStart]
			lastGroup = groupEnd

			// Append the new group value.
			result = result + newValue
		}
		result = result + src[lastGroup:matchEnd] // remaining
	}

	result = result + src[last:] // remaining

	return result
}
