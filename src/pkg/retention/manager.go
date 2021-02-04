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
	"fmt"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/pkg/retention/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
)

// Manager defines operations of managing policy
type Manager interface {
	// Create new policy and return ID
	CreatePolicy(p *policy.Metadata) (int64, error)
	// Update the existing policy
	// Full update
	UpdatePolicy(p *policy.Metadata) error
	// Delete the specified policy
	// No actual use so far
	DeletePolicy(ID int64) error
	// Get the specified policy
	GetPolicy(ID int64) (*policy.Metadata, error)
}

// DefaultManager ...
type DefaultManager struct {
}

// CreatePolicy Create Policy
func (d *DefaultManager) CreatePolicy(p *policy.Metadata) (int64, error) {
	p1 := &models.RetentionPolicy{}
	p1.ScopeLevel = p.Scope.Level
	p1.ScopeReference = p.Scope.Reference
	p1.TriggerKind = p.Trigger.Kind
	data, _ := json.Marshal(p)
	p1.Data = string(data)
	p1.CreateTime = time.Now()
	p1.UpdateTime = p1.CreateTime
	return dao.CreatePolicy(p1)
}

// UpdatePolicy Update Policy
func (d *DefaultManager) UpdatePolicy(p *policy.Metadata) error {
	p1 := &models.RetentionPolicy{}
	p1.ID = p.ID
	p1.ScopeLevel = p.Scope.Level
	p1.ScopeReference = p.Scope.Reference
	p1.TriggerKind = p.Trigger.Kind
	p.ID = 0
	data, _ := json.Marshal(p)
	p.ID = p1.ID
	p1.Data = string(data)
	p1.UpdateTime = time.Now()
	return dao.UpdatePolicy(p1, "scope_level", "trigger_kind", "data", "update_time")
}

// DeletePolicy Delete Policy
func (d *DefaultManager) DeletePolicy(id int64) error {
	return dao.DeletePolicy(id)
}

// GetPolicy Get Policy
func (d *DefaultManager) GetPolicy(id int64) (*policy.Metadata, error) {
	p1, err := dao.GetPolicy(id)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("no such Retention policy with id %v", id)
		}
		return nil, err
	}
	p := &policy.Metadata{}
	if err = json.Unmarshal([]byte(p1.Data), p); err != nil {
		return nil, err
	}
	p.ID = id
	return p, nil
}

// NewManager ...
func NewManager() Manager {
	return &DefaultManager{}
}
