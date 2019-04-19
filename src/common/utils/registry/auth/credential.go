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
	"net/http"

	"github.com/goharbor/harbor/src/common/http/modifier"
)

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type basicAuthCredential struct {
	username string
	password string
}

// NewBasicAuthCredential ...
func NewBasicAuthCredential(username, password string) Credential {
	return &basicAuthCredential{
		username: username,
		password: password,
	}
}

func (b *basicAuthCredential) AddAuthorization(req *http.Request) {
	// only add the authentication info when the username isn't empty
	// the logic is needed for requesting resources from docker hub's
	// public repositories
	if len(b.username) > 0 {
		req.SetBasicAuth(b.username, b.password)
	}
}

// implement github.com/goharbor/harbor/src/common/http/modifier.Modifier
func (b *basicAuthCredential) Modify(req *http.Request) error {
	b.AddAuthorization(req)
	return nil
}
