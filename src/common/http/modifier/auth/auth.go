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

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/secret"
)

// Authorizer is a kind of Modifier used to authorize the requests
type Authorizer modifier.Modifier

// SecretAuthorizer authorizes the requests with the specified secret
type SecretAuthorizer struct {
	secret string
}

// NewSecretAuthorizer returns an instance of SecretAuthorizer
func NewSecretAuthorizer(secret string) *SecretAuthorizer {
	return &SecretAuthorizer{
		secret: secret,
	}
}

// Modify the request by adding secret authentication information
func (s *SecretAuthorizer) Modify(req *http.Request) error {
	if req == nil {
		return errors.New("the request is null")
	}
	err := secret.AddToRequest(req, s.secret)
	return err
}
