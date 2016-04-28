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

package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry/auth"
	"github.com/vmware/harbor/utils/registry/errors"
)

// Registry holds information of a registry entity
type Registry struct {
	Endpoint *url.URL
	client   *http.Client
}

// NewRegistry returns an instance of registry
func NewRegistry(endpoint string, client *http.Client) (*Registry, error) {
	endpoint = strings.TrimRight(endpoint, "/")

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	registry := &Registry{
		Endpoint: u,
		client:   client,
	}

	log.Debugf("initialized a registry client: %s", endpoint)

	return registry, nil
}

// NewRegistryWithUsername returns a Registry instance which will authorize the request
// according to the privileges of user
func NewRegistryWithUsername(endpoint, username string) (*Registry, error) {
	endpoint = strings.TrimRight(endpoint, "/")

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(buildPingURL(endpoint))
	if err != nil {
		return nil, err
	}

	var handlers []auth.Handler
	handler := auth.NewUsernameTokenHandler(username, "registry", "catalog", "*")
	handlers = append(handlers, handler)

	challenges := auth.ParseChallengeFromResponse(resp)
	authorizer := auth.NewRequestAuthorizer(handlers, challenges)

	transport := NewTransport(http.DefaultTransport, []RequestModifier{authorizer})

	registry := &Registry{
		Endpoint: u,
		client: &http.Client{
			Transport: transport,
		},
	}

	return registry, nil
}

// Catalog ...
func (r *Registry) Catalog() ([]string, error) {
	repos := []string{}

	req, err := http.NewRequest("GET", buildCatalogURL(r.Endpoint.String()), nil)
	if err != nil {
		return repos, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return repos, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return repos, err
	}

	if resp.StatusCode == http.StatusOK {
		catalogResp := struct {
			Repositories []string `json:"repositories"`
		}{}

		if err := json.Unmarshal(b, &catalogResp); err != nil {
			return repos, err
		}

		repos = catalogResp.Repositories

		return repos, nil
	}

	return repos, errors.Error{
		StatusCode: resp.StatusCode,
		Message:    string(b),
	}
}

func buildCatalogURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/_catalog", endpoint)
}
