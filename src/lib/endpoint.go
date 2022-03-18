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

package lib

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
)

// ValidateHTTPURL checks whether the provided string is a valid HTTP URL.
// If it is, return the URL in format "scheme://host:port" to avoid the SSRF
func ValidateHTTPURL(s string) (string, error) {
	s = strings.Trim(s, " ")
	s = strings.TrimRight(s, "/")
	if len(s) == 0 {
		return "", errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("empty string")
	}
	if !strings.Contains(s, "://") {
		s = "http://" + s
	}
	url, err := url.Parse(s)
	if err != nil {
		return "", errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("invalid URL: %s", err.Error())
	}
	if url.Scheme != "http" && url.Scheme != "https" {
		return "", errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("invalid HTTP scheme: %s", url.Scheme)
	}
	// To avoid SSRF security issue, refer to #3755 for more detail
	return fmt.Sprintf("%s://%s%s", url.Scheme, url.Host, url.Path), nil
}
