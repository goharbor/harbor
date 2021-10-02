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
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

// DAO is the data access object interface for user
type DAO interface {
	// Create create a user record in the table, it will return the ID of the user
	Create(ctx context.Context, user *commonmodels.User) (int, error)
	// List list users
	List(ctx context.Context, query *q.Query) ([]*commonmodels.User, error)
	// Count counts the number of users
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Update updates the user record based on the model the parm props are the columns will be updated
	Update(ctx context.Context, user *commonmodels.User, props ...string) error
	// Delete delete user
	Delete(ctx context.Context, userID int) error
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

func init() {
	orm.RegisterModel(
		new(User),
	)
}

type dao struct{}

func (d *dao) Delete(ctx context.Context, userID int) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = ormer.Delete(&User{UserID: userID})
	return err
}

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	query = q.MustClone(query)
	query.Keywords["deleted"] = false
	qs, err := orm.QuerySetterForCount(ctx, &User{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (d *dao) Create(ctx context.Context, user *commonmodels.User) (int, error) {
	if user.UserID > 0 {
		return 0, errors.BadRequestError(nil).WithMessage("user ID is set when creating user: %d", user.UserID)
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(toDBUser(user))
	if err != nil {
		return 0, orm.WrapConflictError(err, "user %s or email %s already exists", user.Username, user.Email)
	}
	return int(id), nil
}

func (d *dao) Update(ctx context.Context, user *commonmodels.User, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(toDBUser(user), props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("user with id %d not found", user.UserID)
	}
	return nil
}

// List list users
func (d *dao) List(ctx context.Context, query *q.Query) ([]*commonmodels.User, error) {
	query = q.MustClone(query)
	query.Keywords["deleted"] = false

	qs, err := orm.QuerySetter(ctx, &User{}, query)
	if err != nil {
		return nil, err
	}

	var users []*User
	if _, err := qs.All(&users); err != nil {
		return nil, err
	}

	var retUsers []*commonmodels.User
	for _, u := range users {
		mU := toCommonUser(u)
		retUsers = append(retUsers, mU)
	}

	return retUsers, nil
}
