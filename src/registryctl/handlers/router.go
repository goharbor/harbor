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

package handlers

import (
	"net/http"

	"github.com/goharbor/harbor/src/registryctl/api"
	"github.com/goharbor/harbor/src/registryctl/api/registry/blob"
	"github.com/goharbor/harbor/src/registryctl/api/registry/manifest"
	"github.com/goharbor/harbor/src/registryctl/config"
)

func newRouter(conf config.Configuration) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		api.Health(rw, req)
	})

	mux.HandleFunc("/api/registry/blob/{reference}", func(rw http.ResponseWriter, req *http.Request) {
		blob.NewHandler(conf.StorageDriver).ServeHTTP(rw, req)
	})

	mux.HandleFunc("/api/registry/{name}/manifests/{reference}", func(rw http.ResponseWriter, req *http.Request) {
		manifest.NewHandler(conf.StorageDriver).ServeHTTP(rw, req)
	})

	return mux
}
