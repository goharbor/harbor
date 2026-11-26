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
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/server/middleware"
)

// Middleware rejects OCI distribution ("/v2/") requests whose URL path contains
// a "." or ".." segment.
//
// Harbor authorizes a "/v2/" request by pattern-matching the raw request path
// (server/middleware/v2auth via lib.MatchManifestURLPattern and friends), but
// the beego router resolves the handler after running path.Clean on the same
// path, and the backend registry cleans it again. A request such as
//
//	/v2/public/x/manifests/x/../../../../private/secret/manifests/1.0.0
//
// is authorized as the public repository "public/x" yet dispatched to the
// private repository "private/secret". The authorization decision and the
// served resource therefore disagree.
//
// No legitimate OCI path contains a dot segment (repository, reference and
// digest grammars all forbid a component that is exactly "." or ".."), so the
// safe fix is to reject such requests outright (fail closed) rather than keep
// the two independent normalizations in agreement forever.
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		// Only the distribution ("/v2/") endpoints authorize by matching the raw
		// path, so the auth-vs-routing mismatch is confined to them. Leave every
		// other route untouched.
		if strings.HasPrefix(r.URL.Path, "/v2/") && hasDotSegment(r.URL.Path) {
			lib_http.SendError(w, errors.New("invalid request path: dot segments are not allowed").WithCode(errors.BadRequestCode))
			return
		}
		next.ServeHTTP(w, r)
	}, skippers...)
}

// hasDotSegment reports whether any slash-separated segment of p is "." or "..".
// p is the decoded URL path (net/url has already percent-decoded it), which is
// exactly the value both the auth pattern matcher and the router consume.
func hasDotSegment(p string) bool {
	for len(p) > 0 {
		i := strings.IndexByte(p, '/')
		var seg string
		if i < 0 {
			seg, p = p, ""
		} else {
			seg, p = p[:i], p[i+1:]
		}
		if seg == "." || seg == ".." {
			return true
		}
	}
	return false
}
