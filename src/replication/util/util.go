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

package util

import (
	"net/http"
	"path/filepath"

	"github.com/goharbor/harbor/src/common/utils/registry"
)

// Match returns whether the str matches the pattern
func Match(pattern, str string) (bool, error) {
	return filepath.Match(pattern, str)
}

// GetHTTPTransport can be used to share the common HTTP transport
func GetHTTPTransport(insecure bool) *http.Transport {
	return registry.GetHTTPTransport(insecure)
}
