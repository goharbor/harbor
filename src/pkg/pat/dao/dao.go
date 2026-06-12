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
	"github.com/goharbor/harbor/src/pkg/pat/model"
)

// DAO interface defines data access methods for personal access tokens
type DAO interface {
	Create(ctx context.Context, pat *model.PersonalAccessToken) (int64, error)
	Get(ctx context.Context, id int64) (*model.PersonalAccessToken, error)
	Update(ctx context.Context, pat *model.PersonalAccessToken, props ...string) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, query *q.Query) ([]*model.PersonalAccessToken, error)
	Count(ctx context.Context, query *q.Query) (total int64, err error)
}

// New returns a new DAO instance
func New() DAO {
	return &dao{}
}

type dao struct{}

// Create creates a new personal access token
func (d *dao) Create(ctx context.Context, pat *model.PersonalAccessToken) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(pat)
	if err != nil {
		return 0, orm.WrapConflictError(err, "personal access token %d:%s already exists", pat.UserID, pat.Name)
	}
	return id, nil
}

// Get returns a personal access token by ID
func (d *dao) Get(ctx context.Context, id int64) (*model.PersonalAccessToken, error) {
	pat := &model.PersonalAccessToken{ID: id}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(pat); err != nil {
		return nil, orm.WrapNotFoundError(err, "personal access token %d not found", id)
	}
	return pat, nil
}

// Update updates a personal access token
func (d *dao) Update(ctx context.Context, pat *model.PersonalAccessToken, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	if len(props) == 0 {
		props = []string{"name", "description", "expires_at", "disabled"}
	}
	n, err := ormer.Update(pat, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("personal access token %d not found", pat.ID)
	}
	return nil
}

// Delete deletes a personal access token by ID
func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.PersonalAccessToken{ID: id})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("personal access token %d not found", id)
	}
	return nil
}

// List lists personal access tokens based on the query
func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.PersonalAccessToken, error) {
	pats := []*model.PersonalAccessToken{}
	qs, err := orm.QuerySetter(ctx, &model.PersonalAccessToken{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&pats); err != nil {
		return nil, err
	}
	return pats, nil
}

// Count returns the count of personal access tokens matching the query
func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.PersonalAccessToken{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}
