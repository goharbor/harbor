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

package notary

import (
	"net/http"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// Transport holds information about base transport and modifiers
type Transport struct {
	transport http.RoundTripper
	modifiers []modifier.Modifier
}

// NewTransport ...
func NewTransport(transport http.RoundTripper, modifiers ...modifier.Modifier) *Transport {
	return &Transport{
		transport: transport,
		modifiers: modifiers,
	}
}

// RoundTrip ...
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, modifier := range t.modifiers {
		if err := modifier.Modify(req); err != nil {
			return nil, err
		}
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	log.Debugf("%d | %s %s", resp.StatusCode, req.Method, req.URL.String())

	return resp, err
}
