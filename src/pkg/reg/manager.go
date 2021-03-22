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

package reg

import (
	"context"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/reg/dao"
	"github.com/goharbor/harbor/src/replication/model"
	reg "github.com/goharbor/harbor/src/replication/registry"
)

var (
	// Mgr is the global registry manager instance
	Mgr = NewManager()
)

// Manager defines the registry related operations
type Manager interface {
	// Create the registry
	Create(ctx context.Context, registry *model.Registry) (id int64, err error)
	// Count returns the count of registries according to the query
	Count(ctx context.Context, query *q.Query) (count int64, err error)
	// List registries according to the query
	List(ctx context.Context, query *q.Query) (registries []*model.Registry, err error)
	// Get the registry specified by ID
	Get(ctx context.Context, id int64) (registry *model.Registry, err error)
	// Update the specified registry
	Update(ctx context.Context, registry *model.Registry, props ...string) (err error)
	// Delete the registry specified by ID
	Delete(ctx context.Context, id int64) (err error)
}

// NewManager creates an instance of registry manager
func NewManager() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, registry *model.Registry) (int64, error) {
	reg, err := reg.ToDaoModel(registry)
	if err != nil {
		return 0, err
	}
	return m.dao.Create(ctx, reg)
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.Registry, error) {
	registries, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var regs []*model.Registry
	for _, registry := range registries {
		r, err := reg.FromDaoModel(registry)
		if err != nil {
			return nil, err
		}
		regs = append(regs, r)
	}
	return regs, nil
}

func (m *manager) Get(ctx context.Context, id int64) (*model.Registry, error) {
	registry, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return reg.FromDaoModel(registry)
}

func (m *manager) Update(ctx context.Context, registry *model.Registry, props ...string) error {
	reg, err := reg.ToDaoModel(registry)
	if err != nil {
		return err
	}
	return m.dao.Update(ctx, reg, props...)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}
