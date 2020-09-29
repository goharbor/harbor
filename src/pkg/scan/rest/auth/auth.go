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
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	authorization = "Authorization"
	// Basic ...
	Basic = "Basic"
	// Bearer ...
	Bearer = "Bearer"
	// APIKey ...
	APIKey = "X-ScannerAdapter-API-Key"
)

// Authorizer defines operation for authorizing the requests
type Authorizer interface {
	Authorize(req *http.Request) error
}

// GetAuthorizer is a factory method for getting an authorizer based on the given auth type
func GetAuthorizer(auth, cred string) (Authorizer, error) {
	switch strings.TrimSpace(auth) {
	// No authorizer required
	case "":
		return NewNoAuth(), nil
	case Basic:
		return NewBasicAuth(cred), nil
	case Bearer:
		return NewBearerAuth(cred), nil
	case APIKey:
		return NewAPIKeyAuthorizer(cred), nil
	default:
		return nil, errors.Errorf("auth type %s is not supported", auth)
	}
}
