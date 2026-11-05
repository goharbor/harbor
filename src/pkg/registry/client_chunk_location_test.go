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

	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

// TestPushBlobChunkRejectsCrossHostLocation reproduces the chunk Location redirect: a malicious upstream
// registry answers a chunk upload with a Location pointing at an unrelated host, and Harbor's
// registry client reused it verbatim for the next PATCH/PUT — sending blob bytes and the stored
// registry credential to the attacker host. After the fix the cross-origin Location is rejected and
// the attacker server is never contacted.
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

	// chunk 1 (start==0): initiate + PATCH to upstream; upstream returns attacker as next Location,
	// which is validated against the server that issued it and rejected right there.
	loc, _, err := c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("AB"), 0, 1, "")
	if err == nil {
		// defensive: should the Location ever be handed back, the next chunk must refuse it too.
		_, _, err = c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("CD"), 2, 3, loc)
	}

	if got := atomic.LoadInt64(&attackerHits); got != 0 {
		t.Fatalf("attacker host was contacted %d time(s) with auth %q; cross-host upload Location must be refused (SSRF + credential leak)",
			got, attackerAuth.Load())
	}
	if err == nil {
		t.Fatalf("cross-host Location must fail closed, got nil error (location=%q)", loc)
	}
}

// TestPushBlobChunkFollowsSameOriginRedirect covers a registry that answers the upload POST with a
// same-origin 307 to a different path and then returns a path-relative Location. The Location must
// resolve against the URL that actually answered (the redirect target), not the original request URL.
func TestPushBlobChunkFollowsSameOriginRedirect(t *testing.T) {
	var paths []string
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/repo/blobs/uploads/":
			http.Redirect(w, r, "/sessions/uuid", http.StatusTemporaryRedirect)
		case r.Method == http.MethodPost && r.URL.Path == "/sessions/uuid":
			w.Header().Set("Location", "?_state=s") // path-relative: must resolve against /sessions/uuid
			w.WriteHeader(http.StatusAccepted)
		case r.Method == http.MethodPatch:
			w.Header().Set("Location", r.URL.Path+"?_state=s2")
			w.WriteHeader(http.StatusAccepted)
		default:
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer registry.Close()

	c := NewClientWithAuthorizer(registry.URL, nil, true, "")
	loc, _, err := c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("AB"), 0, 1, "")
	if err != nil {
		t.Fatalf("first chunk: %v", err)
	}
	if _, _, err = c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("CD"), 2, 3, loc); err != nil {
		t.Fatalf("last chunk: %v", err)
	}
	want := []string{
		"/v2/repo/blobs/uploads/",
		"/sessions/uuid",
		"/sessions/uuid?_state=s",
		"/sessions/uuid?_state=s2&digest=sha256%3Adeadbeef",
	}
	if strings.Join(paths, "\n") != strings.Join(want, "\n") {
		t.Fatalf("request sequence mismatch:\n got: %q\nwant: %q", paths, want)
	}
}

// TestPushBlobChunkRejectsCrossOriginRedirect: a redirect that moves the upload to another origin must
// not turn that origin into a trusted upload destination.
func TestPushBlobChunkRejectsCrossOriginRedirect(t *testing.T) {
	var otherHits int64
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&otherHits, 1)
		w.Header().Set("Location", "/v2/repo/blobs/uploads/uuid")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer other.Close()
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/v2/repo/blobs/uploads/", http.StatusTemporaryRedirect)
	}))
	defer registry.Close()

	c := NewClientWithAuthorizer(registry.URL, nil, true, "")
	_, _, err := c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("AB"), 0, 1, "")
	if err == nil || !strings.Contains(err.Error(), "different origin") {
		t.Fatalf("cross-origin redirect must be rejected, got %v", err)
	}
	// Go follows the redirect (one hit) but the Location it returned must never be used.
	if got := atomic.LoadInt64(&otherHits); got > 1 {
		t.Fatalf("other origin contacted %d times; must not be used as upload destination", got)
	}
}

// hostRewritingAuthorizer mimics the AWS ECR authorizer (pkg/reg/adapter/awsecr/auth.go): the adapter
// is configured with the API endpoint but every request is rewritten to the real registry host.
type hostRewritingAuthorizer struct{ host string }

func (a *hostRewritingAuthorizer) Modify(req *http.Request) error {
	req.Host = a.host
	req.URL.Host = a.host
	return nil
}

// TestPushBlobAuthorizerRewritesHost reproduces the ECR regression: the registry answers with an
// absolute Location on the rewritten host, which differs from the configured endpoint but is exactly
// the server the client talked to, so it must be accepted.
func TestPushBlobAuthorizerRewritesHost(t *testing.T) {
	var registry *httptest.Server
	var hits int64
	registry = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		switch r.Method {
		case http.MethodPost, http.MethodPatch:
			w.Header().Set("Location", registry.URL+"/v2/repo/blobs/uploads/uuid?_state=s")
			w.WriteHeader(http.StatusAccepted)
		default:
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer registry.Close()

	u, _ := url.Parse(registry.URL)
	// configured endpoint is a different host (api.ecr.<region>...) that is never contacted
	c := NewClientWithAuthorizer("http://api.ecr.us-east-2.example", &hostRewritingAuthorizer{host: u.Host}, true, "")

	loc, _, err := c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("AB"), 0, 1, "")
	if err != nil {
		t.Fatalf("first chunk: %v", err)
	}
	if _, _, err = c.PushBlobChunk("repo", "sha256:deadbeef", 4, strings.NewReader("CD"), 2, 3, loc); err != nil {
		t.Fatalf("last chunk: %v", err)
	}
	if err = c.PushBlob("repo", "sha256:deadbeef", 4, strings.NewReader("ABCD")); err != nil {
		t.Fatalf("monolithic push: %v", err)
	}
	// POST+PATCH+PUT for the chunked push, POST+PUT for the monolithic one
	if got := atomic.LoadInt64(&hits); got != 5 {
		t.Fatalf("expected 5 requests to the rewritten host, got %d", got)
	}
}

// TestResolveUploadLocation_OriginRules pins the origin rules the fix enforces.
func TestResolveUploadLocation_OriginRules(t *testing.T) {
	const endpoint = "http://good-registry.example:5000"
	origin, _ := url.Parse(endpoint + "/v2/repo/blobs/uploads/")
	rewritten, _ := url.Parse("http://real-registry.example:5000/v2/repo/blobs/uploads/")
	cases := []struct {
		name     string
		origin   *url.URL
		location string
		wantErr  bool
		want     string
	}{
		{"relative resolves to endpoint", origin, "/v2/repo/blobs/uploads/uuid?_state=s", false, endpoint + "/v2/repo/blobs/uploads/uuid?_state=s"},
		{"same-origin absolute allowed", origin, endpoint + "/v2/repo/blobs/uploads/uuid", false, endpoint + "/v2/repo/blobs/uploads/uuid"},
		{"cross-host absolute rejected", origin, "http://evil.attacker.com/steal", true, ""},
		{"cross-port absolute rejected", origin, "http://good-registry.example:9999/steal", true, ""},
		{"scheme-upgrade absolute rejected", origin, "https://good-registry.example:5000/steal", true, ""},
		{"userinfo smuggling rejected", origin, "http://good-registry.example:5000@evil.attacker.com/steal", true, ""},
		{"network-path reference rejected", origin, "//evil.attacker.com/steal", true, ""},
		// authorizer rewrote the host (ECR): Location on the rewritten host is accepted, endpoint host is not
		{"rewritten host accepted", rewritten, "http://real-registry.example:5000/v2/repo/blobs/uploads/uuid", false, "http://real-registry.example:5000/v2/repo/blobs/uploads/uuid"},
		{"endpoint host rejected when request went elsewhere", rewritten, endpoint + "/v2/repo/blobs/uploads/uuid", true, ""},
		{"nil origin falls back to endpoint", nil, endpoint + "/v2/repo/blobs/uploads/uuid", false, endpoint + "/v2/repo/blobs/uploads/uuid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveUploadLocation(endpoint, tc.origin, tc.location)
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

// TestResolveUploadLocation_EndpointPathPrefix covers registries reached through a path prefix
// (e.g. behind a reverse proxy). Distribution with relativeurls returns "/v2/..." without the
// prefix, so the prefix must be kept exactly as before the origin check was added.
func TestResolveUploadLocation_EndpointPathPrefix(t *testing.T) {
	const endpoint = "https://reg.example.com/prefix"
	origin, _ := url.Parse(endpoint + "/v2/foo/blobs/uploads/")
	cases := []struct {
		name     string
		location string
		wantErr  bool
		want     string
	}{
		{"relative keeps endpoint prefix", "/v2/foo/blobs/uploads/uuid?_state=s", false, "https://reg.example.com/prefix/v2/foo/blobs/uploads/uuid?_state=s"},
		{"same-origin absolute with prefix", "https://reg.example.com/prefix/v2/foo/blobs/uploads/uuid", false, "https://reg.example.com/prefix/v2/foo/blobs/uploads/uuid"},
		{"network-path reference rejected", "//evil.attacker.com/steal", true, ""},
		{"cross-host absolute rejected", "https://evil.attacker.com/prefix/v2/foo", true, ""},
		{"userinfo smuggling rejected", "https://reg.example.com@evil.attacker.com/prefix/v2/foo", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveUploadLocation(endpoint, origin, tc.location)
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

// TestBuildBlobUploadURL covers the digest query handling and the structural checks on the
// (already validated) location.
func TestBuildBlobUploadURL(t *testing.T) {
	const loc = "https://reg.example.com/prefix/v2/foo/blobs/uploads/uuid?_state=s"
	got, err := buildChunkBlobUploadURL(loc, "sha256:deadbeef", true)
	if err != nil || got != loc+"&digest=sha256%3Adeadbeef" {
		t.Fatalf("chunk last: got %q, %v", got, err)
	}
	got, err = buildChunkBlobUploadURL(loc, "sha256:deadbeef", false)
	if err != nil || got != loc {
		t.Fatalf("chunk non-last: got %q, %v", got, err)
	}
	got, err = buildMonolithicBlobUploadURL(loc, "sha256:deadbeef")
	if err != nil || got != loc+"&digest=sha256%3Adeadbeef" {
		t.Fatalf("monolithic: got %q, %v", got, err)
	}
	for _, bad := range []string{"", "/v2/relative", "//evil/steal", "https://a@evil/steal"} {
		if _, err := buildChunkBlobUploadURL(bad, "sha256:deadbeef", true); err == nil {
			t.Fatalf("expected rejection for %q", bad)
		}
	}
}
