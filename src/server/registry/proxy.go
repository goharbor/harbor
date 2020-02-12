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
	"github.com/goharbor/harbor/src/core/config"
)

var proxy = newProxy()

func newProxy() http.Handler {
	regURL, _ := config.RegistryURL()
	url, err := url.Parse(regURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse the URL of registry: %v", err))
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	if commonhttp.InternalTLSEnabled() {
		proxy.Transport = commonhttp.GetHTTPTransport(commonhttp.InternalTransport)
	}

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
