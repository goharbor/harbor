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
	"reflect"
)

// BasicAuthHandler handle the basic auth mode.
type BasicAuthHandler struct {
	*BaseHandler
}

// Mode implements @Handler.Mode
func (b *BasicAuthHandler) Mode() string {
	return AuthModeBasic
}

// Authorize implements @Handler.Authorize
func (b *BasicAuthHandler) Authorize(req *http.Request, cred *Credential) error {
	if err := b.BaseHandler.Authorize(req, cred); err != nil {
		return err
	}

	if len(cred.Data) == 0 {
		return errors.New("missing username and/or password")
	}

	key := reflect.ValueOf(cred.Data).MapKeys()[0].String()
	req.SetBasicAuth(key, cred.Data[key])

	return nil
}
