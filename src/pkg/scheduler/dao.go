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

package scheduler

import (
	"context"
	"time"

	beegoorm "github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	beegoorm.RegisterModel(&schedule{})
}

type schedule struct {
	ID                int64     `orm:"pk;auto;column(id)"`
	VendorType        string    `orm:"column(vendor_type)"`
	VendorID          int64     `orm:"column(vendor_id)"`
	CRONType          string    `orm:"column(cron_type)"`
	CRON              string    `orm:"column(cron)"`
	Revision          int64     `orm:"column(revision)"` // to identity the duplicated checkin hook from jobservice
	ExtraAttrs        string    `orm:"column(extra_attrs)"`
	CallbackFuncName  string    `orm:"column(callback_func_name)"`
	CallbackFuncParam string    `orm:"column(callback_func_param)"`
	CreationTime      time.Time `orm:"column(creation_time)"`
	UpdateTime        time.Time `orm:"column(update_time)"`
}

// DAO is the data access object interface for schedule
type DAO interface {
	Create(ctx context.Context, s *schedule) (id int64, err error)
	List(ctx context.Context, query *q.Query) (schedules []*schedule, err error)
	Get(ctx context.Context, id int64) (s *schedule, err error)
	Delete(ctx context.Context, id int64) (err error)
	Update(ctx context.Context, s *schedule, props ...string) (err error)
	UpdateRevision(ctx context.Context, id, revision int64) (n int64, err error)
}

type dao struct{}

func (d *dao) Create(ctx context.Context, schedule *schedule) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(schedule)
	if err != nil {
		if e := orm.AsConflictError(err, "schedule with vendor type: %s vendor ID: %d already exists",
			schedule.VendorType, schedule.VendorID); e != nil {
			err = e
		}
		return 0, err
	}
	return id, nil
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*schedule, error) {
	qs, err := orm.QuerySetter(ctx, &schedule{}, query)
	if err != nil {
		return nil, err
	}

	schedules := []*schedule{}
	if _, err = qs.All(&schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (d *dao) Get(ctx context.Context, id int64) (*schedule, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	schedule := &schedule{
		ID: id,
	}
	if err = ormer.Read(schedule); err != nil {
		if e := orm.AsNotFoundError(err, "schedule %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return schedule, nil
}

func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&schedule{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("schedule %d not found", id)
	}

	return nil
}
func (d *dao) Update(ctx context.Context, schedule *schedule, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(schedule, props...)
	if err != nil {
		if e := orm.AsConflictError(err, "schedule with the same vendor type and vendor ID already exists"); e != nil {
			err = e
		}
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("schedule %d not found", schedule.ID)
	}
	return nil
}

func (d *dao) UpdateRevision(ctx context.Context, id, revision int64) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	return ormer.QueryTable(&schedule{}).Filter("ID", id).Filter("Revision__lt", revision).Update(beegoorm.Params{
		"Revision": revision,
	})
}
