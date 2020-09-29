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
	"github.com/goharbor/harbor/src/pkg/audit/model"
)

// DAO is the data access object for audit log
type DAO interface {
	// Create the audit log
	Create(ctx context.Context, access *model.AuditLog) (id int64, err error)
	// Count returns the total count of audit logs according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List audit logs according to the query
	List(ctx context.Context, query *q.Query) (access []*model.AuditLog, err error)
	// Get the audit log specified by ID
	Get(ctx context.Context, id int64) (access *model.AuditLog, err error)
	// Delete the audit log specified by ID
	Delete(ctx context.Context, id int64) (err error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// Count ...
func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}
	qs, err := orm.QuerySetter(ctx, &model.AuditLog{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

// List ...
func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.AuditLog, error) {
	audit := []*model.AuditLog{}
	qs, err := orm.QuerySetter(ctx, &model.AuditLog{}, query)
	if err != nil {
		return nil, err
	}
	qs = qs.OrderBy("-op_time")
	if _, err = qs.All(&audit); err != nil {
		return nil, err
	}
	return audit, nil
}

// Get ...
func (d *dao) Get(ctx context.Context, id int64) (*model.AuditLog, error) {
	audit := &model.AuditLog{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(audit); err != nil {
		if e := orm.AsNotFoundError(err, "audit %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return audit, nil
}

// Create ...
func (d *dao) Create(ctx context.Context, audit *model.AuditLog) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	// the max length of username in database is 255, replace the last
	// three charaters with "..." if the length is greater than 256
	if len(audit.Username) > 255 {
		audit.Username = audit.Username[:252] + "..."
	}
	id, err := ormer.Insert(audit)
	if err != nil {
		return 0, err
	}
	return id, err
}

// Delete ...
func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.AuditLog{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("access %d not found", id)
	}
	return nil
}
