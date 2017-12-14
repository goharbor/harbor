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
	"io/ioutil"
	"net/http"

	commonhttp "github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/http/modifier/auth"
	"github.com/vmware/harbor/src/jobservice/api"
)

// Client defines the methods that a jobservice client should implement
type Client interface {
	SubmitReplicationJob(*api.ReplicationReq) error
}

// DefaultClient provides a default implement for the interface Client
type DefaultClient struct {
	endpoint string
	client   *commonhttp.Client
}

// Config contains configuration items needed for DefaultClient
type Config struct {
	Secret string
}

// NewDefaultClient returns an instance of DefaultClient
func NewDefaultClient(endpoint string, cfg *Config) *DefaultClient {
	c := &DefaultClient{
		endpoint: endpoint,
	}

	if cfg != nil {
		c.client = commonhttp.NewClient(nil, auth.NewSecretAuthorizer(cfg.Secret))
	}

	return c
}

// SubmitReplicationJob submits a replication job to the jobservice
func (d *DefaultClient) SubmitReplicationJob(replication *api.ReplicationReq) error {
	url := d.endpoint + "/api/jobs/replication"

	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(replication); err != nil {
		return err
	}

	resp, err := d.client.Post(url, "application/json", buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(message),
		}
	}

	return nil
}
