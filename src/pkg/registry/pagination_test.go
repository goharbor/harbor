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

package registry

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

// paginatingRegistry serves one page of catalog and tag results whose Link header is produced by
// next for the requested path, then a last page for any request carrying a "last" query.
func paginatingRegistry(t *testing.T, next func(self, path string) string) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("last") != "" {
			if strings.HasSuffix(r.URL.Path, "/tags/list") {
				_, _ = w.Write([]byte(`{"tags":["2.0"]}`))
				return
			}
			_, _ = w.Write([]byte(`{"repositories":["b"]}`))
			return
		}
		if link := next(srv.URL, r.URL.Path); link != "" {
			w.Header().Set("Link", "<"+link+`>; rel="next"`)
		}
		if strings.HasSuffix(r.URL.Path, "/tags/list") {
			_, _ = w.Write([]byte(`{"tags":["1.0"]}`))
			return
		}
		_, _ = w.Write([]byte(`{"repositories":["a"]}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// attacker records every request and the Authorization header it received.
func attacker(t *testing.T) (*httptest.Server, *atomic.Int32, *atomic.Value) {
	t.Helper()
	var hits atomic.Int32
	var auth atomic.Value
	auth.Store("")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		auth.Store(r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"repositories":["x"],"tags":["x"]}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &hits, &auth
}

func TestPaginationRejectsCrossOriginLink(t *testing.T) {
	evil, hits, auth := attacker(t)
	evilHost := strings.TrimPrefix(evil.URL, "http://")
	cases := map[string]func(self, path string) string{
		"absolute other host": func(_, path string) string { return evil.URL + path + "?last=a" },
		"userinfo concat":     func(_, path string) string { return "@" + evilHost + path + "?last=a" },
		"same host other port": func(self, path string) string {
			u, _ := url.Parse(self)
			return "http://" + u.Hostname() + ":1" + path + "?last=a"
		},
	}
	for name, next := range cases {
		t.Run(name, func(t *testing.T) {
			hits.Store(0)
			srv := paginatingRegistry(t, next)
			// basic.NewAuthorizer attaches credentials to every host, like the JFrog adapter's client.
			cli := NewClientWithAuthorizer(srv.URL, basic.NewAuthorizer("robot", "s3cret"), true, "")

			_, err := cli.Catalog()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "leaves the registry origin")
			_, err = cli.ListTags("library/a")
			require.Error(t, err)

			assert.Zero(t, hits.Load(), "attacker origin must not be contacted")
			assert.Empty(t, auth.Load())
		})
	}
}

func TestPaginationFollowsSameOriginLinks(t *testing.T) {
	cases := map[string]func(self, path string) string{
		"relative":      func(_, path string) string { return path + "?last=a&n=1000" },
		"absolute same": func(self, path string) string { return self + path + "?last=a&n=1000" },
	}
	for name, next := range cases {
		t.Run(name, func(t *testing.T) {
			srv := paginatingRegistry(t, next)
			cli := NewClientWithAuthorizer(srv.URL, basic.NewAuthorizer("robot", "s3cret"), true, "")
			repos, err := cli.Catalog()
			require.NoError(t, err)
			assert.Equal(t, []string{"a", "b"}, repos)
			tags, err := cli.ListTags("library/a")
			require.NoError(t, err)
			assert.Equal(t, []string{"1.0", "2.0"}, tags)
		})
	}
}

func TestNextPageURL(t *testing.T) {
	cases := []struct {
		name, endpoint, link, want string
	}{
		{"relative appended to path prefix", "https://reg.example.com/prefix", "/v2/_catalog?last=a", "https://reg.example.com/prefix/v2/_catalog?last=a"},
		{"absolute same origin kept verbatim", "https://reg.example.com", "https://reg.example.com/v2/_catalog?last=a%2Fb", "https://reg.example.com/v2/_catalog?last=a%2Fb"},
		{"explicit default port", "https://reg.example.com", "https://REG.example.com:443/v2/_catalog?last=a", "https://REG.example.com:443/v2/_catalog?last=a"},
		{"endpoint userinfo preserved", "https://u:p@reg.example.com", "/v2/_catalog?last=a", "https://u:p@reg.example.com/v2/_catalog?last=a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := nextPageURL(tc.endpoint, tc.link)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	rejected := []struct {
		name, endpoint, link string
	}{
		{"userinfo concat", "https://reg.example.com", "@evil.example/v2/_catalog"},
		{"added userinfo same host", "https://reg.example.com", "https://x@reg.example.com/v2/_catalog"},
		{"other host", "https://reg.example.com", "https://evil.example/v2/_catalog"},
		{"scheme downgrade", "https://reg.example.com", "http://reg.example.com/v2/_catalog"},
		{"other port", "https://reg.example.com", "https://reg.example.com:8443/v2/_catalog"},
	}
	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			_, err := nextPageURL(tc.endpoint, tc.link)
			require.Error(t, err)
		})
	}
}

// TestNextPageURLErrorRedactsSecrets keeps credentials and tokens carried in a rejected Link
// out of the returned error, which replication logs.
func TestNextPageURLErrorRedactsSecrets(t *testing.T) {
	for _, link := range []string{
		"https://user:s3cr3t-pass@reg.example.com/v2/_catalog",
		"https://evil.example/v2/_catalog?token=s3cr3t-token",
		"@evil.example/v2/_catalog?token=s3cr3t-token",
	} {
		_, err := nextPageURL("https://reg.example.com", link)
		require.Error(t, err, link)
		assert.NotContains(t, err.Error(), "s3cr3t", link)
	}
}
