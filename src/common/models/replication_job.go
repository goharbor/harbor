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

package models

import (
	"time"
)

const (
	// RepOpTransfer represents the operation of a job to transfer repository to a remote registry/harbor instance.
	RepOpTransfer string = "transfer"
	// RepOpDelete represents the operation of a job to remove repository from a remote registry/harbor instance.
	RepOpDelete string = "delete"
	// RepOpSchedule represents the operation of a job to schedule the real replication process
	RepOpSchedule string = "schedule"
	// RegistryTable is the table name for registry
	RegistryTable = "registry"
	// RepJobTable is the table name for replication jobs
	RepJobTable = "replication_job"
	// RepPolicyTable is table name for replication policies
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
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Status       string    `orm:"column(status)" json:"status"`
	Repository   string    `orm:"column(repository)" json:"repository"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	OpUUID       string    `orm:"column(op_uuid)" json:"op_uuid"`
	Operation    string    `orm:"column(operation)" json:"operation"`
	Tags         string    `orm:"column(tags)" json:"-"`
	TagList      []string  `orm:"-" json:"tags"`
	UUID         string    `orm:"column(job_uuid)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName is required by by beego orm to map RepJob to table replication_job
func (r *RepJob) TableName() string {
	return RepJobTable
}

// TableName is required by by beego orm to map RepPolicy to table replication_policy
func (r *RepPolicy) TableName() string {
	return RepPolicyTable
}

// RepJobQuery holds query conditions for replication job
type RepJobQuery struct {
	PolicyID   int64
	OpUUID     string
	Repository string
	Statuses   []string
	Operations []string
	StartTime  *time.Time
	EndTime    *time.Time
	Pagination
}
