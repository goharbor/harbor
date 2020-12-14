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

package scan

import (
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	orm.RegisterModel(new(Report))
}

// DAO is the data access object interface for report
type DAO interface {
	// Create creates new report
	Create(ctx context.Context, r *Report) (int64, error)
	// Delete the reports according to the query
	DeleteMany(ctx context.Context, query q.Query) (int64, error)
	// List lists the reports with given query parameters.
	List(ctx context.Context, query *q.Query) ([]*Report, error)
	// UpdateReportData only updates the `report` column with conditions matched.
	UpdateReportData(ctx context.Context, uuid string, report string) error
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// Create creates new report
func (d *dao) Create(ctx context.Context, r *Report) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, orm.WrapConflictError(err, "a previous scan report found for artifact %s", r.Digest)
	}

	return o.Insert(r)
}

func (d *dao) DeleteMany(ctx context.Context, query q.Query) (int64, error) {
	if len(query.Keywords) == 0 {
		return 0, errors.New("delete all reports at once is not allowed")
	}

	qs, err := orm.QuerySetter(ctx, &Report{}, &query)
	if err != nil {
		return 0, err
	}

	return qs.Delete()
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*Report, error) {
	qs, err := orm.QuerySetter(ctx, &Report{}, query)
	if err != nil {
		return nil, err
	}

	reports := []*Report{}
	if _, err = qs.All(&reports); err != nil {
		return nil, err
	}

	return reports, nil
}

// UpdateReportData only updates the `report` column with conditions matched.
func (d *dao) UpdateReportData(ctx context.Context, uuid string, report string) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	qt := o.QueryTable(new(Report))

	data := make(orm.Params)
	data["report"] = report

	_, err = qt.Filter("uuid", uuid).Update(data)
	return err
}
