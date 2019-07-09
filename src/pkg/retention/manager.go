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

type DefaultManager struct {
}

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

func (d *DefaultManager) DeletePolicy(id int64) error {
	return dao.DeletePolicy(id)
}

func (d *DefaultManager) GetPolicy(id int64) (*policy.Metadata, error) {
	if p1,err:=dao.GetPolicy(id);err!=nil{
		return nil,err
	}else{
		var p *policy.Metadata
		if err=json.Unmarshal([]byte(p1.Data), p);err!=nil{
			return nil,err
		}else{
			return p,nil
		}
	}
}

func (d *DefaultManager) CreateExecution(execution *Execution) (int64, error) {
	var exec *models.RetentionExecution
	exec.PolicyID=execution.PolicyID
	exec.StartTime=time.Now()
	exec.Status="Running"
	return dao.CreateExecution(exec)
}

func (d *DefaultManager) UpdateExecution(execution *Execution) error {
	var exec *models.RetentionExecution
	exec.ID = execution.ID
	exec.PolicyID=execution.PolicyID
	exec.StartTime=time.Now()
	exec.Status="Running"
	return dao.UpdateExecution(exec)
}

func (d *DefaultManager) ListExecutions(query *q.Query) ([]*Execution, error) {
	return []*Execution{
		{
			ID:        1,
			PolicyID:  1,
			StartTime: time.Now().Add(-time.Minute),
			EndTime:   time.Now(),
			Status:    "Success",
		},
		{
			ID:        2,
			PolicyID:  1,
			StartTime: time.Now().Add(-time.Minute),
			EndTime:   time.Now(),
			Status:    "Failed",
		},
		{
			ID:        3,
			PolicyID:  1,
			StartTime: time.Now().Add(-time.Minute),
			EndTime:   time.Now(),
			Status:    "Running",
		},
	}, nil
}

func (d *DefaultManager) GetExecution(eid int64) (*Execution, error) {
	return &Execution{
		ID:        1,
		PolicyID:  1,
		StartTime: time.Now().Add(-time.Minute),
		EndTime:   time.Now(),
		Status:    "Success",
	}, nil
}

func (d *DefaultManager) ListHistories(executionID int64, query *q.Query) ([]*History, error) {
	panic("implement me")
}

func (d *DefaultManager) AppendHistory(history *History) error {
	panic("implement me")
}

func NewManager() Manager {
	return &DefaultManager{}
}
