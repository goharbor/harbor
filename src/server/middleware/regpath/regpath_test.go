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

package regpath

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type probe struct {
	called bool
	path   string
}

// probeStatus is deliberately not 200, so a pass-through assertion proves the
// next handler's response reached the client rather than the recorder default.
const probeStatus = http.StatusAccepted

func (p *probe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.called = true
	p.path = r.URL.Path
	w.WriteHeader(probeStatus)
}

func TestMiddleware(t *testing.T) {
	cases := []struct {
		name        string
		path        string
		wantBlocked bool
	}{
		// The advisory PoC: authorizes as "public/x" but path.Clean resolves it
		// to the private "private/secret-image". Must be blocked.
		{"manifest traversal (advisory PoC)", "/v2/public/x/manifests/x/../../../../private/secret-image/manifests/1.0.0", true},
		{"traversal to digest", "/v2/pub/x/manifests/x/../../../priv/secret/manifests/sha256:abc", true},
		{"single dot segment", "/v2/pub/./blobs/uploads/", true},
		{"leading dotdot", "/v2/../v2/priv/secret/manifests/latest", true},
		// Legitimate distribution requests must pass untouched.
		{"plain manifest by tag", "/v2/library/ubuntu/manifests/latest", false},
		{"blob upload trailing slash", "/v2/library/ubuntu/blobs/uploads/", false},
		{"base endpoint", "/v2/", false},
		{"catalog", "/v2/_catalog", false},
		{"tag with dots is not a dot segment", "/v2/library/ubuntu/manifests/1.0.0", false},
		// Out of scope: non-/v2 paths are never touched, even with dot segments.
		{"non-v2 path with dotdot untouched", "/api/v2.0/../whatever", false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := &probe{}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://harbor.local"+tt.path, nil)
			// httptest.NewRequest normalizes via url.Parse but preserves dot
			// segments in Path; guard against client-side cleaning.
			req.URL.Path = tt.path
			Middleware()(p).ServeHTTP(rec, req)
			if tt.wantBlocked {
				assert.False(t, p.called, "next handler must not run for %q", tt.path)
				assert.Equal(t, http.StatusBadRequest, rec.Code, "want 400 for %q", tt.path)
			} else {
				assert.True(t, p.called, "next handler must run for %q", tt.path)
				assert.Equal(t, probeStatus, rec.Code, "want pass-through for %q", tt.path)
			}
		})
	}
}

// Percent-encoded dot segments are decoded by net/url before the middleware
// sees r.URL.Path, so they must be rejected exactly like literal ones. These
// requests go through httptest.NewRequest without overriding Path, which is the
// same url.Parse path the server uses.
func TestMiddlewareEncodedDotSegments(t *testing.T) {
	cases := []struct {
		name        string
		target      string
		wantBlocked bool
	}{
		{"encoded dot", "/v2/pub/%2e/blobs/uploads/", true},
		{"encoded dotdot lowercase", "/v2/pub/x/manifests/x/%2e%2e/%2e%2e/%2e%2e/%2e%2e/priv/secret/manifests/1.0.0", true},
		{"encoded dotdot uppercase", "/v2/pub/x/manifests/x/%2E%2E/%2E%2E/%2E%2E/%2E%2E/priv/secret/manifests/1.0.0", true},
		{"mixed case and half encoded", "/v2/pub/x/manifests/x/%2E%2e/.%2e/%2e./%2E%2E/priv/secret/manifests/1.0.0", true},
		{"encoded dots inside a tag are fine", "/v2/library/ubuntu/manifests/1%2e0%2e0", false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p := &probe{}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://harbor.local"+tt.target, nil)
			Middleware()(p).ServeHTTP(rec, req)
			if tt.wantBlocked {
				assert.False(t, p.called, "next handler must not run for %q (decoded %q)", tt.target, req.URL.Path)
				assert.Equal(t, http.StatusBadRequest, rec.Code)
			} else {
				assert.True(t, p.called, "next handler must run for %q", tt.target)
				assert.Equal(t, probeStatus, rec.Code)
			}
		})
	}
}

func TestHasDotSegment(t *testing.T) {
	assert.True(t, hasDotSegment("/v2/a/../b"))
	assert.True(t, hasDotSegment("/v2/./b"))
	assert.True(t, hasDotSegment(".."))
	assert.True(t, hasDotSegment("a/b/.."))
	assert.False(t, hasDotSegment("/v2/library/ubuntu/manifests/1.0.0"))
	assert.False(t, hasDotSegment("/v2/library/ubuntu/blobs/uploads/"))
	assert.False(t, hasDotSegment("..foo/f..oo/foo.."))
}
