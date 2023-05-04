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

// Handler defines how to add authorization data to the requests
// depending on the different auth modes.
type Handler interface {
	// Append authorization data to the request depends on cred modes.
	//
	// If everything is ok, nil error will be returned.
	// Otherwise, an error will be got.
	Authorize(req *http.Request, cred *Credential) error

	// Mode returns the auth mode identity.
	Mode() string
}

// BaseHandler provides some basic functions like validation.
type BaseHandler struct{}

// Mode implements @Handler.Mode
func (b *BaseHandler) Mode() string {
	return "BASE"
}

// Authorize implements @Handler.Authorize
func (b *BaseHandler) Authorize(req *http.Request, cred *Credential) error {
	if req == nil {
		return errors.New("nil request cannot be authorized")
	}

	if cred == nil || cred.Data == nil {
		return errors.New("no credential data provided")
	}

	return nil
}
