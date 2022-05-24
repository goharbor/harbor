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

package auth

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/registry/auth"
)

// authorizer is a customize authorizer for quay adapter which
// inherits lib authorizer.
type authorizer struct {
	innerAuthorizer lib.Authorizer
}

// NewAuthorizer creates an authorizer instance.
func NewAuthorizer(username, password string, insecure bool) lib.Authorizer {
	return &authorizer{innerAuthorizer: auth.NewAuthorizer(username, password, insecure)}
}

// Modify implements the lib.Authorizer.
func (a *authorizer) Modify(req *http.Request) error {
	// if request api is catalog, remove the suffix _catalog
	// to avoid lib authorizer parse scope and adds scope when
	// request token.
	// cause: https://github.com/goharbor/harbor/issues/13200
	if isCatalog(req) {
		// rewrite path
		oldPath := req.URL.Path
		defer func() {
			// resume path
			req.URL.Path = oldPath
		}()
		req.URL.Path = strings.TrimSuffix(req.URL.Path, "_catalog")
	}

	return a.innerAuthorizer.Modify(req)
}

var catalog = regexp.MustCompile("/v2/_catalog$")

// isCatalog detects if the api is /v2/_catalog.
func isCatalog(req *http.Request) bool {
	path := strings.TrimRight(req.URL.Path, "/")
	return catalog.MatchString(path)
}
