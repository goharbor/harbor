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

package jfrog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

// client is a client to interact with Jfrog
type client struct {
	// client is a client to access jfrog
	client   *common_http.Client
	url      string
	insecure bool
	username string
	password string
}

// newClient constructs a jfrog client
func newClient(reg *model.Registry) *client {
	username, password := "", ""
	if reg.Credential != nil {
		username = reg.Credential.AccessKey
		password = reg.Credential.AccessSecret
	}

	return &client{
		client: common_http.NewClient(
			&http.Client{
				Transport: common_http.GetHTTPTransport(common_http.WithInsecure(reg.Insecure)),
			},
			basic.NewAuthorizer(username, password),
		),
		url:      reg.URL,
		insecure: reg.Insecure,
		username: username,
		password: password,
	}
}

// getDockerRepositories gets docker repositories from jfrog
func (c *client) getDockerRepositories() ([]*repository, error) {
	var repositories []*repository
	url := fmt.Sprintf("%s/artifactory/api/repositories?packageType=docker", c.url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return repositories, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return repositories, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return repositories, err
	}

	err = json.Unmarshal(body, &repositories)
	return repositories, err
}

// createDockerRepository creates docker repository on jfrog
func (c *client) createDockerRepository(name string) error {
	ns := newDefaultDockerLocalRepository(name)
	body, err := json.Marshal(ns)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/artifactory/api/repositories/%s", c.url, name)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}
