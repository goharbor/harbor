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

package internal

import (
	"crypto/tls"
	"net/http"
)

var (
	secureHTTPTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	insecureHTTPTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
)

// GetHTTPTransport returns the HTTP transport based on insecure configuration
func GetHTTPTransport(insecure ...bool) *http.Transport {
	if len(insecure) > 0 && insecure[0] {
		return insecureHTTPTransport
	}
	return secureHTTPTransport
}
