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
	"github.com/goharbor/harbor/src/pkg/replication/model"
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
	List(ctx context.Context, query *q.Query) (policies []*model.Policy, err error)
	// Get the replication policy specified by ID
	Get(ctx context.Context, id int64) (policy *model.Policy, err error)
	// Create the replication policy
	Create(ctx context.Context, policy *model.Policy) (id int64, err error)
	// Update the specified replication policy
	Update(ctx context.Context, policy *model.Policy, props ...string) (err error)
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

func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	return m.dao.List(ctx, query)
}

func (m *manager) Get(ctx context.Context, id int64) (*model.Policy, error) {
	return m.dao.Get(ctx, id)
}

func (m *manager) Create(ctx context.Context, policy *model.Policy) (int64, error) {
	return m.dao.Create(ctx, policy)
}

func (m *manager) Update(ctx context.Context, policy *model.Policy, props ...string) error {
	return m.dao.Update(ctx, policy, props...)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}
