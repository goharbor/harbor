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

package replication

import (
	"context"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/replication/dao"
)

var (
	// Mgr is the global replication policy manager instance
	Mgr = NewManager()
)

// Manager defines the replication policy related operations
type Manager interface {
	// Count returns the count of replication policies according to the query
	Count(ctx context.Context, query *q.Query) (count int64, err error)
	// List replication policies according to the query
	List(ctx context.Context, query *q.Query) (policies []*Policy, err error)
	// Get the replication policy specified by ID
	Get(ctx context.Context, id int64) (policy *Policy, err error)
	// Create the replication policy
	Create(ctx context.Context, policy *Policy) (id int64, err error)
	// Update the specified replication policy
	Update(ctx context.Context, policy *Policy, props ...string) (err error)
	// Delete the replication policy specified by ID
	Delete(ctx context.Context, id int64) (err error)
}

// NewManager creates an instance of replication policy manager
func NewManager() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*Policy, error) {
	policies, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var result []*Policy
	for _, policy := range policies {
		p := &Policy{}
		if err = p.From(policy); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (m *manager) Get(ctx context.Context, id int64) (*Policy, error) {
	policy, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	p := &Policy{}
	if err = p.From(policy); err != nil {
		return nil, err
	}
	return p, nil
}

func (m *manager) Create(ctx context.Context, policy *Policy) (int64, error) {
	p, err := policy.To()
	if err != nil {
		return 0, err
	}
	return m.dao.Create(ctx, p)
}

func (m *manager) Update(ctx context.Context, policy *Policy, props ...string) error {
	p, err := policy.To()
	if err != nil {
		return err
	}
	return m.dao.Update(ctx, p, props...)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}
