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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/jobservice/config"
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

// BuildBlobURL ...
func BuildBlobURL(endpoint, repository, digest string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/%s", endpoint, repository, digest)
}

//GetTokenForRepo is a temp solution for job handler to get a token for clair.
func GetTokenForRepo(repository string) (string, error) {
	u, err := url.Parse(config.InternalTokenServiceEndpoint())
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Add("service", "harbor-registry")
	q.Add("scope", fmt.Sprintf("repository:%s:pull", repository))
	u.RawQuery = q.Encode()
	r, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	c := &http.Cookie{Name: models.UISecretCookie, Value: config.JobserviceSecret()}
	r.AddCookie(c)
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unexpected response from token service, code: %d, %s", resp.StatusCode, string(b))
	}
	tk := models.Token{}
	if err := json.Unmarshal(b, &tk); err != nil {
		return "", err
	}
	return tk.Token, nil
}
