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

package csrf

import (
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/middleware"
)

// protect judges each request on the origin that browser actually used, so it
// needs no configuration and is safe for concurrent use.
var protect = http.NewCrossOriginProtection()

// safeMethods are the methods RFC 7231 defines as safe, which carry no state
// change and so need no cross-origin check.
var safeMethods = []string{http.MethodGet, http.MethodHead, http.MethodOptions}

// check applies the cross-origin protection, then closes the one gap it leaves
// open by design.
//
// CrossOriginProtection admits a request carrying neither Sec-Fetch-Site nor
// Origin, reading it as non-browser traffic. That is the right default for a
// library, but not here: csrfSkipper has already excused the registry, API and
// service routes that non-browser clients use, so an unsafe request arriving
// without either header is one we cannot vouch for. Refusing it matches what the
// previous token scheme did, which rejected any write that failed to present a
// token.
func check(req *http.Request) error {
	if err := protect.Check(req); err != nil {
		// protect.Check compares Origin against Host including the port. Browsers
		// send no Sec-Fetch-Site over plain HTTP, so the check falls to that
		// comparison, and a proxy that rewrites Host drops the port from it — the
		// bundled nginx sets Host to $host — making a legitimate same-origin write
		// look cross-origin. Re-admit exactly that: no Sec-Fetch-Site, and an
		// Origin whose host matches the request Host once the port is set aside.
		if req.Header.Get("Sec-Fetch-Site") == "" && originHostMatchesIgnoringPort(req) {
			return nil
		}
		return err
	}

	if slices.Contains(safeMethods, req.Method) {
		return nil
	}

	if req.Header.Get("Sec-Fetch-Site") == "" && req.Header.Get("Origin") == "" {
		return errors.New("request carries neither Sec-Fetch-Site nor Origin, cannot confirm its origin")
	}

	return nil
}

// originHostMatchesIgnoringPort reports whether the request's Origin names the
// same host as the request targeted, disregarding the port. A differing host —
// a foreign site or a sibling subdomain — never matches, so the cross-site and
// CVE-2025-24358 rejections stand.
func originHostMatchesIgnoringPort(req *http.Request) bool {
	origin := req.Header.Get("Origin")
	if origin == "" {
		return false
	}
	o, err := url.Parse(origin)
	if err != nil || o.Hostname() == "" {
		return false
	}
	return o.Hostname() == hostname(req.Host)
}

// hostname returns the host of a Host header or authority without any port and
// without IPv6 brackets, so it can be compared to url.URL.Hostname.
func hostname(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return strings.Trim(host, "[]")
}

// Middleware rejects cross-origin requests that carry a session. It reads the
// Fetch metadata headers browsers have sent since 2023, falling back to
// comparing Origin against the Host the client addressed.
//
// Deliberately no configured endpoint takes part in the decision. Harbor is
// commonly reachable on several ingresses, over different schemes, behind
// proxies that rewrite Host; a single configured origin describes none of that,
// and every attempt to reconstruct "the" expected origin gets one of those
// deployments wrong.
func Middleware() func(handler http.Handler) http.Handler {
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		if err := check(req); err != nil {
			log.Debugf("Rejected cross-origin request for %s: %v", lib.TrimLineBreaks(req.URL.Path), err)
			lib_http.SendError(rw, errors.New(err).WithCode(errors.ForbiddenCode))
			return
		}

		next.ServeHTTP(rw, req)
	}, csrfSkipper)
}

// csrfSkipper makes sure only some of the uris accessed by non-UI client can skip the csrf check
func csrfSkipper(req *http.Request) bool {
	path := req.URL.Path
	if (strings.HasPrefix(path, "/v2/") ||
		strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/service/")) && !lib.GetCarrySession(req.Context()) {
		return true
	}
	return false
}
