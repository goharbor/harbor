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
	"github.com/goharbor/harbor/src/pkg/immutable/dao/model"
)

// DAO defines the interface to access the ImmutableRule data model
type DAO interface {
	CreateImmutableRule(ctx context.Context, ir *model.ImmutableRule) (int64, error)
	UpdateImmutableRule(ctx context.Context, projectID int64, ir *model.ImmutableRule) error
	ToggleImmutableRule(ctx context.Context, id int64, status bool) error
	GetImmutableRule(ctx context.Context, id int64) (*model.ImmutableRule, error)
	Count(ctx context.Context, query *q.Query) (int64, error)
	ListImmutableRules(ctx context.Context, query *q.Query) ([]*model.ImmutableRule, error)
	DeleteImmutableRule(ctx context.Context, id int64) error
}

// New creates a default implementation for DAO
func New() DAO {
	return &iDao{}
}

type iDao struct{}

// CreateImmutableRule creates the Immutable Rule
func (i *iDao) CreateImmutableRule(ctx context.Context, ir *model.ImmutableRule) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	ir.Disabled = false
	id, err := ormer.Insert(ir)
	if err != nil {
		if e := orm.AsConflictError(err, "immutable rule already exists"); e != nil {
			err = e
		}
	}
	return id, err
}

// UpdateImmutableRule update the immutable rules
func (i *iDao) UpdateImmutableRule(ctx context.Context, projectID int64, ir *model.ImmutableRule) error {
	ir.ProjectID = projectID
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(ir, "TagFilter")
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("immutable %d not found", ir.ID)
	}
	return nil
}

// ToggleImmutableRule enable/disable immutable rules
func (i *iDao) ToggleImmutableRule(ctx context.Context, id int64, status bool) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	ir := &model.ImmutableRule{ID: id, Disabled: status}
	n, err := ormer.Update(ir, "Disabled")
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("immutable %d not found", ir.ID)
	}
	return nil
}

// GetImmutableRule get immutable rule
func (i *iDao) GetImmutableRule(ctx context.Context, id int64) (*model.ImmutableRule, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	ir := &model.ImmutableRule{ID: id}
	if err = ormer.Read(ir); err != nil {
		if e := orm.AsNotFoundError(err, "immutable rule %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return ir, nil
}

// QueryImmutableRuleByProjectID get all immutable rule by project
func (i *iDao) ListImmutableRules(ctx context.Context, query *q.Query) ([]*model.ImmutableRule, error) {
	rules := []*model.ImmutableRule{}
	qs, err := orm.QuerySetter(ctx, &model.ImmutableRule{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// Count ...
func (i *iDao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.ImmutableRule{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

// DeleteImmutableRule delete the immutable rule
func (i *iDao) DeleteImmutableRule(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	ir := &model.ImmutableRule{ID: id}

	n, err := ormer.Delete(ir)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("immutable rule %d not found", id)
	}
	return nil
}
