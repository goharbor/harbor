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

import "net/http"

// Middleware receives a handler and returns another handler.
// The returned handler can do some customized task according to
// the requirement
type Middleware func(http.Handler) http.Handler

// WithMiddlewares apply the middlewares to the handler.
// The middlewares are executed in the order that they are applied
func WithMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// Skipper defines a function to skip middleware.
// Returning true skips processing the middleware.
type Skipper func(*http.Request) bool

// New make a middleware from fn which type is func(w http.ResponseWriter, r *http.Request, next http.Handler)
func New(fn func(http.ResponseWriter, *http.Request, http.Handler), skippers ...Skipper) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, skipper := range skippers {
				if skipper(r) {
					next.ServeHTTP(w, r)
					return
				}
			}

			fn(w, r, next)
		})
	}
}
