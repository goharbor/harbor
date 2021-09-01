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

package security

import (
	"net/http"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	generators = []generator{
		&secret{},
		&oidcCli{},
		&v2Token{},
		&idToken{},
		&authProxy{},
		&robot{},
		&basicAuth{},
		&session{},
		&proxyCacheSecret{},
	}
)

// security context generator
type generator interface {
	Generate(req *http.Request) security.Context
}

// Middleware returns a security context middleware that populates the security context into the request context
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		log := log.G(r.Context())
		mode, err := config.AuthMode(r.Context())
		if err == nil {
			r = r.WithContext(lib.WithAuthMode(r.Context(), mode))
		} else {
			log.Warningf("failed to get auth mode: %v", err)
		}
		for _, generator := range generators {
			if ctx := generator.Generate(r); ctx != nil {
				r = r.WithContext(security.NewContext(r.Context(), ctx))
				break
			}
		}
		next.ServeHTTP(w, r)
	}, skippers...)
}

// UnauthorizedMiddleware returns a security context middleware
// that populates the unauthorized security context when not security context found in the request context
func UnauthorizedMiddleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if _, ok := security.FromContext(r.Context()); !ok {
			u := &unauthorized{}
			r = r.WithContext(security.NewContext(r.Context(), u.Generate(r)))
		}

		next.ServeHTTP(w, r)
	}, skippers...)
}
