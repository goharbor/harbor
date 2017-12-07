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

package client

import (
	"net/http"

	"github.com/vmware/harbor/src/common/http/client/auth"
)

// Client defines the method that a HTTP client should implement
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// AuthorizedClient authorizes the requests before sending them
type AuthorizedClient struct {
	client     *http.Client
	authorizer auth.Authorizer
}

// NewAuthorizedClient returns an instance of the AuthorizedClient
func NewAuthorizedClient(authorizer auth.Authorizer, client ...*http.Client) *AuthorizedClient {
	c := &AuthorizedClient{
		authorizer: authorizer,
	}
	if len(client) > 0 {
		c.client = client[0]
	}
	if c.client == nil {
		c.client = &http.Client{}
	}
	return c
}

// Do authorizes the request before sending it
func (a *AuthorizedClient) Do(req *http.Request) (*http.Response, error) {
	if a.authorizer != nil {
		if err := a.authorizer.Authorize(req); err != nil {
			return nil, err
		}
	}
	return a.client.Do(req)
}
