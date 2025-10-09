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

package session

import (
	"net/http"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
)

// Middleware returns a session middleware that populates the information indicates whether
// the request carries session or not
func Middleware() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// We can check the cookie directly b/c the filter and controllerRegistry is executed after middleware, so no session
			// cookie is added by beego.
			_, err := r.Cookie(config.SessionCookieName)
			if err == nil {
				r = r.WithContext(lib.WithCarrySession(r.Context(), true))
			}
			handler.ServeHTTP(w, r)
		})
	}
}
