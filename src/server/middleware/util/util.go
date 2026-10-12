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
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
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
		if after, ok := strings.CutPrefix(path, prefix); ok {
			parts := strings.Split(after, "/")
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

	// Scanner pull is authorised by the scanner-pull RBAC action, which is carried in the
	// signed bearer token's scope and cannot be forged, so it may bypass the policy.
	//
	// The former cosign/notation exemptions here were gated on the request User-Agent, a
	// fully client-controlled string (CWE-807): any push-capable
	// principal could skip both the content-trust and the vulnerability-prevention policy by
	// sending "User-Agent: cosign". They have been removed. A genuinely signed artifact is
	// still recognised below by its actual signature accessory, not by any client header.
	// The pull-before-first-signature bootstrap for content-trust is preserved, opt-in only,
	// by the content-trust middleware via LegacySignerBootstrapPull.
	if ok && secCtx.Name() == "v2token" {
		if secCtx.Can(r.Context(), rbac.ActionScannerPull, project.NewNamespace(projectID).Resource(rbac.ResourceRepository)) {
			return true, nil
		}
	}

	accs, err := accessory.Mgr.List(r.Context(), q.New(q.KeyWords{"ArtifactID": artID}))
	if err != nil {
		return false, err
	}
	if len(accs) > 0 && (accs[0].GetData().Type == model.TypeCosignSignature || accs[0].GetData().Type == model.TypeNotationSignature) {
		return true, nil
	}

	return false, nil
}

// LegacySignerPullEnabled reports whether the legacy User-Agent-based signer-pull
// exemption for the content-trust policy is enabled.
//
// Unset or empty defaults to true to avoid breaking existing cosign and notation
// signing workflows on upgrade (where signing clients must pull the unsigned manifest
// to compute the first signature). Operators can set CONTENT_TRUST_LEGACY_SIGNER_PULL_ENABLED=false
// to strictly enforce content-trust on all manifest pulls; a malformed value fails closed.
// It never affects the vulnerability-prevention policy.
func LegacySignerPullEnabled() bool {
	val := strings.TrimSpace(os.Getenv("CONTENT_TRUST_LEGACY_SIGNER_PULL_ENABLED"))
	if val == "" {
		return true
	}
	enabled, err := strconv.ParseBool(val)
	return err == nil && enabled
}

// LegacySignerBootstrapPull reports whether the request looks like a push-capable signing
// client fetching a subject manifest before pushing its first signature. The decision
// includes the client-supplied User-Agent and is therefore spoofable, so callers must gate
// it behind LegacySignerPullEnabled and must never use it for the vulnerability-prevention
// policy.
func LegacySignerBootstrapPull(r *http.Request, projectID int64) bool {
	secCtx, ok := security.FromContext(r.Context())
	if !ok || secCtx.Name() != "v2token" {
		return false
	}
	if !secCtx.Can(r.Context(), rbac.ActionPush, project.NewNamespace(projectID).Resource(rbac.ResourceRepository)) {
		return false
	}
	ua := r.UserAgent()
	return strings.Contains(ua, "cosign") || strings.Contains(ua, "notation")
}
