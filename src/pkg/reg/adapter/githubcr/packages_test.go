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

package githubcr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

func writePackages(w http.ResponseWriter, names ...string) {
	packages := []*githubPackage{}
	for _, name := range names {
		packages = append(packages, &githubPackage{Name: name})
	}
	_ = json.NewEncoder(w).Encode(packages)
}

// mockPackages is the data of the mock GitHub API. It maps each endpoint to
// its pages of package names. my-org is an org with two pages. someone is a
// user, so its org endpoint does not exist. Other owners do not exist.
var mockPackages = map[string][][]string{
	"/orgs/my-org/packages": {
		{"apps/web-ui", "apps/api", "apps-bundle"},
		{"charts/platform", "apps/base/nginx"},
	},
	"/users/someone/packages": {
		{"tool"},
	},
}

// mockGitHubAPI serves mockPackages like the GitHub Packages API. It returns
// one page for each request and a Link header if there is a next page.
func mockGitHubAPI(t *testing.T) *httptest.Server {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "container", r.URL.Query().Get("package_type"))

		pages, ok := mockPackages[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			page, _ = strconv.Atoi(p)
		}
		if page < 1 || page > len(pages) {
			http.NotFound(w, r)
			return
		}
		if page < len(pages) {
			next := fmt.Sprintf("%s%s?package_type=container&per_page=100&page=%d", server.URL, r.URL.Path, page+1)
			w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="next"`, next))
		}
		writePackages(w, pages[page-1]...)
	}))
	return server
}

func newTestAdapter(apiURL string) *adapter {
	return &adapter{
		client: common_http.NewClient(nil),
		apiURL: apiURL,
	}
}

func Test_listRepositories(t *testing.T) {
	server := mockGitHubAPI(t)
	defer server.Close()
	a := newTestAdapter(server.URL)

	tests := []struct {
		name    string
		pattern string
		want    []string
		wantErr bool
	}{
		{
			name:    "all packages of an org, on two pages",
			pattern: "my-org/**",
			want: []string{
				"my-org/apps/web-ui", "my-org/apps/api", "my-org/apps-bundle",
				"my-org/charts/platform", "my-org/apps/base/nginx",
			},
		},
		{
			name:    "nested repositories only",
			pattern: "my-org/apps/**",
			want:    []string{"my-org/apps/web-ui", "my-org/apps/api", "my-org/apps/base/nginx"},
		},
		{
			name:    "a single star does not cross a slash",
			pattern: "my-org/apps*",
			want:    []string{"my-org/apps-bundle"},
		},
		{
			name:    "user instead of org",
			pattern: "someone/*",
			want:    []string{"someone/tool"},
		},
		{
			name:    "two owners",
			pattern: "{my-org,someone}/t*",
			want:    []string{"someone/tool"},
		},
		{
			name:    "no match",
			pattern: "my-org/nothing/**",
			want:    nil,
		},
		{
			name:    "wildcard owner",
			pattern: "*/apps/**",
			wantErr: true,
		},
		{
			name:    "owner does not exist",
			pattern: "nobody/**",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.listRepositories(tt.pattern)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_listPackages_basicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "user", username)
		assert.Equal(t, "token", password)
		writePackages(w, "a")
	}))
	defer server.Close()
	a := &adapter{
		client: common_http.NewClient(nil, basic.NewAuthorizer("user", "token")),
		apiURL: server.URL,
	}

	packages, err := a.listPackages("my-org")
	require.NoError(t, err)
	require.Len(t, packages, 1)
	assert.Equal(t, "a", packages[0].Name)
}

func Test_listPackages_errors(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name: "bad credentials",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
			},
			wantErr: "Bad credentials",
		},
		{
			name:    "org and user do not exist",
			handler: http.NotFound,
			wantErr: "code 404",
		},
		{
			name: "invalid JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("{"))
			},
			wantErr: "unexpected end of JSON input",
		},
		{
			name: "next page on a different host",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Link", `<https://evil.example.com/orgs/o/packages?page=2>; rel="next"`)
				writePackages(w, "a")
			},
			wantErr: "resolves to a different origin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			_, err := newTestAdapter(server.URL).listPackages("o")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func Test_ownersOf(t *testing.T) {
	tests := []struct {
		pattern string
		want    []string
		wantErr bool
	}{
		{pattern: "my-org/**", want: []string{"my-org"}},
		{pattern: "my-org", want: []string{"my-org"}},
		{pattern: "{a,b}/apps/**", want: []string{"a", "b"}},
		{pattern: "*/apps", wantErr: true},
		{pattern: "**", wantErr: true},
		{pattern: "my-*/x", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got, err := ownersOf(tt.pattern)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
