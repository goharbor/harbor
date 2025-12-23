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

package url

import (
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/server/middleware"
)

// Middleware validates the URL query string, rejecting requests with:
//   - Invalid semicolon separators
//   - Null bytes (that can cause DB errors)
//   - Invalid UTF-8 sequences
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if r.URL != nil && r.URL.RawQuery != "" {
			if err := validateQueryString(r.URL.RawQuery); err != nil {
				lib_http.SendError(w, err)
				return
			}
		}
		next.ServeHTTP(w, r)
	}, skippers...)
}

// containsInvalidChars checks for null bytes or invalid UTF-8
func containsInvalidChars(s string) bool {
	return strings.Contains(s, "\x00") || !utf8.ValidString(s)
}

func validateQueryString(rawQuery string) error {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return errors.New(err).WithCode(errors.BadRequestCode)
	}

	for key, vals := range values {
		if containsInvalidChars(key) {
			return errors.BadRequestError(nil).
				WithMessage("query parameter key contains invalid characters (null bytes or invalid UTF-8)")
		}
		for _, v := range vals {
			if containsInvalidChars(v) {
				return errors.BadRequestError(nil).
					WithMessagef("query parameter %q contains invalid characters (null bytes or invalid UTF-8)", key)
			}
		}
	}
	return nil
}
