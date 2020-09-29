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

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/config"
)

const (
	secretPrefix = "Harbor-Secret"
	authHeader   = "Authorization"
)

// Authenticator defined behaviors of doing auth checking.
type Authenticator interface {
	// Auth incoming request
	//
	// req *http.Request: the incoming request
	//
	// Returns:
	// nil returned if successfully done
	// otherwise an error returned
	DoAuth(req *http.Request) error
}

// SecretAuthenticator implements interface 'Authenticator' based on simple secret.
type SecretAuthenticator struct{}

// DoAuth implements same method in interface 'Authenticator'.
func (sa *SecretAuthenticator) DoAuth(req *http.Request) error {
	if req == nil {
		return errors.New("nil request")
	}

	h := strings.TrimSpace(req.Header.Get(authHeader))
	if utils.IsEmptyStr(h) {
		return fmt.Errorf("header '%s' missing", authHeader)
	}

	if !strings.HasPrefix(h, secretPrefix) {
		return fmt.Errorf("'%s' should start with '%s'", authHeader, secretPrefix)
	}

	secret := strings.TrimSpace(strings.TrimPrefix(h, secretPrefix))
	// incase both two are empty
	if utils.IsEmptyStr(secret) {
		return errors.New("empty secret is not allowed")
	}

	expectedSecret := config.GetUIAuthSecret()
	if expectedSecret != secret {
		return errors.New("unauthorized")
	}

	return nil
}
