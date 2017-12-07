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

package replicator

import (
	"github.com/vmware/harbor/src/common/http/client"
	"github.com/vmware/harbor/src/jobservice/api"
	jobserviceclient "github.com/vmware/harbor/src/jobservice/client"
)

// Replicator submits the replication work to the jobservice
type Replicator interface {
	Replicate(*api.ReplicationReq) error
}

// DefaultReplicator provides a default implement for Replicator
type DefaultReplicator struct {
	client jobserviceclient.Client
}

// NewDefaultReplicator returns an instance of DefaultReplicator
func NewDefaultReplicator(endpoint string, client ...client.Client) *DefaultReplicator {
	return &DefaultReplicator{
		client: jobserviceclient.NewDefaultClient(endpoint, client...),
	}
}

// Replicate ...
func (d *DefaultReplicator) Replicate(replication *api.ReplicationReq) error {
	return d.client.SubmitReplicationJob(replication)
}
