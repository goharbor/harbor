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

package models

import (
	"time"

	"github.com/astaxie/beego/validation"
	"github.com/vmware/harbor/src/common/utils"
)

const (
	//RepOpTransfer represents the operation of a job to transfer repository to a remote registry/harbor instance.
	RepOpTransfer string = "transfer"
	//RepOpDelete represents the operation of a job to remove repository from a remote registry/harbor instance.
	RepOpDelete string = "delete"
	//RepOpSchedule represents the operation of a job to schedule the real replication process
	RepOpSchedule string = "schedule"
	//RepTargetTable is the table name for replication targets
	RepTargetTable = "replication_target"
	//RepJobTable is the table name for replication jobs
	RepJobTable = "replication_job"
	//RepPolicyTable is table name for replication policies
	RepPolicyTable = "replication_policy"
)

// RepPolicy is the model for a replication policy, which associate to a project and a target (destination)
type RepPolicy struct {
	ID                int64     `orm:"pk;auto;column(id)"`
	ProjectID         int64     `orm:"column(project_id)" `
	TargetID          int64     `orm:"column(target_id)"`
	Name              string    `orm:"column(name)"`
	Description       string    `orm:"column(description)"`
	Trigger           string    `orm:"column(cron_str)"`
	Filters           string    `orm:"column(filters)"`
	ReplicateDeletion bool      `orm:"column(replicate_deletion)"`
	CreationTime      time.Time `orm:"column(creation_time);auto_now_add"`
	UpdateTime        time.Time `orm:"column(update_time);auto_now"`
	Deleted           bool      `orm:"column(deleted)"`
}

// RepJob is the model for a replication job, which is the execution unit on job service, currently it is used to transfer/remove
// a repository to/from a remote registry instance.
type RepJob struct {
	ID         int64    `orm:"pk;auto;column(id)" json:"id"`
	Status     string   `orm:"column(status)" json:"status"`
	Repository string   `orm:"column(repository)" json:"repository"`
	PolicyID   int64    `orm:"column(policy_id)" json:"policy_id"`
	Operation  string   `orm:"column(operation)" json:"operation"`
	Tags       string   `orm:"column(tags)" json:"-"`
	TagList    []string `orm:"-" json:"tags"`
	UUID       string   `orm:"column(job_uuid)" json:"-"`
	//	Policy       RepPolicy `orm:"-" json:"policy"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// RepTarget is the model for a replication targe, i.e. destination, which wraps the endpoint URL and username/password of a remote registry.
type RepTarget struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	URL          string    `orm:"column(url)" json:"endpoint"`
	Name         string    `orm:"column(name)" json:"name"`
	Username     string    `orm:"column(username)" json:"username"`
	Password     string    `orm:"column(password)" json:"password"`
	Type         int       `orm:"column(target_type)" json:"type"`
	Insecure     bool      `orm:"column(insecure)" json:"insecure"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// Valid ...
func (r *RepTarget) Valid(v *validation.Validation) {
	if len(r.Name) == 0 {
		v.SetError("name", "can not be empty")
	}

	if len(r.Name) > 64 {
		v.SetError("name", "max length is 64")
	}

	url, err := utils.ParseEndpoint(r.URL)
	if err != nil {
		v.SetError("endpoint", err.Error())
	} else {
		// Prevent SSRF security issue #3755
		r.URL = url.Scheme + "://" + url.Host + url.Path
		if len(r.URL) > 64 {
			v.SetError("endpoint", "max length is 64")
		}
	}

	// password is encoded using base64, the length of this field
	// in DB is 64, so the max length in request is 48
	if len(r.Password) > 48 {
		v.SetError("password", "max length is 48")
	}
}

//TableName is required by by beego orm to map RepTarget to table replication_target
func (r *RepTarget) TableName() string {
	return RepTargetTable
}

//TableName is required by by beego orm to map RepJob to table replication_job
func (r *RepJob) TableName() string {
	return RepJobTable
}

//TableName is required by by beego orm to map RepPolicy to table replication_policy
func (r *RepPolicy) TableName() string {
	return RepPolicyTable
}

// RepJobQuery holds query conditions for replication job
type RepJobQuery struct {
	PolicyID   int64
	Repository string
	Statuses   []string
	Operations []string
	StartTime  *time.Time
	EndTime    *time.Time
	Pagination
}
