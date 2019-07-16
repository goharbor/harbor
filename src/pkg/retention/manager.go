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
	"encoding/json"
	"github.com/goharbor/harbor/src/pkg/retention/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	"time"
)

// Manager defines operations of managing policy
type Manager interface {
	// Create new policy and return uuid
	CreatePolicy(p *policy.Metadata) (int64, error)
	// Update the existing policy
	// Full update
	UpdatePolicy(p *policy.Metadata) error
	// Delete the specified policy
	// No actual use so far
	DeletePolicy(ID int64) error
	// Get the specified policy
	GetPolicy(ID int64) (*policy.Metadata, error)
	// Create a new retention execution
	CreateExecution(execution *Execution) (int64, error)
	// Update the specified execution
	UpdateExecution(execution *Execution) error
	// Get the specified execution
	GetExecution(eid int64) (*Execution, error)
	// List execution histories
	ListExecutions(query *q.Query) ([]*Execution, error)
	// Add new history
	AppendHistory(history *History) error
	// List all the histories marked by the specified execution
	ListHistories(executionID int64, query *q.Query) ([]*History, error)
}

// DefaultManager ...
type DefaultManager struct {
}

// CreatePolicy Create Policy
func (d *DefaultManager) CreatePolicy(p *policy.Metadata) (int64, error) {
	var p1 *models.RetentionPolicy
	p1.ScopeLevel = p.Scope.Level
	p1.TriggerKind = p.Trigger.Kind
	data, _ := json.Marshal(p)
	p1.Data = string(data)
	p1.CreateTime = time.Now()
	p1.UpdateTime = p1.CreateTime
	return dao.CreatePolicy(p1)
}

// UpdatePolicy Update Policy
func (d *DefaultManager) UpdatePolicy(p *policy.Metadata) error {
	var p1 *models.RetentionPolicy
	p1.ID = p.ID
	p1.ScopeLevel = p.Scope.Level
	p1.TriggerKind = p.Trigger.Kind
	p.ID = 0
	data, _ := json.Marshal(p)
	p.ID = p1.ID
	p1.Data = string(data)
	p1.UpdateTime = time.Now()
	return dao.UpdatePolicy(p1)
}

// DeletePolicy Delete Policy
func (d *DefaultManager) DeletePolicy(id int64) error {
	return dao.DeletePolicy(id)
}

// GetPolicy Get Policy
func (d *DefaultManager) GetPolicy(id int64) (*policy.Metadata, error) {
	p1, err := dao.GetPolicy(id)
	if err != nil {
		return nil, err
	}
	var p *policy.Metadata
	if err = json.Unmarshal([]byte(p1.Data), p); err != nil {
		return nil, err
	}
	return p, nil
}

// CreateExecution Create Execution
func (d *DefaultManager) CreateExecution(execution *Execution) (int64, error) {
	var exec *models.RetentionExecution
	exec.PolicyID = execution.PolicyID
	exec.StartTime = time.Now()
	exec.Status = "Running"
	return dao.CreateExecution(exec)
}

// UpdateExecution Update Execution
func (d *DefaultManager) UpdateExecution(execution *Execution) error {
	var exec *models.RetentionExecution
	exec.ID = execution.ID
	exec.PolicyID = execution.PolicyID
	exec.StartTime = time.Now()
	exec.Status = "Running"
	return dao.UpdateExecution(exec)
}

// ListExecutions List Executions
func (d *DefaultManager) ListExecutions(query *q.Query) ([]*Execution, error) {
	execs, err := dao.ListExecutions(query)
	if err != nil {
		return nil, err
	}
	var execs1 []*Execution
	for _, e := range execs {
		var e1 *Execution
		e1.ID = e.ID
		e1.PolicyID = e.PolicyID
		e1.Status = e.Status
		e1.StartTime = e.StartTime
		e1.EndTime = e.EndTime
		execs1 = append(execs1, e1)
	}
	return execs1, nil
}

// GetExecution Get Execution
func (d *DefaultManager) GetExecution(eid int64) (*Execution, error) {
	e, err := dao.GetExecution(eid)
	if err != nil {
		return nil, err
	}
	var e1 *Execution
	e1.ID = e.ID
	e1.PolicyID = e.PolicyID
	e1.Status = e.Status
	e1.StartTime = e.StartTime
	e1.EndTime = e.EndTime
	return e1, nil
}

// ListHistories List Histories
func (d *DefaultManager) ListHistories(executionID int64, query *q.Query) ([]*History, error) {
	his, err := dao.ListExecHistories(executionID, query)
	if err != nil {
		return nil, err
	}
	var his1 []*History
	for _, h := range his {
		var h1 *History
		h1.ExecutionID = h.ExecutionID
		h1.Artifact = h.Artifact
		h1.Rule.ID = h.RuleID
		h1.Rule.DisplayText = h.RuleDisplayText
		h1.Timestamp = h.Timestamp
		his1 = append(his1, h1)
	}
	return his1, nil
}

// AppendHistory Append History
func (d *DefaultManager) AppendHistory(h *History) error {
	var h1 *models.RetentionTask
	h1.ExecutionID = h.ExecutionID
	h1.Artifact = h.Artifact
	h1.RuleID = h.Rule.ID
	h1.RuleDisplayText = h.Rule.DisplayText
	h1.Timestamp = h.Timestamp
	_, err := dao.AppendExecHistory(h1)
	return err
}

// NewManager ...
func NewManager() Manager {
	return &DefaultManager{}
}
