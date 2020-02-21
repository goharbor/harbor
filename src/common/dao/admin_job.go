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
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// AddAdminJob ...
func AddAdminJob(job *models.AdminJob) (int64, error) {
	o := GetOrmer()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	sql := "insert into admin_job (job_name, job_parameters, job_kind, status, job_uuid, cron_str, creation_time, update_time) values (?, ?, ?, ?, ?, ?, ?, ?) RETURNING id"
	var id int64
	now := time.Now()
	err := o.Raw(sql, job.Name, job.Parameters, job.Kind, job.Status, job.UUID, job.Cron, now, now).QueryRow(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetAdminJob ...
func GetAdminJob(id int64) (*models.AdminJob, error) {
	o := GetOrmer()
	aj := models.AdminJob{ID: id}
	err := o.Read(&aj)
	if err == orm.ErrNoRows {
		return nil, err
	}
	return &aj, nil
}

// DeleteAdminJob ...
func DeleteAdminJob(id int64) error {
	o := GetOrmer()
	_, err := o.Raw(`update admin_job 
		set deleted = true where id = ?`, id).Exec()
	return err
}

// UpdateAdminJobStatus ...
func UpdateAdminJobStatus(id int64, status string, statusCode uint16, revision int64) error {
	o := GetOrmer()
	qt := o.QueryTable(&models.AdminJob{})

	// The generated sql statement example:{
	//
	// UPDATE "admin_job" SET "update_time" = $1, "status" = $2, "status_code" = $3, "revision" = $4
	// WHERE "id" IN ( SELECT T0."id" FROM "admin_job" T0 WHERE
	// ( T0."revision" = $5 AND T0."status_code" < $6 ) OR ( T0."revision" < $7 )
	// AND T0."id" = $8  )
	//
	// }
	cond := orm.NewCondition()
	c1 := cond.And("revision", revision).And("status_code__lt", statusCode)
	c2 := cond.And("revision__lt", revision)
	c := cond.AndCond(c1).OrCond(c2)

	data := make(orm.Params)
	data["status"] = status
	data["status_code"] = statusCode
	data["revision"] = revision
	data["update_time"] = time.Now()

	n, err := qt.SetCond(c).Filter("id", id).Update(data)

	if n == 0 {
		log.Warningf("no records are updated when updating admin job %d", id)
	}

	return err
}

// SetAdminJobUUID ...
func SetAdminJobUUID(id int64, uuid string) error {
	o := GetOrmer()
	j := models.AdminJob{
		ID:   id,
		UUID: uuid,
	}
	n, err := o.Update(&j, "UUID")
	if n == 0 {
		log.Warningf("no records are updated when updating admin job %d", id)
	}
	return err
}

// GetTop10AdminJobsOfName ...
func GetTop10AdminJobsOfName(name string) ([]*models.AdminJob, error) {
	o := GetOrmer()
	jobs := []*models.AdminJob{}
	n, err := o.Raw(`select * from admin_job 
		where deleted = false and job_name = ? order by id desc limit 10`, name).QueryRows(&jobs)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, nil
	}
	return jobs, err
}

// GetAdminJobs get admin jobs bases on query conditions
func GetAdminJobs(query *models.AdminJobQuery) ([]*models.AdminJob, error) {
	adjs := []*models.AdminJob{}
	qs := adminQueryConditions(query)
	if query.Size > 0 {
		qs = qs.Limit(query.Size)
		if query.Page > 0 {
			qs = qs.Offset((query.Page - 1) * query.Size)
		}
	}
	_, err := qs.All(&adjs)
	return adjs, err
}

// adminQueryConditions
func adminQueryConditions(query *models.AdminJobQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.AdminJob{})

	if query.ID > 0 {
		qs = qs.Filter("ID", query.ID)
	}
	if len(query.Kind) > 0 {
		qs = qs.Filter("Kind", query.Kind)
	}
	if len(query.Name) > 0 {
		qs = qs.Filter("Name", query.Name)
	}
	if len(query.Status) > 0 {
		qs = qs.Filter("Status", query.Status)
	}
	if len(query.UUID) > 0 {
		qs = qs.Filter("UUID", query.UUID)
	}
	qs = qs.Filter("Deleted", false)
	return qs.OrderBy("-ID")

}
