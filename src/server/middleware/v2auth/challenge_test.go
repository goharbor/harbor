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

package v2auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib"
)

func TestBearerChallenge(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		auth   string
		scope  string
		mount  bool
	}{
		{"pull", http.MethodGet, "Bearer token", "repository:project_1/image:pull", false},
		{"head", http.MethodHead, "Bearer token", "repository:project_1/image:pull", false},
		{"push", http.MethodPut, "Bearer token", "repository:project_1/image:pull,push", false},
		{"upload", http.MethodPost, "Bearer token", "repository:project_1/image:pull,push", false},
		{"patch", http.MethodPatch, "Bearer token", "repository:project_1/image:pull,push", false},
		{"delete", http.MethodDelete, "Bearer token", "repository:project_1/image:delete", false},
		{"mount", http.MethodPost, "Bearer token", "repository:project_1/image:pull,push repository:project_2/image:pull", true},
		{"case_insensitive_scheme", http.MethodGet, "bEaReR token", "repository:project_1/image:pull", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			info := lib.ArtifactInfo{Repository: "project_1/image", ProjectName: "project_1"}
			if tc.mount {
				info.BlobMountRepository = "project_2/image"
				info.BlobMountProjectName = "project_2"
			}
			req := httptest.NewRequest(tc.method, "https://harbor.test/v2/project_1/image/manifests/latest", nil)
			req.Header.Set(authHeader, tc.auth)
			req = req.WithContext(lib.WithArtifactInfo(context.Background(), info))
			require.Equal(t, `Bearer realm="https://harbor.test/service/token",service="harbor-registry",scope="`+tc.scope+`"`, getChallenge(req, accessList(req)))
		})
	}
}

func TestChallengeAuthSchemes(t *testing.T) {
	for _, tc := range []struct {
		name      string
		auth      string
		path      string
		challenge string
	}{
		{"anonymous_login", "", "/v2/", `Bearer realm="https://harbor.test/service/token",service="harbor-registry"`},
		{"bearer_login", "Bearer token", "/v2/", `Bearer realm="https://harbor.test/service/token",service="harbor-registry"`},
		{"basic_login", "Basic dTpw", "/v2/", `Basic realm="harbor"`},
		{"unknown_scheme", "Digest token", "/v2/", `Basic realm="harbor"`},
		{"bearer_prefix_is_not_scheme", "BearerOther token", "/v2/", `Basic realm="harbor"`},
		{"anonymous_catalog", "", "/v2/_catalog", `Basic realm="harbor"`},
		{"bearer_catalog", "Bearer token", "/v2/_catalog", `Basic realm="harbor"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "https://harbor.test"+tc.path, nil)
			req.Header.Set(authHeader, tc.auth)
			require.Equal(t, tc.challenge, getChallenge(req, accessList(req)))
		})
	}
}
