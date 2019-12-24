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

package manifest

import (
	"net/http"
	"net/http/httputil"
)

// NewHandler returns the handler to handler manifest requests
func NewHandler(proxy *httputil.ReverseProxy) http.Handler {
	return &handler{
		proxy: proxy,
	}
}

type handler struct {
	proxy *httputil.ReverseProxy
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodHead:
		h.head(w, req)
	case http.MethodGet:
		h.get(w, req)
	case http.MethodDelete:
		h.delete(w, req)
	case http.MethodPut:
		h.put(w, req)
	}
}

// make sure the artifact exist before proxying the request to the backend registry
func (h *handler) head(w http.ResponseWriter, req *http.Request) {
	// TODO check the existence
	h.proxy.ServeHTTP(w, req)
}

// make sure the artifact exist before proxying the request to the backend registry
func (h *handler) get(w http.ResponseWriter, req *http.Request) {
	// TODO check the existence
	h.proxy.ServeHTTP(w, req)
}

func (h *handler) delete(w http.ResponseWriter, req *http.Request) {
	// TODO implement, just delete from database
}

func (h *handler) put(w http.ResponseWriter, req *http.Request) {
	h.proxy.ServeHTTP(w, req)
}
