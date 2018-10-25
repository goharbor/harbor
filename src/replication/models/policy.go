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

package models

import (
	"time"
)

// ReplicationPolicy defines the structure of a replication policy.
type ReplicationPolicy struct {
	ID                int64 // UUID of the policy
	Name              string
	Description       string
	Filters           []Filter
	ReplicateDeletion bool
	Trigger           *Trigger // The trigger of the replication
	ProjectIDs        []int64  // Projects attached to this policy
	TargetIDs         []int64
	Namespaces        []string // The namespaces are used to set immediate trigger
	CreationTime      time.Time
	UpdateTime        time.Time
}

// QueryParameter defines the parameters used to do query selection.
type QueryParameter struct {
	// Query by page, couple with pageSize
	Page int64

	// Size of each page, couple with page
	PageSize int64

	// Query by project ID
	ProjectID int64

	// Query by name
	Name string
}

// ReplicationPolicyQueryResult is the query result of replication policy
type ReplicationPolicyQueryResult struct {
	Total    int64
	Policies []*ReplicationPolicy
}
