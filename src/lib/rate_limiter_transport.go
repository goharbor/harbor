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

package lib

import (
	"net/http"

	"go.uber.org/ratelimit"
)

type limitTransport struct {
	http.RoundTripper
	limiter ratelimit.Limiter
}

var _ http.RoundTripper = limitTransport{}

func (t limitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.limiter.Take()
	return t.RoundTripper.RoundTrip(req)
}

func NewRateLimitedTransport(rate int, transport http.RoundTripper) http.RoundTripper {
	return &limitTransport{
		RoundTripper: transport,
		limiter:      ratelimit.New(rate),
	}
}
