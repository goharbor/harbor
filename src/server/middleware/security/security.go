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
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/internal"
)

var (
	generators = []generator{
		&secret{},
		&oidcCli{},
		&idToken{},
		&authProxy{},
		&robot{},
		&basicAuth{},
		&session{},
		&unauthorized{},
	}
)

// security context generator
type generator interface {
	Generate(req *http.Request) security.Context
}

// Middleware returns a security context middleware that populates the security context into the request context
func Middleware() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := log.G(r.Context())
			mode, err := config.AuthMode()
			if err == nil {
				r = r.WithContext(internal.WithAuthMode(r.Context(), mode))
			} else {
				log.Warningf("failed to get auth mode: %v", err)
			}
			for _, generator := range generators {
				if ctx := generator.Generate(r); ctx != nil {
					r = r.WithContext(security.NewContext(r.Context(), ctx))
					break
				}
			}
			handler.ServeHTTP(w, r)
		})
	}
}
