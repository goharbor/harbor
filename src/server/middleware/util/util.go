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
	"github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"net/http"
	"path"
	"strings"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/pkg/distribution"
)

// ParseProjectName parse project name from v2 and v2.0 API URL path
func ParseProjectName(r *http.Request) string {
	path := path.Clean(r.URL.EscapedPath())

	var projectName string

	prefixes := []string{
		fmt.Sprintf("/api/%s/projects/", api.APIVersion), // v2.0 management APIs
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

// SkipPolicyChecking ...
func SkipPolicyChecking(r *http.Request, projectID, artID int64) (bool, error) {
	secCtx, ok := security.FromContext(r.Context())

	// 1, scanner pull access can bypass.
	// 2, cosign pull can bypass, it needs to pull the manifest before pushing the signature.
	// 3, pull cosign signature can bypass.
	if ok && secCtx.Name() == "v2token" {
		if secCtx.Can(r.Context(), rbac.ActionScannerPull, project.NewNamespace(projectID).Resource(rbac.ResourceRepository)) ||
			(secCtx.Can(r.Context(), rbac.ActionPush, project.NewNamespace(projectID).Resource(rbac.ResourceRepository)) &&
				strings.Contains(r.UserAgent(), "cosign")) {
			return true, nil
		}
	}

	accs, err := accessory.Mgr.List(r.Context(), q.New(q.KeyWords{"ArtifactID": artID}))
	if err != nil {
		return false, err
	}
	if len(accs) > 0 && accs[0].GetData().Type == model.TypeCosignSignature {
		return true, nil
	}

	return false, nil
}
