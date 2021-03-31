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
	"github.com/goharbor/harbor/src/pkg/replication/model"
)

// DAO defines the DAO operations of replication policy
type DAO interface {
	// Count returns the count of replication policies according to the query
	Count(ctx context.Context, query *q.Query) (count int64, err error)
	// List the replication policies according to the query
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

// NewDAO creates an instance of DAO
func NewDAO() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.Policy{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	policies := []*model.Policy{}
	qs, err := orm.QuerySetter(ctx, &model.Policy{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&policies); err != nil {
		return nil, err
	}
	return policies, nil
}

func (d *dao) Get(ctx context.Context, id int64) (*model.Policy, error) {
	policy := &model.Policy{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(policy); err != nil {
		if e := orm.AsNotFoundError(err, "replication policy %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return policy, nil
}

func (d *dao) Create(ctx context.Context, policy *model.Policy) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(policy)
	if e := orm.AsConflictError(err, "replication policy %s already exists", policy.Name); e != nil {
		err = e
	}
	return id, err
}

func (d *dao) Update(ctx context.Context, policy *model.Policy, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(policy, props...)
	if e := orm.AsConflictError(err, "replication policy %s already exists", policy.Name); e != nil {
		err = e
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("replication policy %d not found", policy.ID)
	}
	return nil
}

func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.Policy{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("replication policy %d not found", id)
	}
	return nil
}
