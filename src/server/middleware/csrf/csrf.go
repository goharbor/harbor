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

package csrf

import (
	"net/http"
	"strings"
	"sync"

	csrf "filippo.io/csrf/gorilla"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	once    sync.Once
	protect func(handler http.Handler) http.Handler
)

func handleError(w http.ResponseWriter, r *http.Request) {
	lib_http.SendError(w, errors.New(csrf.FailureReason(r)).WithCode(errors.ForbiddenCode))
}

// Middleware initialize the middleware to apply csrf selectively
func Middleware() func(handler http.Handler) http.Handler {
	once.Do(func() {
		protect = csrf.Protect(nil,
			csrf.ErrorHandler(http.HandlerFunc(handleError)),
		)
	})
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		protect(next).ServeHTTP(rw, req)
	}, csrfSkipper)
}

// csrfSkipper makes sure only some of the uris accessed by non-UI client can skip the csrf check
func csrfSkipper(req *http.Request) bool {
	path := req.URL.Path
	if (strings.HasPrefix(path, "/v2/") ||
		strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/service/")) && !lib.GetCarrySession(req.Context()) {
		return true
	}
	return false
}
