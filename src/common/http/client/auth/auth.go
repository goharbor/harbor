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
	"errors"
	"net/http"
)

const (
	secretCookieName = "secret"
)

// Authorizer authorizes the requests
type Authorizer interface {
	Authorize(*http.Request) error
}

// CookieAuthorizer authorizes the requests by adding cookie specified by name and value
type CookieAuthorizer struct {
	name  string
	value string
}

// NewCookieAuthorizer returns an instance of CookieAuthorizer
func NewCookieAuthorizer(name, value string) *CookieAuthorizer {
	return &CookieAuthorizer{
		name:  name,
		value: value,
	}
}

// Authorize the request with the cookie
func (c *CookieAuthorizer) Authorize(req *http.Request) error {
	if req == nil {
		return errors.New("the request is null")
	}

	req.AddCookie(&http.Cookie{
		Name:  c.name,
		Value: c.value,
	})
	return nil
}

// SecretAuthorizer authorizes the requests with the specified secret
type SecretAuthorizer struct {
	authorizer *CookieAuthorizer
}

// NewSecretAuthorizer returns an instance of SecretAuthorizer
func NewSecretAuthorizer(secret string) *SecretAuthorizer {
	return &SecretAuthorizer{
		authorizer: NewCookieAuthorizer(secretCookieName, secret),
	}
}

// Authorize the request with the secret
func (s *SecretAuthorizer) Authorize(req *http.Request) error {
	return s.authorizer.Authorize(req)
}
