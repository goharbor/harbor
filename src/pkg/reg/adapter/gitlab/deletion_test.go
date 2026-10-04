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

package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

type deletionResponse struct {
	code int
	body string
	next string
}

// Responses are keyed by method and escaped request URI so tests verify the
// project-path encoding, pagination, authentication, and exact deletion target.
func deletionAdapter(t *testing.T, responses map[string]deletionResponse) (*adapter, func() []string) {
	t.Helper()
	requests := []string{}
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.RequestURI()
		mu.Lock()
		requests = append(requests, key)
		mu.Unlock()
		// The unmodified adapter takes this Registry v2 path. A HEAD succeeds,
		// but DELETE reproduces the unauthorized response reported in #21066.
		if strings.HasPrefix(r.URL.Path, "/v2/") {
			if r.Method == http.MethodHead {
				w.Header().Set("Docker-Content-Digest", "sha256:"+strings.Repeat("a", 64))
				w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"errors":[{"code":"UNAUTHORIZED","message":"authentication required"}]}`)
			}
			return
		}
		if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
			t.Error("missing GitLab API authentication")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		response, ok := responses[key]
		if !ok {
			t.Errorf("unexpected request: %s", key)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-Next-Page", response.next)
		w.WriteHeader(response.code)
		fmt.Fprint(w, response.body)
	}))
	t.Cleanup(server.Close)
	return &adapter{
		Adapter: native.NewAdapterWithAuthorizer(&model.Registry{URL: server.URL}, nil),
		clientGitlabAPI: &Client{
			url: server.URL, token: "test-token", client: common_http.NewClient(server.Client()),
		},
	}, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), requests...)
	}
}

func deletionRoutes() map[string]deletionResponse {
	return map[string]deletionResponse{
		"GET /api/v4/projects/team%2Fsubgroup%2Fproject%2Fimage": {code: 404},
		"GET /api/v4/projects/team%2Fsubgroup%2Fproject": {
			code: 200, body: `{"id":7,"path_with_namespace":"team/subgroup/project"}`,
		},
		"GET /api/v4/projects/7/registry/repositories?per_page=50": {
			code: 200, body: `[{"id":9,"path":"other/project/image"}]`, next: "2",
		},
		"GET /api/v4/projects/7/registry/repositories?page=2&per_page=50": {
			code: 200, body: `[{"id":11,"path":"team/subgroup/project/image"}]`,
		},
	}
}

func TestGitLabDeleteTag(t *testing.T) {
	for _, manifest := range []bool{false, true} {
		t.Run(fmt.Sprintf("manifest=%t", manifest), func(t *testing.T) {
			routes := deletionRoutes()
			routes["DELETE /api/v4/projects/7/registry/repositories/11/tags/v1.0"] = deletionResponse{code: 204}
			a, requests := deletionAdapter(t, routes)
			var err error
			if manifest {
				err = a.DeleteManifest("team/subgroup/project/image", "v1.0")
			} else {
				err = a.DeleteTag("team/subgroup/project/image", "v1.0")
			}
			require.NoError(t, err)
			require.Len(t, requests(), 5)
			require.Equal(t, "DELETE /api/v4/projects/7/registry/repositories/11/tags/v1.0", requests()[4])
		})
	}
}

func TestGitLabDeleteTagHTTPFailures(t *testing.T) {
	for _, code := range []int{401, 403, 404, 500} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			routes := deletionRoutes()
			routes["DELETE /api/v4/projects/7/registry/repositories/11/tags/latest"] = deletionResponse{code: code, body: "private-provider-response"}
			a, _ := deletionAdapter(t, routes)
			err := a.DeleteTag("team/subgroup/project/image", "latest")
			var httpErr *common_http.Error
			require.ErrorAs(t, err, &httpErr)
			require.Equal(t, code, httpErr.Code)
			require.NotContains(t, err.Error(), "private-provider-response")
		})
	}
}

func TestGitLabDeletionLookupFailures(t *testing.T) {
	cases := []struct {
		name string
		key  string
		res  deletionResponse
		code int
	}{
		{"project authentication", "GET /api/v4/projects/team%2Fsubgroup%2Fproject%2Fimage", deletionResponse{code: 401}, 401},
		{"project forbidden", "GET /api/v4/projects/team%2Fsubgroup%2Fproject%2Fimage", deletionResponse{code: 403}, 403},
		{"project server error", "GET /api/v4/projects/team%2Fsubgroup%2Fproject%2Fimage", deletionResponse{code: 500}, 500},
		{"repository error", "GET /api/v4/projects/7/registry/repositories?per_page=50", deletionResponse{code: 403}, 403},
		{"different repository", "GET /api/v4/projects/7/registry/repositories?per_page=50", deletionResponse{code: 200, body: `[{"id":11,"path":"team/subgroup/project/other"}]`}, 404},
		{"different project", "GET /api/v4/projects/team%2Fsubgroup%2Fproject", deletionResponse{code: 200, body: `{"id":7,"path_with_namespace":"other/project"}`}, 0},
		{"invalid JSON", "GET /api/v4/projects/team%2Fsubgroup%2Fproject", deletionResponse{code: 200, body: `{`}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			routes := deletionRoutes()
			routes[tc.key] = tc.res
			a, requests := deletionAdapter(t, routes)
			err := a.DeleteTag("team/subgroup/project/image", "latest")
			require.Error(t, err)
			if tc.code != 0 {
				var httpErr *common_http.Error
				require.ErrorAs(t, err, &httpErr)
				require.Equal(t, tc.code, httpErr.Code)
			}
			for _, request := range requests() {
				require.False(t, strings.HasPrefix(request, "DELETE "))
			}
		})
	}
}

func TestGitLabDeleteDigest(t *testing.T) {
	dig := "sha256:" + strings.Repeat("a", 64)
	for _, failure := range []string{"", "details", "delete", "not found", "disappeared"} {
		t.Run(failure, func(t *testing.T) {
			routes := deletionRoutes()
			routes["GET /api/v4/projects/7/registry/repositories/11/tags?per_page=50"] = deletionResponse{code: 200, body: `[{"name":"v1"},{"name":"other"}]`, next: "2"}
			routes["GET /api/v4/projects/7/registry/repositories/11/tags?page=2&per_page=50"] = deletionResponse{code: 200, body: `[{"name":"latest"}]`}
			for _, tag := range []string{"v1", "latest"} {
				routes["GET /api/v4/projects/7/registry/repositories/11/tags/"+tag] = deletionResponse{code: 200, body: fmt.Sprintf(`{"digest":%q}`, dig)}
				routes["DELETE /api/v4/projects/7/registry/repositories/11/tags/"+tag] = deletionResponse{code: 204}
			}
			routes["GET /api/v4/projects/7/registry/repositories/11/tags/other"] = deletionResponse{code: 200, body: `{"digest":"sha256:other"}`}
			switch failure {
			case "details":
				routes["GET /api/v4/projects/7/registry/repositories/11/tags/latest"] = deletionResponse{code: 403}
			case "delete":
				routes["DELETE /api/v4/projects/7/registry/repositories/11/tags/v1"] = deletionResponse{code: 403}
			case "not found":
				routes["GET /api/v4/projects/7/registry/repositories/11/tags?per_page=50"] = deletionResponse{code: 200, body: `[]`}
			case "disappeared":
				routes["GET /api/v4/projects/7/registry/repositories/11/tags/other"] = deletionResponse{code: 404}
				routes["DELETE /api/v4/projects/7/registry/repositories/11/tags/v1"] = deletionResponse{code: 404}
			}
			a, requests := deletionAdapter(t, routes)
			err := a.DeleteManifest("team/subgroup/project/image", dig)
			var deleted []string
			for _, request := range requests() {
				if strings.HasPrefix(request, "DELETE ") {
					deleted = append(deleted, request)
				}
			}
			if failure == "details" || failure == "delete" || failure == "not found" {
				var httpErr *common_http.Error
				require.ErrorAs(t, err, &httpErr)
				if failure == "not found" {
					require.Equal(t, 404, httpErr.Code)
				} else {
					require.Equal(t, 403, httpErr.Code)
				}
				if failure == "delete" {
					require.Len(t, deleted, 1)
				} else {
					require.Empty(t, deleted)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, []string{
				"DELETE /api/v4/projects/7/registry/repositories/11/tags/v1",
				"DELETE /api/v4/projects/7/registry/repositories/11/tags/latest",
			}, deleted)
		})
	}
}

func TestGitLabDeletionInvalidReferences(t *testing.T) {
	a, requests := deletionAdapter(t, nil)
	for _, repository := range []string{"project", "team//project", "team/../project", "/team/project", "team/project/"} {
		require.Error(t, a.DeleteTag(repository, "latest"))
	}
	for _, tag := range []string{"", ".", "..", "-latest", "../latest", "sha256:invalid", "v1?debug=true", "tag#ref", strings.Repeat("a", 129)} {
		require.Error(t, a.DeleteManifest("team/project", tag))
	}
	gitlabDigest := "sha256:" + strings.Repeat("a", 64)
	require.Error(t, a.DeleteTag("team/project", gitlabDigest))
	require.Empty(t, requests())
}
