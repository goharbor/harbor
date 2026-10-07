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
	"strings"
	"sync/atomic"
	"testing"

	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

// TestPushBlobChunkRejectsCrossHostLocation reproduces the chunk Location redirect: a malicious upstream
// registry answers a chunk upload with a Location pointing at an unrelated host, and Harbor's
// registry client reused it verbatim for the next PATCH/PUT — sending blob bytes and the stored
// registry credential to the attacker host. After the fix, buildChunkBlobUploadURL rejects the
// cross-origin Location and the attacker server is never contacted.
func TestPushBlobChunkRejectsCrossHostLocation(t *testing.T) {
	var attackerHits int64
	var attackerAuth atomic.Value
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&attackerHits, 1)
		attackerAuth.Store(r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusCreated)
	}))
	defer attacker.Close()

	var upstream *httptest.Server
	upstream = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost: // initiate: hand back a same-host upload session
			w.Header().Set("Location", upstream.URL+"/v2/repo/blobs/uploads/uuid?_state=s")
			w.WriteHeader(http.StatusAccepted)
		case http.MethodPatch: // first chunk: return a CROSS-HOST Location for the next chunk
			w.Header().Set("Location", attacker.URL+"/v2/evil")
			w.WriteHeader(http.StatusAccepted)
		default:
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer upstream.Close()

	c := NewClientWithAuthorizer(upstream.URL, basic.NewAuthorizer("robot$replication", "sup3r-s3cret"), true, "")

	// chunk 1 (start==0): initiate + PATCH to upstream; upstream returns attacker as next Location.
	loc, _, err := c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("AB"), 0, 1, "")
	if err != nil {
		t.Fatalf("first chunk should succeed against the legitimate upstream: %v", err)
	}

	// chunk 2 (last): the client is asked to reuse the attacker Location returned above.
	_, _, err = c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("CD"), 2, 3, loc)

	if got := atomic.LoadInt64(&attackerHits); got != 0 {
		t.Fatalf("attacker host was contacted %d time(s) with auth %q; cross-host upload Location must be refused (SSRF + credential leak)",
			got, attackerAuth.Load())
	}
	if err == nil {
		t.Fatalf("second chunk against a cross-host Location must fail closed, got nil error (location=%q)", loc)
	}
}

// TestBuildChunkBlobUploadURL_LocationOrigin pins the origin rules the fix enforces.
func TestBuildChunkBlobUploadURL_LocationOrigin(t *testing.T) {
	const endpoint = "http://good-registry.example:5000"
	cases := []struct {
		name     string
		location string
		wantErr  bool
		wantHost string
	}{
		{"relative resolves to endpoint", "/v2/repo/blobs/uploads/uuid?_state=s", false, "good-registry.example:5000"},
		{"same-origin absolute allowed", "http://good-registry.example:5000/v2/repo/blobs/uploads/uuid", false, "good-registry.example:5000"},
		{"cross-host absolute rejected", "http://evil.attacker.com/steal", true, ""},
		{"cross-port absolute rejected", "http://good-registry.example:9999/steal", true, ""},
		{"scheme-upgrade absolute rejected", "https://good-registry.example:5000/steal", true, ""},
		{"userinfo smuggling rejected", "http://good-registry.example:5000@evil.attacker.com/steal", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildChunkBlobUploadURL(endpoint, tc.location, "sha256:deadbeef", true)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected rejection, got url=%q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(got, tc.wantHost) {
				t.Fatalf("resolved URL %q does not target expected host %q", got, tc.wantHost)
			}
		})
	}
}

// TestBuildBlobUploadURL_EndpointPathPrefix covers registries reached through a path prefix
// (e.g. behind a reverse proxy). Distribution with relativeurls returns "/v2/..." without the
// prefix, so the prefix must be kept exactly as before the origin check was added.
func TestBuildBlobUploadURL_EndpointPathPrefix(t *testing.T) {
	const endpoint = "https://reg.example.com/prefix"
	cases := []struct {
		name     string
		location string
		wantErr  bool
		want     string
	}{
		{"relative keeps endpoint prefix", "/v2/foo/blobs/uploads/uuid?_state=s", false,
			"https://reg.example.com/prefix/v2/foo/blobs/uploads/uuid?_state=s&digest=sha256%3Adeadbeef"},
		{"same-origin absolute with prefix", "https://reg.example.com/prefix/v2/foo/blobs/uploads/uuid", false,
			"https://reg.example.com/prefix/v2/foo/blobs/uploads/uuid?digest=sha256%3Adeadbeef"},
		{"network-path reference rejected", "//evil.attacker.com/steal", true, ""},
		{"cross-host absolute rejected", "https://evil.attacker.com/prefix/v2/foo", true, ""},
		{"userinfo smuggling rejected", "https://reg.example.com@evil.attacker.com/prefix/v2/foo", true, ""},
	}
	builders := map[string]func(string) (string, error){
		"chunk": func(loc string) (string, error) {
			return buildChunkBlobUploadURL(endpoint, loc, "sha256:deadbeef", true)
		},
		"monolithic": func(loc string) (string, error) {
			return buildMonolithicBlobUploadURL(endpoint, loc, "sha256:deadbeef")
		},
	}
	for bname, build := range builders {
		for _, tc := range cases {
			t.Run(bname+"/"+tc.name, func(t *testing.T) {
				got, err := build(tc.location)
				if tc.wantErr {
					if err == nil {
						t.Fatalf("expected rejection, got url=%q", got)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Fatalf("got %q, want %q", got, tc.want)
				}
			})
		}
	}
}
