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
	"github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/http/modifier/auth"
)

// Replication holds information for submiting a replication job
type Replication struct {
	PolicyID   int64    `json:"policy_id"`
	Repository string   `json:"repository"`
	Operation  string   `json:"operation"`
	Tags       []string `json:"tags"`
}

// Client defines the methods that a jobservice client should implement
type Client interface {
	SubmitReplicationJob(*Replication) error
	StopReplicationJobs(policyID int64) error
}

// DefaultClient provides a default implement for the interface Client
type DefaultClient struct {
	endpoint string
	client   *http.Client
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
		c.client = http.NewClient(nil, auth.NewSecretAuthorizer(cfg.Secret))
	}

	return c
}

// SubmitReplicationJob submits a replication job to the jobservice
func (d *DefaultClient) SubmitReplicationJob(replication *Replication) error {
	url := d.endpoint + "/api/jobs/replication"
	return d.client.Post(url, replication)
}

// StopReplicationJobs stop replication jobs of the policy specified by the policy ID
func (d *DefaultClient) StopReplicationJobs(policyID int64) error {
	url := d.endpoint + "/api/jobs/replication/actions"
	return d.client.Post(url, &struct {
		PolicyID int64  `json:"policy_id"`
		Action   string `json:"action"`
	}{
		PolicyID: policyID,
		Action:   "stop",
	})
}
