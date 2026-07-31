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

	beego_orm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/modelimport/model"
)

// DAO is the data access object for model import policies.
type DAO interface {
	Count(ctx context.Context, query *q.Query) (int64, error)
	Create(ctx context.Context, policy *model.Policy) (int64, error)
	Update(ctx context.Context, policy *model.Policy, props ...string) error
	Get(ctx context.Context, id int64) (*model.Policy, error)
	GetByName(ctx context.Context, projectID int64, name string) (*model.Policy, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, query *q.Query) ([]*model.Policy, error)
}

// New returns an instance of the default DAO.
func New() DAO {
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

func (d *dao) Create(ctx context.Context, policy *model.Policy) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(policy)
	if err != nil {
		if e := orm.AsConflictError(err, "model import policy %s already exists", policy.Name); e != nil {
			err = e
		}
		return 0, err
	}
	return id, nil
}

func (d *dao) Update(ctx context.Context, policy *model.Policy, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	id, err := ormer.Update(policy, props...)
	if err != nil {
		return err
	}
	if id == 0 {
		return errors.NotFoundError(nil).WithMessagef("model import policy %d not found", policy.ID)
	}
	return nil
}

func (d *dao) Get(ctx context.Context, id int64) (*model.Policy, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	policy := &model.Policy{ID: id}
	if err = ormer.Read(policy); err != nil {
		if e := orm.AsNotFoundError(err, "model import policy %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return policy, nil
}

func (d *dao) GetByName(ctx context.Context, projectID int64, name string) (*model.Policy, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	policy := &model.Policy{Name: name, ProjectID: projectID}
	if err = ormer.Read(policy, "Name", "ProjectID"); err != nil {
		if e := orm.AsNotFoundError(err, "model import policy %s not found", name); e != nil {
			err = e
		}
		return nil, err
	}
	return policy, nil
}

func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.Policy{ID: id})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("model import policy %d not found", id)
	}
	return nil
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	var qs beego_orm.QuerySeter
	qs, err := orm.QuerySetter(ctx, &model.Policy{}, query)
	if err != nil {
		return nil, err
	}
	var policies []*model.Policy
	if _, err = qs.All(&policies); err != nil {
		return nil, err
	}
	return policies, nil
}
