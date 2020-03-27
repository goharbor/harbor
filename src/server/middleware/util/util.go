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
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/pkg/distribution"
)

// ParseProjectName parse project name from v2 and v2.0 API URL path
func ParseProjectName(r *http.Request) string {
	path := path.Clean(r.URL.EscapedPath())

	var projectName string

	prefixes := []string{
		fmt.Sprintf("/api/%s/projects/", api.APIVersion), // v2.0 management APIs
		"/api/chartrepo/", // chartmuseum APIs
		fmt.Sprintf("/api/%s/chartrepo/", api.APIVersion), // chartmuseum Label APIs
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
			if len(parts) > 0 {
				projectName = parts[0]
				break
			}
		}
	}

	if projectName == "" && strings.HasPrefix(path, "/v2/") {
		// v2 APIs
		projectName = distribution.ParseProjectName(path)
	}

	return projectName
}
