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

package dao

import (
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/role/model"
)

// DAO defines the interface to access the role data model
type DAO interface {
	// Create ...
	Create(ctx context.Context, r *model.Role) (int64, error)

	// Update ...
	Update(ctx context.Context, r *model.Role, props ...string) error

	// Get ...
	Get(ctx context.Context, id int64) (*model.Role, error)

	// Count returns the total count of roles according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// List ...
	List(ctx context.Context, query *q.Query) ([]*model.Role, error)

	// Delete ...
	Delete(ctx context.Context, id int64) error
}

// New creates a default implementation for Dao
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, r *model.Role) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(r)
	if err != nil {
		return 0, orm.WrapConflictError(err, "role %s already exists", r.Name)
	}
	return id, err
}

func (d *dao) Update(ctx context.Context, r *model.Role, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(r, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("role %d not found", r.ID)
	}
	return nil
}

func (d *dao) Get(ctx context.Context, id int64) (*model.Role, error) {
	r := &model.Role{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(r); err != nil {
		return nil, orm.WrapNotFoundError(err, "role %d not found", id)
	}
	return r, nil
}

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.Role{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.Role{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("role %d not found", id)
	}
	return nil
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.Role, error) {
	roles := []*model.Role{}

	qs, err := orm.QuerySetter(ctx, &model.Role{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&roles); err != nil {
		return nil, err
	}
	return roles, nil
}
