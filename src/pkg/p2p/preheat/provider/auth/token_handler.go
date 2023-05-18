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
	"fmt"
	"net/http"
)

// TokenAuthHandler handles the OAuth auth mode.
type TokenAuthHandler struct {
	*BaseHandler
}

// Mode implements @Handler.Mode
func (t *TokenAuthHandler) Mode() string {
	return AuthModeOAuth
}

// Authorize implements @Handler.Authorize
func (t *TokenAuthHandler) Authorize(req *http.Request, cred *Credential) error {
	if err := t.BaseHandler.Authorize(req, cred); err != nil {
		return err
	}

	if _, ok := cred.Data["token"]; !ok {
		return errors.New("missing OAuth token")
	}

	authData := fmt.Sprintf("%s %s", "Bearer", cred.Data["token"])
	req.Header.Set("Authorization", authData)

	return nil
}
