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

package azurecr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/docker/distribution/registry/client/auth/challenge"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/registry/auth"
)

var (
	// scopeActionMetadataRead represents metadata_read action
	scopeActionMetadataRead = "metadata_read"
	// scopeTypeRepository represents repository resource
	scopeTypeRepository = "repository"
)

var _ lib.Authorizer = &authorizer{}

type token struct {
	AccessToken string `json:"access_token"`
}

// authorizer is a customize authorizer for azurecr adapter which
// inherits lib authorizer.
type authorizer struct {
	registry        *model.Registry
	innerAuthorizer lib.Authorizer
	client          *http.Client
}

func newAuthorizer(registry *model.Registry) *authorizer {
	var username, password string
	if registry.Credential != nil {
		username = registry.Credential.AccessKey
		password = registry.Credential.AccessSecret
	}

	return &authorizer{
		registry:        registry,
		innerAuthorizer: auth.NewAuthorizer(username, password, registry.Insecure),
		client: &http.Client{Transport: commonhttp.GetHTTPTransport(
			commonhttp.WithInsecure(registry.Insecure),
			commonhttp.WithCACert(registry.CACertificate),
		)},
	}
}

func (a *authorizer) Modify(req *http.Request) error {
	if !isTagList(req.URL) {
		// pass through non tag list api
		return a.innerAuthorizer.Modify(req)
	}

	// tag list api should fetch token
	url, err := a.buildTokenAPI(req.URL)
	if err != nil {
		return err
	}

	tokenReq, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil
	}

	if a.registry.Credential != nil {
		tokenReq.SetBasicAuth(a.registry.Credential.AccessKey, a.registry.Credential.AccessSecret)
	}

	resp, err := a.client.Do(tokenReq)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tk token
	if err = json.Unmarshal(body, &tk); err != nil {
		return err
	}

	if tk.AccessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tk.AccessToken))
	}

	return nil
}

// buildTokenAPI builds token request API path.
func (a *authorizer) buildTokenAPI(u *url.URL) (*url.URL, error) {
	v2URL, err := url.Parse(u.Scheme + "://" + u.Host + "/v2/")
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Get(v2URL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	challenges := challenge.ResponseChallenges(resp)
	if len(challenges) == 0 {
		return nil, errors.New("invalid response challenges")
	}
	cm := map[string]challenge.Challenge{}
	for _, challenge := range challenges {
		cm[challenge.Scheme] = challenge
	}

	challenge, exist := cm["bearer"]
	if !exist {
		return nil, errors.New("no bearer challenge found")
	}

	tokenURL, err := url.Parse(challenge.Parameters["realm"])
	if err != nil {
		return nil, err
	}

	query := tokenURL.Query()
	query.Add("service", challenge.Parameters["service"])

	var repository string
	if subs := lib.V2TagListURLRe.FindStringSubmatch(u.Path); len(subs) >= 2 {
		// tag
		repository = subs[1]
	}

	if repository == "" {
		return nil, errors.Errorf("invalid repository name, url: %s", u.String())
	}

	query.Add("scope", fmt.Sprintf("%s:%s:%s", scopeTypeRepository, repository, scopeActionMetadataRead))
	tokenURL.RawQuery = query.Encode()
	return tokenURL, nil
}

// isTagList checks the request whether tag list API.
func isTagList(u *url.URL) bool {
	return lib.V2TagListURLRe.Match([]byte(u.Path))
}
