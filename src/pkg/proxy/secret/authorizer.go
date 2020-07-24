//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package secret

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/lib"
)

const (
	secretPrefix = "Proxy-Cache-Secret"
)

// NewAuthorizer returns an instance of the authorizer
func NewAuthorizer() lib.Authorizer {
	return &authorizer{}
}

type authorizer struct{}

func (s *authorizer) Modify(req *http.Request) error {
	if req == nil {
		return errors.New("the request is null")
	}
	repository, _, ok := lib.MatchManifestURLPattern(req.URL.Path)
	if !ok {
		repository, _, ok = lib.MatchBlobURLPattern(req.URL.Path)
		if !ok {
			repository, ok = lib.MatchBlobUploadURLPattern(req.URL.Path)
			if !ok {
				return nil
			}
		}
	}
	secret := GetManager().Generate(repository)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", secretPrefix, secret))
	return nil
}

// GetSecret gets the secret from the request authorization header
func GetSecret(req *http.Request) string {
	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, secretPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(auth, secretPrefix))
}
