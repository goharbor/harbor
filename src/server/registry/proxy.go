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
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/config"
)

var proxy = newProxy()

func newProxy() http.Handler {
	regURL, _ := config.RegistryURL()
	url, err := url.Parse(regURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse the URL of registry: %v", err))
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	// The reverse proxy forwards all /v2/ traffic to a single host (the registry),
	// so it deserves a dedicated transport with a large per-host idle connection
	// pool. Relying on http.DefaultTransport (MaxIdleConnsPerHost=2) closes almost
	// every connection after use under high concurrency, piling up TIME_WAIT
	// sockets and eventually exhausting ephemeral ports (EADDRNOTAVAIL).
	opts := []func(*http.Transport){
		func(tr *http.Transport) {
			tr.MaxIdleConns = 1024
			tr.MaxIdleConnsPerHost = 1024
		},
	}
	if commonhttp.InternalTLSEnabled() {
		opts = append(opts, commonhttp.WithInternalTLSConfig())
	}
	proxy.Transport = commonhttp.NewTransport(opts...)

	proxy.Director = basicAuthDirector(proxy.Director)
	return proxy
}

func basicAuthDirector(d func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		d(r)
		if r != nil {
			u, p := config.RegistryCredential()
			r.SetBasicAuth(u, p)
		}
	}
}
