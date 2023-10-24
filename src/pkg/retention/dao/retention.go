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

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
)

// CreatePolicy Create Policy
func CreatePolicy(ctx context.Context, p *models.RetentionPolicy) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	return o.Insert(p)
}

// UpdatePolicy Update Policy
func UpdatePolicy(ctx context.Context, p *models.RetentionPolicy, cols ...string) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = o.Update(p, cols...)
	return err
}

// DeletePolicy Update Policy
func DeletePolicy(ctx context.Context, id int64) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	p := &models.RetentionPolicy{
		ID: id,
	}
	_, err = o.Delete(p)
	return err
}

// GetPolicy Get Policy
func GetPolicy(ctx context.Context, id int64) (*models.RetentionPolicy, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	p := &models.RetentionPolicy{
		ID: id,
	}
	if err := o.Read(p); err != nil {
		return nil, err
	}
	return p, nil
}

// ListPolicies list retention policy by query
func ListPolicies(ctx context.Context, query *q.Query) ([]*models.RetentionPolicy, error) {
	plcs := []*models.RetentionPolicy{}
	qs, err := orm.QuerySetter(ctx, &models.RetentionPolicy{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&plcs); err != nil {
		return nil, err
	}
	return plcs, nil
}
