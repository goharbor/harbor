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
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego/validation"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/replication"
)

const (
	//RepOpTransfer represents the operation of a job to transfer repository to a remote registry/harbor instance.
	RepOpTransfer string = "transfer"
	//RepOpDelete represents the operation of a job to remove repository from a remote registry/harbor instance.
	RepOpDelete string = "delete"
	//UISecretCookie is the cookie name to contain the UI secret
	UISecretCookie string = "secret"
	//RepTargetTable is the table name for replication targets
	RepTargetTable = "replication_target"
	//RepJobTable is the table name for replication jobs
	RepJobTable = "replication_job"
	//RepPolicyTable is table name for replication policies
	RepPolicyTable = "replication_policy"
)

// RepPolicy is the model for a replication policy, which associate to a project and a target (destination)
type RepPolicy struct {
	ID                        int64        `orm:"pk;auto;column(id)" json:"id"`
	ProjectID                 int64        `orm:"column(project_id)" json:"project_id"`
	ProjectName               string       `orm:"-" json:"project_name,omitempty"`
	TargetID                  int64        `orm:"column(target_id)" json:"target_id"`
	TargetName                string       `orm:"-" json:"target_name,omitempty"`
	Name                      string       `orm:"column(name)" json:"name"`
	Enabled                   int          `orm:"column(enabled)" json:"enabled"`
	Description               string       `orm:"column(description)" json:"description"`
	Trigger                   *RepTrigger  `orm:"-" json:"trigger"`
	TriggerInDB               string       `orm:"column(cron_str)" json:"-"`
	Filters                   []*RepFilter `orm:"-" json:"filters"`
	FiltersInDB               string       `orm:"column(filters)" json:"-"`
	ReplicateExistingImageNow bool         `orm:"-" json:"replicate_existing_image_now"`
	ReplicateDeletion         bool         `orm:"column(replicate_deletion)" json:"replicate_deletion"`
	StartTime                 time.Time    `orm:"column(start_time)" json:"start_time"`
	CreationTime              time.Time    `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime                time.Time    `orm:"column(update_time);auto_now" json:"update_time"`
	ErrorJobCount             int          `orm:"-" json:"error_job_count"`
	Deleted                   int          `orm:"column(deleted)" json:"deleted"`
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

	if r.Trigger != nil {
		r.Trigger.Valid(v)
	}

	for _, filter := range r.Filters {
		filter.Valid(v)
	}

	if err := r.Marshal(); err != nil {
		v.SetError("trigger or filters", err.Error())
	}

	if len(r.TriggerInDB) > 256 {
		v.SetError("trigger", "max length is 256")
	}

	if len(r.FiltersInDB) > 1024 {
		v.SetError("filters", "max length is 1024")
	}
}

// Marshal marshal RepTrigger and RepFilter array to json string
func (r *RepPolicy) Marshal() error {
	if r.Trigger != nil {
		b, err := json.Marshal(r.Trigger)
		if err != nil {
			return err
		}
		r.TriggerInDB = string(b)
	}

	if r.Filters != nil {
		b, err := json.Marshal(r.Filters)
		if err != nil {
			return err
		}
		r.FiltersInDB = string(b)
	}
	return nil
}

// Unmarshal unmarshal json string to RepTrigger and RepFilter array
func (r *RepPolicy) Unmarshal() error {
	if len(r.TriggerInDB) > 0 {
		trigger := &RepTrigger{}
		if err := json.Unmarshal([]byte(r.TriggerInDB), &trigger); err != nil {
			return err
		}
		r.Trigger = trigger
	}

	if len(r.FiltersInDB) > 0 {
		filter := []*RepFilter{}
		if err := json.Unmarshal([]byte(r.FiltersInDB), &filter); err != nil {
			return err
		}
		r.Filters = filter
	}
	return nil
}

// RepFilter holds information for the replication policy filter
type RepFilter struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Valid ...
func (r *RepFilter) Valid(v *validation.Validation) {
	if !(r.Type == replication.FilterItemKindProject ||
		r.Type == replication.FilterItemKindRepository ||
		r.Type == replication.FilterItemKindTag) {
		v.SetError("filter.type", fmt.Sprintf("invalid filter type: %s", r.Type))
	}

	if len(r.Value) == 0 {
		v.SetError("filter.value", "can not be empty")
	}
}

// RepTrigger holds information for the replication policy trigger
type RepTrigger struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// Valid ...
func (r *RepTrigger) Valid(v *validation.Validation) {
	if !(r.Type == replication.TriggerKindManually ||
		r.Type == replication.TriggerKindSchedule ||
		r.Type == replication.TriggerKindImmediately) {
		v.SetError("trigger.type", fmt.Sprintf("invalid trigger type: %s", r.Type))
	}
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
