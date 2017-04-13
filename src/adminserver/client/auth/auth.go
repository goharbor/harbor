// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
)

// Authorizer authorizes request
type Authorizer interface {
	Authorize(*http.Request) error
}

// NewSecretAuthorizer returns an instance of secretAuthorizer
func NewSecretAuthorizer(cookieName, secret string) Authorizer {
	return &secretAuthorizer{
		cookieName: cookieName,
		secret:     secret,
	}
}

type secretAuthorizer struct {
	cookieName string
	secret     string
}

func (s *secretAuthorizer) Authorize(req *http.Request) error {
	if req == nil {
		return nil
	}

	req.AddCookie(&http.Cookie{
		Name:  s.cookieName,
		Value: s.secret,
	})

	return nil
}
