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

package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

// TestGetAndIteratePaginationCrossOriginLinkRejected reproduces the pagination Link SSRF:
// an attacker-controlled upstream returns a Link: rel="next" header whose value
// ("@<host>/...") makes the naive scheme+host+link concatenation re-parse with the
// attacker's host as the network host. The credential-bearing client must NOT follow
// that link to a second origin, and must NOT leak the configured Basic Authorization
// header to it.
func TestGetAndIteratePaginationCrossOriginLinkRejected(t *testing.T) {
	var attackerHits int32
	var leakedAuth atomic.Value
	leakedAuth.Store("")

	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attackerHits, 1)
		leakedAuth.Store(r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`["attacker-controlled"]`))
	}))
	defer attacker.Close()

	attackerHost := mustHost(t, attacker.URL) // 127.0.0.1:<attacker port>

	// Each vector, when concatenated by the vulnerable code or resolved naively, re-points
	// the request at attackerHost (a different origin than the upstream server).
	vectors := map[string]string{
		"userinfo_concat":      "@" + attackerHost + "/harvest",       // scheme://up@attacker/harvest
		"protocol_relative":    "//" + attackerHost + "/harvest",      // protocol-relative to attacker
		"absolute_crossorigin": "http://" + attackerHost + "/harvest", // absolute cross-origin
	}

	for name, linkValue := range vectors {
		t.Run(name, func(t *testing.T) {
			atomic.StoreInt32(&attackerHits, 0)
			leakedAuth.Store("")

			// Emit the malicious "next" link only on the first response so that any
			// same-origin resolution (e.g. the vulnerable concat mangling "//host" back
			// onto the upstream host) terminates instead of looping; the invariant under
			// test is solely that the attacker origin is never reached.
			var served int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if atomic.AddInt32(&served, 1) == 1 {
					w.Header().Set("Link", "<"+linkValue+`>; rel="next"`)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`["repo-a"]`))
			}))
			defer upstream.Close()

			// The client carries the remote registry's stored Basic credentials, exactly as
			// the harbor replication adapter builds it (base/adapter.go).
			client := NewClient(&http.Client{Transport: GetHTTPTransport()},
				basic.NewAuthorizer("robot$replication", "sup3r-s3cret"))

			var repos []string
			err := client.GetAndIteratePagination(upstream.URL+"/v2/_catalog", &repos)

			if got := atomic.LoadInt32(&attackerHits); got != 0 {
				t.Fatalf("SSRF via %s: credential-bearing client followed the cross-origin Link to the attacker host %d time(s); leaked Authorization=%q",
					name, got, leakedAuth.Load())
			}
			// Silently stopping would hand callers a truncated list as if it were complete.
			if err == nil {
				t.Fatalf("cross-origin Link via %s must fail pagination, got nil error and partial result %v", name, repos)
			}
		})
	}
}

// TestGetAndIteratePaginationSameOriginLinkFollowed is the backward-compatibility arm:
// legitimate relative and absolute "next" links on the same origin must still paginate.
func TestGetAndIteratePaginationSameOriginLinkFollowed(t *testing.T) {
	var page int32
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch atomic.AddInt32(&page, 1) {
		case 1:
			w.Header().Set("Link", `</v2/_catalog?last=repo-a&n=1>; rel="next"`)
			w.Write([]byte(`["repo-a"]`))
		case 2:
			w.Header().Set("Link", "<"+srv.URL+`/v2/_catalog?last=repo-b&n=1>; rel="next"`)
			w.Write([]byte(`["repo-b"]`))
		default:
			w.Write([]byte(`["repo-c"]`))
		}
	}))
	defer srv.Close()

	client := NewClient(&http.Client{Transport: GetHTTPTransport()},
		basic.NewAuthorizer("robot$replication", "sup3r-s3cret"))

	var repos []string
	if err := client.GetAndIteratePagination(srv.URL+"/v2/_catalog", &repos); err != nil {
		t.Fatalf("same-origin pagination failed: %v", err)
	}
	if len(repos) != 3 || repos[0] != "repo-a" || repos[1] != "repo-b" || repos[2] != "repo-c" {
		t.Fatalf("expected all pages aggregated, got %v", repos)
	}
}

func TestResolveNextLink(t *testing.T) {
	base, err := url.Parse("http://registry.example.com/v2/_catalog?n=1")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		link    string
		want    string
		wantErr string
	}{
		{name: "relative", link: "/v2/_catalog?last=a", want: "http://registry.example.com/v2/_catalog?last=a"},
		{name: "absolute same origin", link: "http://registry.example.com/v2/_catalog?last=a", want: "http://registry.example.com/v2/_catalog?last=a"},
		{name: "explicit default port", link: "http://registry.example.com:80/v2/_catalog?last=a", want: "http://registry.example.com:80/v2/_catalog?last=a"},
		{name: "host case insensitive", link: "http://REGISTRY.example.com/v2/_catalog?last=a", want: "http://REGISTRY.example.com/v2/_catalog?last=a"},
		{name: "same origin userinfo", link: "//robot@registry.example.com/v2/_catalog", wantErr: "userinfo"},
		{name: "absolute userinfo", link: "http://robot:pw@registry.example.com/v2/_catalog", wantErr: "userinfo"},
		{name: "other host", link: "http://evil.example.com/v2/_catalog", wantErr: "different origin"},
		{name: "protocol relative other host", link: "//evil.example.com/v2/_catalog", wantErr: "different origin"},
		{name: "other port", link: "http://registry.example.com:8080/v2/_catalog", wantErr: "different origin"},
		{name: "other scheme", link: "https://registry.example.com/v2/_catalog", wantErr: "different origin"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			next, err := resolveNextLink(base, c.link)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("expected error containing %q, got %v (next=%v)", c.wantErr, err, next)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := next.String(); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func mustHost(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return u.Host
}
