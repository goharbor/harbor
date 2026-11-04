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

package role

import (
	"context"
	"math"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/role/dao"
	"github.com/goharbor/harbor/src/pkg/role/model"
)

var (
	// Mgr is a global variable for the default role manager implementation
	Mgr = NewManager()
)

// Manager ...
type Manager interface {
	// Get ...
	Get(ctx context.Context, id int64) (*model.Role, error)

	// Count returns the total count of robots according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// Create ...
	Create(ctx context.Context, m *model.Role) (int64, error)

	// Delete ...
	Delete(ctx context.Context, id int64) error

	// Update ...
	Update(ctx context.Context, m *model.Role, props ...string) error

	// List ...
	List(ctx context.Context, query *q.Query) ([]*model.Role, error)
}

var _ Manager = &manager{}

type manager struct {
	dao dao.DAO
}

// NewManager return a new instance of defaultRoleManager
func NewManager() Manager {
	return &manager{
		dao: dao.New(),
	}
}

// The role_id column is a 32-bit integer, so an ID outside that range can't
// match any row. Passing it to the database fails with an encoding error
// instead of a not found one.
func checkRoleID(id int64) error {
	if id > math.MaxInt32 || id < math.MinInt32 {
		return errors.NotFoundError(nil).WithMessagef("role %d not found", id)
	}
	return nil
}

// Get ...
func (m *manager) Get(ctx context.Context, id int64) (*model.Role, error) {
	if err := checkRoleID(id); err != nil {
		return nil, err
	}
	return m.dao.Get(ctx, id)
}

// Count ...
func (m *manager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return m.dao.Count(ctx, query)
}

// Create ...
func (m *manager) Create(ctx context.Context, r *model.Role) (int64, error) {
	return m.dao.Create(ctx, r)
}

// Delete ...
func (m *manager) Delete(ctx context.Context, id int64) error {
	if err := checkRoleID(id); err != nil {
		return err
	}
	return m.dao.Delete(ctx, id)
}

// Update ...
func (m *manager) Update(ctx context.Context, r *model.Role, props ...string) error {
	if r != nil {
		if err := checkRoleID(r.ID); err != nil {
			return err
		}
	}
	return m.dao.Update(ctx, r, props...)
}

// List ...
func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.Role, error) {
	return m.dao.List(ctx, query)
}
