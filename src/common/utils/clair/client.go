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

package clair

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//	"path"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// Client communicates with clair endpoint to scan image and get detailed scan result
type Client struct {
	endpoint string
	//need to customize the logger to write output to job log.
	logger *log.Logger
	client *http.Client
}

// NewClient creates a new instance of client, set the logger as the job's logger if it's used in a job handler.
func NewClient(endpoint string, logger *log.Logger) *Client {
	if logger == nil {
		logger = log.DefaultLogger()
	}
	return &Client{
		endpoint: endpoint,
		logger:   logger,
		client:   &http.Client{},
	}
}

// ScanLayer calls Clair's API to scan a layer.
func (c *Client) ScanLayer(l models.ClairLayer) error {
	layer := models.ClairLayerEnvelope{
		Layer: &l,
		Error: nil,
	}
	data, err := json.Marshal(layer)
	if err != nil {
		return err
	}
	c.logger.Infof("endpoint: %s", c.endpoint)
	c.logger.Infof("body: %s", string(data))
	req, err := http.NewRequest("POST", c.endpoint+"/v1/layers", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set(http.CanonicalHeaderKey("Content-Type"), "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	c.logger.Infof("response code: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusCreated {
		c.logger.Warningf("Unexpected status code: %d", resp.StatusCode)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("Unexpected status code: %d, text: %s", resp.StatusCode, string(b))
	}
	c.logger.Infof("Returning.")
	return nil
}

// GetResult calls Clair's API to get layers with detailed vulnerability list
func (c *Client) GetResult(layerName string) (*models.ClairLayerEnvelope, error) {
	req, err := http.NewRequest("GET", c.endpoint+"/v1/layers/"+layerName+"?features&vulnerabilities", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d, text: %s", resp.StatusCode, string(b))
	}
	var res models.ClairLayerEnvelope
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
