// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// AddAdminJob ...
func AddAdminJob(job *models.AdminJob) (int64, error) {
	o := GetOrmer()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	return o.Insert(job)
}

// GetAdminJob ...
func GetAdminJob(id int64) (*models.AdminJob, error) {
	o := GetOrmer()
	aj := models.AdminJob{ID: id}
	err := o.Read(&aj)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &aj, nil
}

// DeleteAdminJob ...
func DeleteAdminJob(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.AdminJob{ID: id})
	return err
}

// UpdateAdminJobStatus ...
func UpdateAdminJobStatus(id int64, status string) error {
	o := GetOrmer()
	j := models.AdminJob{
		ID:         id,
		Status:     status,
		UpdateTime: time.Now(),
	}
	n, err := o.Update(&j, "Status", "UpdateTime")
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
func adminQueryConditions(query ...*models.AdminJobQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(new(models.AdminJob))
	if len(query) == 0 || query[0] == nil {
		return qs
	}

	q := query[0]
	if len(q.Kind) > 0 {
		qs = qs.Filter("Kind", q.Kind)
	}
	if len(q.Name) > 0 {
		qs = qs.Filter("Name", q.Name)
	}
	if len(q.Status) > 0 {
		qs = qs.Filter("Status", q.Status)
	}
	return qs
}
