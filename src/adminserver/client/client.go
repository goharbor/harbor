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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vmware/harbor/src/adminserver/client/auth"
	"github.com/vmware/harbor/src/adminserver/systeminfo/imagestorage"
	"github.com/vmware/harbor/src/common/utils"
)

// Client defines methods that an Adminserver client should implement
type Client interface {
	// Ping tests the connection with server
	Ping() error
	// GetCfgs returns system configurations
	GetCfgs() (map[string]interface{}, error)
	// UpdateCfgs updates system configurations
	UpdateCfgs(map[string]interface{}) error
	// ResetCfgs resets system configuratoins form environment variables
	ResetCfgs() error
	// Capacity returns the capacity of image storage
	Capacity() (*imagestorage.Capacity, error)
}

// NewClient return an instance of Adminserver client
func NewClient(baseURL string, authorizer auth.Authorizer) Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}
	return &client{
		baseURL:    baseURL,
		client:     &http.Client{},
		authorizer: authorizer,
	}
}

type client struct {
	baseURL    string
	client     *http.Client
	authorizer auth.Authorizer
}

// do creates request and authorizes it if authorizer is not nil
func (c *client) do(method, relativePath string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + relativePath
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if c.authorizer != nil {
		if err := c.authorizer.Authorize(req); err != nil {
			return nil, err
		}
	}
	return c.client.Do(req)
}

func (c *client) Ping() error {
	addr := strings.Split(c.baseURL, "://")[1]
	if !strings.Contains(addr, ":") {
		addr = addr + ":80"
	}

	return utils.TestTCPConn(addr, 60, 2)
}

// GetCfgs ...
func (c *client) GetCfgs() (map[string]interface{}, error) {
	resp, err := c.do(http.MethodGet, "/api/configurations", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get configurations: %d %s",
			resp.StatusCode, b)
	}

	cfgs := map[string]interface{}{}
	if err = json.Unmarshal(b, &cfgs); err != nil {
		return nil, err
	}

	return cfgs, nil
}

// UpdateCfgs ...
func (c *client) UpdateCfgs(cfgs map[string]interface{}) error {
	data, err := json.Marshal(cfgs)
	if err != nil {
		return err
	}

	resp, err := c.do(http.MethodPut, "/api/configurations", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to update configurations: %d %s",
			resp.StatusCode, b)
	}

	return nil
}

// ResetCfgs ...
func (c *client) ResetCfgs() error {
	resp, err := c.do(http.MethodPost, "/api/configurations/reset", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to reset configurations: %d %s",
			resp.StatusCode, b)
	}

	return nil
}

// Capacity ...
func (c *client) Capacity() (*imagestorage.Capacity, error) {
	resp, err := c.do(http.MethodGet, "/api/systeminfo/capacity", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get capacity: %d %s",
			resp.StatusCode, b)
	}

	capacity := &imagestorage.Capacity{}
	if err = json.Unmarshal(b, capacity); err != nil {
		return nil, err
	}

	return capacity, nil
}
