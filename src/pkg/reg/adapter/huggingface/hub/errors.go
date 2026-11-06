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

package hub

import (
	"io"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
)

const maxErrorBody = 1024

// statusError converts a non-2xx Hub response into a lib/errors error. It consumes the body.
func statusError(resp *http.Response, hasToken bool) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	msg := resp.Header.Get("X-Error-Message")
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}
	code := codeOf(resp.StatusCode)
	// The Hub answers an anonymous request for a missing or private model with a bare 401
	// ("Invalid username or password."). Reporting it as not found lets the proxy cache return
	// 404 and honour ProxyCacheLocalOnNotFound. Gated models carry X-Error-Code GatedRepo.
	if resp.StatusCode == http.StatusUnauthorized && !hasToken && resp.Header.Get("X-Error-Code") == "" {
		code = errors.NotFoundCode
		msg = "not found, or private and requires a token: " + msg
	}
	return errors.New(nil).WithCode(code).
		WithMessagef("hugging face hub %s %s: http status code: %d, error code: %q, message: %s",
			resp.Request.Method, resp.Request.URL.Path, resp.StatusCode, resp.Header.Get("X-Error-Code"), msg)
}

func codeOf(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return errors.UnAuthorizedCode
	case http.StatusForbidden:
		return errors.ForbiddenCode
	case http.StatusNotFound:
		return errors.NotFoundCode
	case http.StatusTooManyRequests:
		return errors.RateLimitCode
	default:
		return errors.GeneralCode
	}
}
