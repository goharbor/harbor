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
)

// DAO is the data access object for accessory
type DAO interface {
	// Count returns the total count of accessory according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List accessory according to the query
	List(ctx context.Context, query *q.Query) (accs []*Accessory, err error)
	// Get the accessory specified by ID
	Get(ctx context.Context, id int64) (accessory *Accessory, err error)
	// Create the accessory
	Create(ctx context.Context, accessory *Accessory) (id int64, err error)
	// Delete the accessory specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// DeleteAccessories deletes accessories by query
	DeleteAccessories(ctx context.Context, query *q.Query) (int64, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}
	qs, err := orm.QuerySetterForCount(ctx, &Accessory{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*Accessory, error) {
	accs := []*Accessory{}
	qs, err := orm.QuerySetter(ctx, &Accessory{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&accs); err != nil {
		return nil, err
	}
	return accs, nil
}

func (d *dao) Get(ctx context.Context, id int64) (*Accessory, error) {
	acc := &Accessory{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(acc); err != nil {
		if e := orm.AsNotFoundError(err, "accessory %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return acc, nil
}

func (d *dao) Create(ctx context.Context, acc *Accessory) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(acc)
	if err != nil {
		if e := orm.AsConflictError(err, "accessory %s already exists under the artifact %d",
			acc.Digest, acc.SubjectArtifactID); e != nil {
			err = e
		} else if e := orm.AsForeignKeyError(err, "the accessory %s tries to attach to a non existing artifact %d",
			acc.Digest, acc.SubjectArtifactID); e != nil {
			err = e
		}
	}
	return id, err
}

func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&Accessory{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("accessory %d not found", id)
	}
	return nil
}

func (d *dao) DeleteAccessories(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetter(ctx, &Accessory{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Delete()
}
