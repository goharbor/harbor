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
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/pkg/errors"
)

func init() {
	orm.RegisterModel(new(Report))
}

// CreateReport creates new report
func CreateReport(r *Report) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(r)
}

// DeleteReport deletes the given report
func DeleteReport(uuid string) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	// Delete report with query way
	count, err := qt.Filter("uuid", uuid).Delete()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no report with uuid %s deleted", uuid)
	}

	return nil
}

// ListReports lists the reports with given query parameters.
// Keywords in query here will be enforced with `exact` way.
func ListReports(query *q.Query) ([]*Report, error) {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	if query != nil {
		if len(query.Keywords) > 0 {
			for k, v := range query.Keywords {
				if vv, ok := v.([]interface{}); ok {
					qt = qt.Filter(fmt.Sprintf("%s__in", k), vv...)
				}

				qt = qt.Filter(k, v)
			}
		}

		if query.PageNumber > 0 && query.PageSize > 0 {
			qt = qt.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
		}
	}

	l := make([]*Report, 0)
	_, err := qt.All(&l)

	return l, err
}

// UpdateReportData only updates the `report` column with conditions matched.
func UpdateReportData(uuid string, report string, statusRev int64) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	data := make(orm.Params)
	data["report"] = report
	data["status_rev"] = statusRev

	count, err := qt.Filter("uuid", uuid).
		Filter("status_rev__lte", statusRev).Update(data)

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no report with uuid %s updated", uuid)
	}

	return nil
}

// UpdateReportStatus updates the report `status` with conditions matched.
func UpdateReportStatus(trackID string, status string, statusCode int, statusRev int64) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	data := make(orm.Params)
	data["status"] = status
	data["status_code"] = statusCode
	data["status_rev"] = statusRev

	// Technically it is not correct, just to avoid changing interface and adding more code.
	// running==2
	if statusCode > 2 {
		data["end_time"] = time.Now().UTC()
	}

	count, err := qt.Filter("track_id", trackID).
		Filter("status_rev__lte", statusRev).
		Filter("status_code__lte", statusCode).Update(data)

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no report with track_id %s updated", trackID)
	}

	return nil
}

// UpdateJobID updates the report `job_id` column
func UpdateJobID(trackID string, jobID string) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	params := make(orm.Params, 1)
	params["job_id"] = jobID
	_, err := qt.Filter("track_id", trackID).Update(params)

	return err
}
