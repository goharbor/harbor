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

package middleware

import (
	"net/http"
	"regexp"
)

// Skipper defines a function to skip middleware.
// Returning true skips processing the middleware.
type Skipper func(*http.Request) bool

// MethodAndPathSkipper returns skipper which
// will skip the middleware when r.Method equals the method and r.URL.Path matches the re
// when method is "*" it equals all http method
func MethodAndPathSkipper(method string, re *regexp.Regexp) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		if (method == "*" || r.Method == method) && re.MatchString(r.URL.Path) {
			return true
		}

		return false
	}
}
