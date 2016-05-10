/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package auth

import (
	"net/http"

	au "github.com/docker/distribution/registry/client/auth"
)

// Handler authorizes requests according to the schema
type Handler interface {
	// Scheme : basic, bearer
	Scheme() string
	//AuthorizeRequest adds basic auth or token auth to the header of request
	AuthorizeRequest(req *http.Request, params map[string]string) error
}

// RequestAuthorizer holds a handler list, which will authorize request.
// Implements interface RequestModifier
type RequestAuthorizer struct {
	handlers   []Handler
	challenges []au.Challenge
}

// NewRequestAuthorizer ...
func NewRequestAuthorizer(handlers []Handler, challenges []au.Challenge) *RequestAuthorizer {
	return &RequestAuthorizer{
		handlers:   handlers,
		challenges: challenges,
	}
}

// ModifyRequest adds authorization to the request
func (r *RequestAuthorizer) ModifyRequest(req *http.Request) error {
	for _, handler := range r.handlers {
		for _, challenge := range r.challenges {
			if handler.Scheme() == challenge.Scheme {
				if err := handler.AuthorizeRequest(req, challenge.Parameters); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
