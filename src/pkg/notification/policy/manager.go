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

package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/policy/dao"
	"github.com/goharbor/harbor/src/pkg/notification/policy/model"
)

var (
	// Mgr is a global variable for the default notification policies
	Mgr = NewManager()
)

// Manager manages the notification policies
type Manager interface {
	// Create new policy
	Create(ctx context.Context, policy *model.Policy) (int64, error)
	// List the policies, returns the policy list and error
	List(ctx context.Context, query *q.Query) ([]*model.Policy, error)
	// Count the policies, returns the policy count and error
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Get policy with specified ID
	Get(ctx context.Context, id int64) (*model.Policy, error)
	// Update the specified policy
	Update(ctx context.Context, policy *model.Policy) error
	// Delete the specified policy
	Delete(ctx context.Context, policyID int64) error
	// GetRelatedPolices get event type related policies in project
	GetRelatedPolices(ctx context.Context, projectID int64, eventType string) ([]*model.Policy, error)
}

var _ Manager = &manager{}

type manager struct {
	dao dao.DAO
}

// NewManager ...
func NewManager() Manager {
	return &manager{
		dao: dao.New(),
	}
}

// Create notification policy
func (m *manager) Create(ctx context.Context, policy *model.Policy) (int64, error) {
	t := time.Now()
	policy.CreationTime = t
	policy.UpdateTime = t

	err := policy.ConvertToDBModel()
	if err != nil {
		return 0, err
	}
	return m.dao.Create(ctx, policy)
}

// List the notification policies, returns the policy list and error
func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	policies := []*model.Policy{}
	persisPolicies, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}

	for _, policy := range persisPolicies {
		err := policy.ConvertFromDBModel()
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// Count the notification policies, returns the count and error
func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

// Get notification policy with specified ID
func (m *manager) Get(ctx context.Context, id int64) (*model.Policy, error) {
	policy, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	if err := policy.ConvertFromDBModel(); err != nil {
		return nil, err
	}
	return policy, err
}

// Update the specified notification policy
func (m *manager) Update(ctx context.Context, policy *model.Policy) error {
	policy.UpdateTime = time.Now()
	err := policy.ConvertToDBModel()
	if err != nil {
		return err
	}
	return m.dao.Update(ctx, policy)
}

// Delete the specified notification policy
func (m *manager) Delete(ctx context.Context, policyID int64) error {
	return m.dao.Delete(ctx, policyID)
}

// GetRelatedPolices get policies including event type in project
func (m *manager) GetRelatedPolices(ctx context.Context, projectID int64, eventType string) ([]*model.Policy, error) {
	policies, err := m.List(ctx, q.New(q.KeyWords{"project_id": projectID}))
	if err != nil {
		return nil, fmt.Errorf("failed to get notification policies with projectID %d: %v", projectID, err)
	}

	var result []*model.Policy

	for _, ply := range policies {
		if !ply.Enabled {
			continue
		}
		for _, t := range ply.EventTypes {
			if t != eventType {
				continue
			}
			result = append(result, ply)
		}
	}

	return result, nil
}
