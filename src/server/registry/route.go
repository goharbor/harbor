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

package registry

import (
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/server/middleware/immutable"
	"github.com/goharbor/harbor/src/server/middleware/manifestinfo"
	"github.com/goharbor/harbor/src/server/middleware/readonly"
	"github.com/goharbor/harbor/src/server/registry/manifest"
	"github.com/goharbor/harbor/src/server/router"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// RegisterRoutes for OCI registry APIs
func RegisterRoutes() {
	// TODO remove
	regURL, _ := config.RegistryURL()
	url, _ := url.Parse(regURL)
	proxy := httputil.NewSingleHostReverseProxy(url)

	router.NewRoute().Path("/v2/*").Handler(New(url))
	router.NewRoute().
		Method(http.MethodPut).
		Path("/v2/*/manifests/:reference").
		Middleware(readonly.Middleware()).
		Middleware(manifestinfo.Middleware()).
		Middleware(immutable.MiddlewarePush()).
		Handler(manifest.NewHandler(proxy))
}
