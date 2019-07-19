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

package registryproxy

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type proxyHandler struct {
	handler http.Handler
}

// New ...
func New(urls ...string) http.Handler {
	var registryURL string
	var err error
	if len(urls) > 1 {
		log.Errorf("the parm, urls should have only 0 or 1 elements")
		return nil
	}
	if len(urls) == 0 {
		registryURL, err = config.RegistryURL()
		if err != nil {
			log.Error(err)
			return nil
		}
	} else {
		registryURL = urls[0]
	}
	targetURL, err := url.Parse(registryURL)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &proxyHandler{
		handler: httputil.NewSingleHostReverseProxy(targetURL),
	}

}

// ServeHTTP ...
func (ph proxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ph.handler.ServeHTTP(rw, req)
}
