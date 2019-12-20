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
	"github.com/goharbor/harbor/src/pkg/project"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/goharbor/harbor/src/server/registry/catalog"
	"github.com/goharbor/harbor/src/server/registry/manifest"
	"github.com/goharbor/harbor/src/server/registry/tag"
	"github.com/gorilla/mux"
)

// New return the registry instance to handle the registry APIs
func New(url *url.URL) http.Handler {
	// TODO add a director to add the basic auth for docker registry
	// TODO customize the reverse proxy to improve the performance?
	proxy := httputil.NewSingleHostReverseProxy(url)

	// create the root rooter
	rootRouter := mux.NewRouter()
	rootRouter.StrictSlash(true)

	// handle catalog
	rootRouter.Path("/v2/_catalog").Methods(http.MethodGet).Handler(catalog.NewHandler())

	// handle list tag
	rootRouter.Path("/v2/{name:.*}/tags/list").Methods(http.MethodGet).Handler(tag.NewHandler())

	// handle manifest
	// TODO maybe we should split it into several sub routers based on the method
	manifestRouter := rootRouter.Path("/v2/{name:.*}/manifests/{reference}").Subrouter()
	manifestRouter.NewRoute().Methods(http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete).
		Handler(manifest.NewHandler(project.Mgr, proxy))

	// handle blob
	// as we need to apply middleware to the blob requests, so create a sub router to handle the blob APIs
	blobRouter := rootRouter.PathPrefix("/v2/{name:.*}/blobs/").Subrouter()
	blobRouter.NewRoute().Methods(http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete).
		Handler(proxy)

	// all other APIs are proxy to the backend docker registry
	rootRouter.PathPrefix("/").Handler(proxy)

	// register middlewares
	// TODO add auth middleware
	// TODO apply the existing middlewares
	// rootRouter.Use(mux.MiddlewareFunc(middleware))

	return rootRouter
}
