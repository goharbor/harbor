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

package apiversion

import (
	"github.com/goharbor/harbor/src/internal"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

// Middleware returns a middleware that set the API version into the context
func Middleware(version string) middleware.Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := internal.SetAPIVersion(req.Context(), version)
			handler.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}
