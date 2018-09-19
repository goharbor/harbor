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

package secret

import (
	"fmt"
	"net/http"
	"strings"
)

// HeaderPrefix is the prefix of the value of Authorization header.
// It has the space.
const HeaderPrefix = "Harbor-Secret "

// FromRequest tries to get Harbor Secret from request header.
// It will return empty string if the reqeust is nil.
func FromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	auth := req.Header.Get("Authorization")
	if strings.HasPrefix(auth, HeaderPrefix) {
		return strings.TrimPrefix(auth, HeaderPrefix)
	}
	return ""
}

// AddToRequest add the secret to request
func AddToRequest(req *http.Request, secret string) error {
	if req == nil {
		return fmt.Errorf("input request is nil, unable to set secret")
	}
	req.Header.Set("Authorization", fmt.Sprintf("%s%s", HeaderPrefix, secret))
	return nil
}
