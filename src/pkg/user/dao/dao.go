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
	"github.com/goharbor/harbor/src/pkg/user/models"
)

// DAO is the data access object interface for user
type DAO interface {
	// List list users
	List(ctx context.Context, query *q.Query) ([]*models.User, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// List list users
func (d *dao) List(ctx context.Context, query *q.Query) ([]*models.User, error) {
	query = q.MustClone(query)
	query.Keywords["deleted"] = false

	qs, err := orm.QuerySetter(ctx, &models.User{}, query)
	if err != nil {
		return nil, err
	}

	var users []*models.User
	if _, err := qs.All(&users); err != nil {
		return nil, err
	}

	return users, nil
}
