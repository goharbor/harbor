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

package auth

import (
	"errors"
	"net/http"
)

// NoneAuthHandler handles the case of no credentail required.
type NoneAuthHandler struct{}

// Mode implements @Handler.Mode
func (nah *NoneAuthHandler) Mode() string {
	return AuthModeNone
}

// Authorize implements @Handler.Authorize
func (nah *NoneAuthHandler) Authorize(req *http.Request, _ *Credential) error {
	if req == nil {
		return errors.New("nil request cannot be authorized")
	}

	// Do nothing
	return nil
}
