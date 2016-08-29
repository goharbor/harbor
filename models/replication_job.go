/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package models

import (
	"time"

	"github.com/astaxie/beego/validation"
	"github.com/vmware/harbor/utils"
)

const (
	//JobPending ...
	JobPending string = "pending"
	//JobRunning ...
	JobRunning string = "running"
	//JobError ...
	JobError string = "error"
	//JobStopped ...
	JobStopped string = "stopped"
	//JobFinished ...
	JobFinished string = "finished"
	//JobCanceled ...
	JobCanceled string = "canceled"
	//JobRetrying indicate the job needs to be retried, it will be scheduled to the end of job queue by statemachine after an interval.
	JobRetrying string = "retrying"
	//JobContinue is the status returned by statehandler to tell statemachine to move to next possible state based on trasition table.
	JobContinue string = "_continue"
	//RepOpTransfer represents the operation of a job to transfer repository to a remote registry/harbor instance.
	RepOpTransfer string = "transfer"
	//RepOpDelete represents the operation of a job to remove repository from a remote registry/harbor instance.
	RepOpDelete string = "delete"
	//UISecretCookie is the cookie name to contain the UI secret
	UISecretCookie string = "uisecret"
)

// RepPolicy is the model for a replication policy, which associate to a project and a target (destination)
type RepPolicy struct {
	ID          int64  `orm:"column(id)" json:"id"`
	ProjectID   int64  `orm:"column(project_id)" json:"project_id"`
	ProjectName string `json:"project_name,omitempty"`
	TargetID    int64  `orm:"column(target_id)" json:"target_id"`
	TargetName  string `json:"target_name,omitempty"`
	Name        string `orm:"column(name)" json:"name"`
	//	Target       RepTarget `orm:"-" json:"target"`
	Enabled       int       `orm:"column(enabled)" json:"enabled"`
	Description   string    `orm:"column(description)" json:"description"`
	CronStr       string    `orm:"column(cron_str)" json:"cron_str"`
	StartTime     time.Time `orm:"column(start_time)" json:"start_time"`
	CreationTime  time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime    time.Time `orm:"column(update_time);auto_now" json:"update_time"`
	ErrorJobCount int       `json:"error_job_count"`
	Deleted       int       `orm:"column(deleted)" json:"deleted"`
}

// Valid ...
func (r *RepPolicy) Valid(v *validation.Validation) {
	if len(r.Name) == 0 {
		v.SetError("name", "can not be empty")
	}

	if len(r.Name) > 256 {
		v.SetError("name", "max length is 256")
	}

	if r.ProjectID <= 0 {
		v.SetError("project_id", "invalid")
	}

	if r.TargetID <= 0 {
		v.SetError("target_id", "invalid")
	}

	if r.Enabled != 0 && r.Enabled != 1 {
		v.SetError("enabled", "must be 0 or 1")
	}

	if len(r.CronStr) > 256 {
		v.SetError("cron_str", "max length is 256")
	}
}

// RepJob is the model for a replication job, which is the execution unit on job service, currently it is used to transfer/remove
// a repository to/from a remote registry instance.
type RepJob struct {
	ID         int64    `orm:"column(id)" json:"id"`
	Status     string   `orm:"column(status)" json:"status"`
	Repository string   `orm:"column(repository)" json:"repository"`
	PolicyID   int64    `orm:"column(policy_id)" json:"policy_id"`
	Operation  string   `orm:"column(operation)" json:"operation"`
	Tags       string   `orm:"column(tags)" json:"-"`
	TagList    []string `orm:"-" json:"tags"`
	//	Policy       RepPolicy `orm:"-" json:"policy"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// RepTarget is the model for a replication targe, i.e. destination, which wraps the endpoint URL and username/password of a remote registry.
type RepTarget struct {
	ID           int64     `orm:"column(id)" json:"id"`
	URL          string    `orm:"column(url)" json:"endpoint"`
	Name         string    `orm:"column(name)" json:"name"`
	Username     string    `orm:"column(username)" json:"username"`
	Password     string    `orm:"column(password)" json:"password"`
	Type         int       `orm:"column(target_type)" json:"type"`
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

	if len(r.URL) == 0 {
		v.SetError("endpoint", "can not be empty")
	}

	r.URL = utils.FormatEndpoint(r.URL)

	if len(r.URL) > 64 {
		v.SetError("endpoint", "max length is 64")
	}

	// password is encoded using base64, the length of this field
	// in DB is 64, so the max length in request is 48
	if len(r.Password) > 48 {
		v.SetError("password", "max length is 48")
	}
}

//TableName is required by by beego orm to map RepTarget to table replication_target
func (r *RepTarget) TableName() string {
	return "replication_target"
}

//TableName is required by by beego orm to map RepJob to table replication_job
func (r *RepJob) TableName() string {
	return "replication_job"
}

//TableName is required by by beego orm to map RepPolicy to table replication_policy
func (r *RepPolicy) TableName() string {
	return "replication_policy"
}
