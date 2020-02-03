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
	"github.com/goharbor/harbor/src/server/middleware/contenttrust"
	"github.com/goharbor/harbor/src/server/middleware/vulnerable"

	"github.com/goharbor/harbor/src/core/config"
	pkg_repo "github.com/goharbor/harbor/src/pkg/repository"
	pkg_tag "github.com/goharbor/harbor/src/pkg/tag"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/immutable"
	"github.com/goharbor/harbor/src/server/middleware/manifestinfo"
	"github.com/goharbor/harbor/src/server/middleware/readonly"
	"github.com/goharbor/harbor/src/server/middleware/regtoken"
	"github.com/goharbor/harbor/src/server/registry/blob"
	"github.com/goharbor/harbor/src/server/registry/catalog"
	"github.com/goharbor/harbor/src/server/registry/manifest"
	"github.com/goharbor/harbor/src/server/registry/tag"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// New return the registry instance to handle the registry APIs
func New(url *url.URL) http.Handler {
	// TODO customize the reverse proxy to improve the performance?
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Director = basicAuthDirector(proxy.Director)

	// create the root rooter
	rootRouter := mux.NewRouter()
	rootRouter.StrictSlash(true)

	// handle catalog
	rootRouter.Path("/v2/_catalog").Methods(http.MethodGet).Handler(catalog.NewHandler(pkg_repo.Mgr))

	// handle list tag
	rootRouter.Path("/v2/{name:.*}/tags/list").Methods(http.MethodGet).Handler(tag.NewHandler(pkg_repo.Mgr, pkg_tag.Mgr))

	// handle manifest
	// TODO maybe we should split it into several sub routers based on the method
	manifestRouter := rootRouter.Path("/v2/{name:.*}/manifests/{reference}").Subrouter()
	manifestRouter.NewRoute().Methods(http.MethodGet).Handler(middleware.WithMiddlewares(manifest.NewHandler(proxy), manifestinfo.Middleware(), regtoken.Middleware(), contenttrust.Middleware(), vulnerable.Middleware()))
	manifestRouter.NewRoute().Methods(http.MethodHead).Handler(manifest.NewHandler(proxy))
	manifestRouter.NewRoute().Methods(http.MethodDelete).Handler(middleware.WithMiddlewares(manifest.NewHandler(proxy), readonly.Middleware(), manifestinfo.Middleware(), immutable.MiddlewareDelete()))

	// handle blob
	// as we need to apply middleware to the blob requests, so create a sub router to handle the blob APIs
	blobRouter := rootRouter.PathPrefix("/v2/{name:.*}/blobs/").Subrouter()
	blobRouter.NewRoute().Methods(http.MethodGet).Handler(blob.NewHandler(proxy))
	blobRouter.NewRoute().Methods(http.MethodHead).Handler(blob.NewHandler(proxy))
	blobRouter.NewRoute().Methods(http.MethodPost).Handler(middleware.WithMiddlewares(blob.NewHandler(proxy), readonly.Middleware()))
	blobRouter.NewRoute().Methods(http.MethodPut).Handler(middleware.WithMiddlewares(blob.NewHandler(proxy), readonly.Middleware()))
	blobRouter.NewRoute().Methods(http.MethodPatch).Handler(middleware.WithMiddlewares(blob.NewHandler(proxy), readonly.Middleware()))
	blobRouter.NewRoute().Methods(http.MethodDelete).Handler(middleware.WithMiddlewares(blob.NewHandler(proxy), readonly.Middleware()))

	// all other APIs are proxy to the backend docker registry
	rootRouter.PathPrefix("/").Handler(proxy)

	// register middlewares
	// TODO add auth middleware
	// TODO apply the existing middlewares
	// rootRouter.Use(mux.MiddlewareFunc(middleware))

	return rootRouter
}

func basicAuthDirector(d func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		d(r)
		if r != nil && !middleware.SkipInjectRegistryCred(r.Context()) {
			u, p := config.RegistryCredential()
			r.SetBasicAuth(u, p)
		}
	}
}
