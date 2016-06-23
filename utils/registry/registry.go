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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	registry_error "github.com/vmware/harbor/utils/registry/error"
	"github.com/vmware/harbor/utils/registry/utils"
)

// Registry holds information of a registry entity
type Registry struct {
	Endpoint *url.URL
	client   *http.Client
}

// NewRegistry returns an instance of registry
func NewRegistry(endpoint string, client *http.Client) (*Registry, error) {
	u, err := utils.ParseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	registry := &Registry{
		Endpoint: u,
		client:   client,
	}

	return registry, nil
}

// NewRegistryWithModifiers returns an instance of Registry according to the modifiers
func NewRegistryWithModifiers(endpoint string, insecure bool, modifiers ...Modifier) (*Registry, error) {
	u, err := utils.ParseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	transport := NewTransport(t, modifiers...)

	return &Registry{
		Endpoint: u,
		client: &http.Client{
			Transport: transport,
		},
	}, nil
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
		return repos, parseError(err)
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
		return parseError(err)
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
