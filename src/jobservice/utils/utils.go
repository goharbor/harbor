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
	"fmt"
	"net/http"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/jobservice/config"
)

//NewRepositoryClient create a repository client with scope type "reopsitory" and scope as the repository it would access.
func NewRepositoryClient(endpoint string, insecure bool, credential auth.Credential,
	tokenServiceEndpoint, repository string) (*registry.Repository, error) {
	authorizer := auth.NewStandardTokenAuthorizer(credential, insecure,
		tokenServiceEndpoint)

	uam := &userAgentModifier{
		userAgent: "harbor-registry-client",
	}

	return registry.NewRepositoryWithModifiers(repository, endpoint, insecure, authorizer, uam)
}

type userAgentModifier struct {
	userAgent string
}

// Modify adds user-agent header to the request
func (u *userAgentModifier) Modify(req *http.Request) error {
	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), u.userAgent)
	return nil
}

// BuildBlobURL ...
func BuildBlobURL(endpoint, repository, digest string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/%s", endpoint, repository, digest)
}

//GetTokenForRepo is used for job handler to get a token for clair.
func GetTokenForRepo(repository string) (string, error) {
	c := &http.Cookie{Name: models.UISecretCookie, Value: config.JobserviceSecret()}
	credentail := auth.NewCookieCredential(c)
	token, err := auth.GetToken(config.InternalTokenServiceEndpoint(), true, credentail, []*auth.Scope{&auth.Scope{
		Type:    "repository",
		Name:    repository,
		Actions: []string{"pull"},
	}})
	if err != nil {
		return "", err
	}

	return token.Token, nil
}
