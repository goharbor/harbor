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

package retention

import (
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

// Manager defines operations of managing policy
type Manager interface {
	// Create new policy and return uuid
	CreatePolicy(p *policy.Metadata) (string, error)
	// Update the existing policy
	// Full update
	UpdatePolicy(p *policy.Metadata) error
	// Delete the specified policy
	// No actual use so far
	DeletePolicy(ID string) error
	// Get the specified policy
	GetPolicy(ID string) (*policy.Metadata, error)
	// Create a new retention execution
	CreateExecution(execution *Execution) (string, error)
	// Update the specified execution
	UpdateExecution(execution *Execution) error
	// Get the specified execution
	GetExecution(eid string) (*Execution, error)
	// List execution histories
	ListExecutions(query *q.Query) ([]*Execution, error)
	// Add new history
	AppendHistory(history *History) error
	// List all the histories marked by the specified execution
	ListHistories(executionID string, query *q.Query) ([]*History, error)
}
