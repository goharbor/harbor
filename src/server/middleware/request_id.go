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

	"github.com/google/uuid"
)

// HeaderXRequestID X-Request-ID header
const HeaderXRequestID = "X-Request-ID"

// RequestID middleware which add X-Request-ID header in the http request when not exist
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get(HeaderXRequestID)
			if rid == "" {
				rid = uuid.New().String()
				r.Header.Set(HeaderXRequestID, rid)
			}

			w.Header().Set(HeaderXRequestID, rid)
			next.ServeHTTP(w, r)
		})
	}
}
