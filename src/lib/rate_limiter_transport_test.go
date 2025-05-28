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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_newRateLimitedTransport(t *testing.T) {
	tests := []struct {
		name      string
		rate      int
		transport http.RoundTripper
	}{
		{"1qps", 1, http.DefaultTransport},
	}
	var req = &http.Request{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRateLimitedTransport(tt.rate, tt.transport)
			start := time.Now()
			for i := 0; i <= tt.rate; i++ {
				got.RoundTrip(req)
			}
			used := int64(time.Since(start).Milliseconds()) / int64(tt.rate)
			assert.GreaterOrEqualf(t, used/int64(tt.rate), int64(1e3/tt.rate), "used %d ms per req", used/int64(tt.rate))
		})
	}
}
