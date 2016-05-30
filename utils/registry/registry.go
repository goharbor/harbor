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
	registry_error "github.com/vmware/harbor/utils/registry/error"
)

const (
	// UserAgent is used to decorate the request so it can be identified by webhook.
	UserAgent string = "registry-client"
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

	client, err := newClient(endpoint, username, nil, "registry", "catalog", "*")
	if err != nil {
		return nil, err
	}

	registry := &Registry{
		Endpoint: u,
		client:   client,
	}

	log.Debugf("initialized a registry client with username: %s %s", endpoint, username)

	return registry, nil
}

// NewRegistryWithCredential returns a Registry instance which associate to a crendential.
// And Credential is essentially a decorator for client to docorate the request before sending it to the registry.
func NewRegistryWithCredential(endpoint string, credential auth.Credential) (*Registry, error) {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimRight(endpoint, "/")
	if !strings.HasPrefix(endpoint, "http://") &&
		!strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	client, err := newClient(endpoint, "", credential, "", "", "")
	if err != nil {
		return nil, err
	}

	registry := &Registry{
		Endpoint: u,
		client:   client,
	}

	log.Debugf("initialized a registry client with credential: %s", endpoint)

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
		if regErr, ok := err.(*registry_error.Error); ok {
			return repos, regErr
		}

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

	return repos, &registry_error.Error{
		StatusCode: resp.StatusCode,
		Detail:     string(b),
	}
}

// Ping ...
func (r *Registry) Ping() error {
	req, err := http.NewRequest("GET", buildPingURL(r.Endpoint.String()), nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if regErr, ok := urlErr.Err.(*registry_error.Error); ok {
				return &registry_error.Error{
					StatusCode: regErr.StatusCode,
					Detail:     regErr.Detail,
				}
			}
			return urlErr.Err
		}

		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return &registry_error.Error{
		StatusCode: resp.StatusCode,
		Detail:     string(b),
	}
}

func buildCatalogURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/_catalog", endpoint)
}

func newClient(endpoint, username string, credential auth.Credential,
	scopeType, scopeName string, scopeActions ...string) (*http.Client, error) {

	endpoint = strings.TrimRight(endpoint, "/")
	resp, err := http.Get(buildPingURL(endpoint))
	if err != nil {
		return nil, err
	}

	var handlers []auth.Handler
	var handler auth.Handler
	if credential != nil {
		handler = auth.NewStandardTokenHandler(credential, scopeType, scopeName, scopeActions...)
	} else {
		handler = auth.NewUsernameTokenHandler(username, scopeType, scopeName, scopeActions...)
	}

	handlers = append(handlers, handler)

	challenges := auth.ParseChallengeFromResponse(resp)
	authorizer := auth.NewRequestAuthorizer(handlers, challenges)
	headerModifier := NewHeaderModifier(map[string]string{http.CanonicalHeaderKey("User-Agent"): UserAgent})

	transport := NewTransport(http.DefaultTransport, []RequestModifier{authorizer, headerModifier})
	return &http.Client{
		Transport: transport,
	}, nil
}
