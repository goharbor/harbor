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
	"os"
	"strings"
	"sync"

	"github.com/gorilla/csrf"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/middleware"
)

const (
	csrfKeyEnv  = "CSRF_KEY"
	tokenHeader = "X-Harbor-CSRF-Token"
)

var (
	once       sync.Once
	secureFlag = true
	protect    func(handler http.Handler) http.Handler
)

// attachToken makes sure if csrf generate a new token it will be included in the response header
func attachToken(w http.ResponseWriter, r *http.Request) {
	if t := csrf.Token(r); len(t) > 0 {
		w.Header().Set(tokenHeader, t)
	} else {
		log.Warningf("token not found in context, skip attaching")
	}
}

func handleError(w http.ResponseWriter, r *http.Request) {
	attachToken(w, r)
	lib_http.SendError(w, errors.New(csrf.FailureReason(r)).WithCode(errors.ForbiddenCode))
}

func attach(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		attachToken(rw, req)
		handler.ServeHTTP(rw, req)
	})
}

// Middleware initialize the middleware to apply csrf selectively
func Middleware() func(handler http.Handler) http.Handler {
	once.Do(func() {
		key := os.Getenv(csrfKeyEnv)
		if len(key) == 0 {
			key = utils.GenerateRandomString()
		} else if len(key) != 32 {
			log.Errorf("Invalid CSRF key length from the environment: %s. Please ensure the key length is 32 characters.", key)
			protect = func(_ http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					lib_http.SendError(w, errors.New("invalid CSRF key length from the environment. Please ensure the key length is 32 characters"))
				})
			}
			return
		}
		secureFlag = secureCookie()
		protect = csrf.Protect([]byte(key), csrf.RequestHeader(tokenHeader),
			csrf.Secure(secureFlag),
			csrf.ErrorHandler(http.HandlerFunc(handleError)),
			csrf.SameSite(csrf.SameSiteStrictMode),
			csrf.Path("/"))
	})
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		protect(attach(next)).ServeHTTP(rw, req)
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

func secureCookie() bool {
	ep, err := config.ExtEndpoint()
	if err != nil {
		log.Warningf("Failed to get external endpoint: %v, set cookie secure flag to true", err)
		return true
	}
	return !strings.HasPrefix(strings.ToLower(ep), "http://")
}
