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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/registry"
)

// Authorizer authorizes requests according to the schema
type Authorizer interface {
	// Scheme : basic, bearer
	Scheme() string
	//Authorize adds basic auth or token auth to the header of request
	Authorize(req *http.Request, params map[string]string) error
}

// AuthorizerStore holds a authorizer list, which will authorize request.
// And it implements interface Modifier
type AuthorizerStore struct {
	authorizers []Authorizer
	ping        *url.URL
	challenges  []challenge.Challenge
}

// NewAuthorizerStore ...
func NewAuthorizerStore(endpoint string, insecure bool, authorizers ...Authorizer) (*AuthorizerStore, error) {
	endpoint = utils.FormatEndpoint(endpoint)

	client := &http.Client{
		Transport: registry.GetHTTPTransport(insecure),
		Timeout:   30 * time.Second,
	}

	pingURL := buildPingURL(endpoint)
	resp, err := client.Get(pingURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	challenges := ParseChallengeFromResponse(resp)
	ping, err := url.Parse(pingURL)
	if err != nil {
		return nil, err
	}
	return &AuthorizerStore{
		authorizers: authorizers,
		ping:        ping,
		challenges:  challenges,
	}, nil
}

func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}

// Modify adds authorization to the request
func (a *AuthorizerStore) Modify(req *http.Request) error {
	//only handle the requests sent to registry
	v2Index := strings.Index(req.URL.Path, "/v2/")
	if v2Index == -1 {
		return nil
	}

	ping := url.URL{
		Host:   req.URL.Host,
		Scheme: req.URL.Scheme,
		Path:   req.URL.Path[:v2Index+4],
	}

	if ping.Host != a.ping.Host || ping.Scheme != a.ping.Scheme ||
		ping.Path != a.ping.Path {
		return nil
	}

	for _, challenge := range a.challenges {
		for _, authorizer := range a.authorizers {
			if authorizer.Scheme() == challenge.Scheme {
				if err := authorizer.Authorize(req, challenge.Parameters); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
