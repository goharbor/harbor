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

package utils

import (
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"net/http"
)

//NewRepositoryClient create a repository client with scope type "reopsitory" and scope as the repository it would access.
func NewRepositoryClient(endpoint string, insecure bool, credential auth.Credential,
	tokenServiceEndpoint, repository string, actions ...string) (*registry.Repository, error) {
	authorizer := auth.NewStandardTokenAuthorizer(credential, insecure,
		tokenServiceEndpoint, "repository", repository, actions...)

	store, err := auth.NewAuthorizerStore(endpoint, insecure, authorizer)
	if err != nil {
		return nil, err
	}

	uam := &userAgentModifier{
		userAgent: "harbor-registry-client",
	}

	client, err := registry.NewRepositoryWithModifiers(repository, endpoint, insecure, store, uam)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type userAgentModifier struct {
	userAgent string
}

// Modify adds user-agent header to the request
func (u *userAgentModifier) Modify(req *http.Request) error {
	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), u.userAgent)
	return nil
}
