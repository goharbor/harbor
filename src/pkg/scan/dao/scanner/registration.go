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

package scanner

import (
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	orm.RegisterModel(new(Registration))
}

// GetTotalOfRegistrations returns the total count of scanner registrations according to the query.
func GetTotalOfRegistrations(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &Registration{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

// AddRegistration adds a new registration
func AddRegistration(ctx context.Context, r *Registration) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	id, err := o.Insert(r)
	if err != nil {
		return 0, orm.WrapConflictError(err, "registration name or url already exists")
	}

	return id, nil
}

// GetRegistration gets the specified registration
func GetRegistration(ctx context.Context, UUID string) (*Registration, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	e := &Registration{}
	qs := o.QueryTable(new(Registration))

	if err := qs.Filter("uuid", UUID).One(e); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			// Not existing case
			return nil, nil
		}
		return nil, err
	}

	return e, nil
}

// UpdateRegistration update the specified registration
func UpdateRegistration(ctx context.Context, r *Registration, cols ...string) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	count, err := o.Update(r, cols...)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no item with UUID %s is updated", r.UUID)
	}

	return nil
}

// DeleteRegistration deletes the registration with the specified UUID
func DeleteRegistration(ctx context.Context, UUID string) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	qt := o.QueryTable(new(Registration))

	// delete with query way
	count, err := qt.Filter("uuid", UUID).Delete()

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no item with UUID %s is deleted", UUID)
	}

	return nil
}

// ListRegistrations lists all the existing registrations
func ListRegistrations(ctx context.Context, query *q.Query) ([]*Registration, error) {
	query = q.MustClone(query)

	qs, err := orm.QuerySetter(ctx, &Registration{}, query)
	if err != nil {
		return nil, err
	}

	// Order the list
	if query.Sorting != "" {
		qs = qs.OrderBy(query.Sorting)
	} else {
		qs = qs.OrderBy("-is_default", "-create_time")
	}

	l := make([]*Registration, 0)
	_, err = qs.All(&l)

	return l, err
}

// SetDefaultRegistration sets the specified registration as default one
func SetDefaultRegistration(ctx context.Context, UUID string) error {
	f := func(ctx context.Context) error {
		o, err := orm.FromContext(ctx)
		if err != nil {
			return err
		}

		var count int64
		qt := o.QueryTable(new(Registration))
		count, err = qt.Filter("uuid", UUID).
			Filter("disabled", false).
			Update(orm.Params{
				"is_default": true,
			})
		if err != nil {
			return err
		}
		if count == 0 {
			return errors.NotFoundError(nil).WithMessage("registration %s not found", UUID)
		}

		qt2 := o.QueryTable(new(Registration))
		_, err = qt2.Exclude("uuid__exact", UUID).
			Filter("is_default", true).
			Update(orm.Params{
				"is_default": false,
			})

		return err
	}

	return orm.WithTransaction(f)(orm.SetTransactionOpNameToContext(ctx, "tx-scan-set-default-registration"))
}

// GetDefaultRegistration gets the default registration
func GetDefaultRegistration(ctx context.Context) (*Registration, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	qt := o.QueryTable(new(Registration))

	e := &Registration{}
	if err := qt.Filter("is_default", true).One(e); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return e, nil
}
