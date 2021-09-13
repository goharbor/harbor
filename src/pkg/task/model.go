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

package task

import (
	"encoding/json"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

// const definitions
const (
	ExecutionTriggerManual   = "MANUAL"
	ExecutionTriggerSchedule = "SCHEDULE"
	ExecutionTriggerEvent    = "EVENT"
)

// Execution is one run for one action. It contains one or more tasks and provides the summary view of the tasks
type Execution struct {
	ID int64 `json:"id"`
	// indicate the execution type: replication/GC/retention/scan/etc.
	VendorType string `json:"vendor_type"`
	// the ID of vendor policy/rule/etc. e.g. replication policy ID
	VendorID int64  `json:"vendor_id"`
	Status   string `json:"status"`
	// the detail message to explain the status in some cases. e.g.
	// 1. After creating the execution, there may be some errors before creating tasks, the
	// "StatusMessage" can contain the error message
	// 2. The execution may contain no tasks, "StatusMessage" can be used to explain the case
	StatusMessage string       `json:"status_message"`
	Metrics       *dao.Metrics `json:"metrics"`
	// trigger type: manual/schedule/event
	Trigger string `json:"trigger"`
	// the customized attributes for different kinds of consumers
	ExtraAttrs map[string]interface{} `json:"extra_attrs"`
	StartTime  time.Time              `json:"start_time"`
	UpdateTime time.Time              `json:"update_time"`
	EndTime    time.Time              `json:"end_time"`
}

// IsOnGoing returns true when the execution is running
func (exec *Execution) IsOnGoing() bool {
	switch job.Status(exec.Status) {
	case job.RunningStatus:
		return true
	default:
		return false
	}
}

// Task is the unit for running. It stores the jobservice job records and related information
type Task struct {
	ID int64 `json:"id"`
	// indicate the task type: replication/GC/retention/scan/etc.
	VendorType  string `json:"vendor_type"`
	ExecutionID int64  `json:"execution_id"`
	Status      string `json:"status"`
	// the detail message to explain the status in some cases. e.g.
	// When the job is failed to submit to jobservice, this field can be used to explain the reason
	StatusMessage string `json:"status_message"`
	// the underlying job may retry several times
	RunCount int32 `json:"run_count"`
	// the ID of jobservice job
	JobID string `json:"job_id"`
	// the customized attributes for different kinds of consumers
	ExtraAttrs map[string]interface{} `json:"extra_attrs"`
	// the time that the task record created
	CreationTime time.Time `json:"creation_time"`
	// the time that the underlying job starts
	StartTime      time.Time `json:"start_time"`
	UpdateTime     time.Time `json:"update_time"`
	EndTime        time.Time `json:"end_time"`
	StatusRevision int64     `json:"status_revision"`
}

// From constructs a task from DAO model
func (t *Task) From(task *dao.Task) {
	t.ID = task.ID
	t.VendorType = task.VendorType
	t.ExecutionID = task.ExecutionID
	t.Status = task.Status
	t.StatusMessage = task.StatusMessage
	t.RunCount = task.RunCount
	t.JobID = task.JobID
	t.CreationTime = task.CreationTime
	t.StartTime = task.StartTime
	t.UpdateTime = task.UpdateTime
	t.EndTime = task.EndTime
	t.StatusRevision = task.StatusRevision
	if len(task.ExtraAttrs) > 0 {
		extras := map[string]interface{}{}
		if err := json.Unmarshal([]byte(task.ExtraAttrs), &extras); err != nil {
			log.Errorf("failed to unmarshal the extra attributes of task %d: %v", task.ID, err)
			return
		}
		t.ExtraAttrs = extras
	}
}

// GetStringFromExtraAttrs returns the string value specified by key
func (t *Task) GetStringFromExtraAttrs(key string) string {
	if len(t.ExtraAttrs) == 0 {
		return ""
	}
	rt, exist := t.ExtraAttrs[key]
	if !exist {
		return ""
	}
	str, ok := rt.(string)
	if !ok {
		return ""
	}
	return str
}

// GetBoolFromExtraAttrs returns the bool value specified by key
func (t *Task) GetBoolFromExtraAttrs(key string) bool {
	if len(t.ExtraAttrs) == 0 {
		return false
	}
	rt, exist := t.ExtraAttrs[key]
	if !exist {
		return false
	}
	b, ok := rt.(bool)
	if !ok {
		return false
	}
	return b
}

// GetNumFromExtraAttrs returns the num value specified by key
func (t *Task) GetNumFromExtraAttrs(key string) float64 {
	if len(t.ExtraAttrs) == 0 {
		return 0
	}
	rt, exist := t.ExtraAttrs[key]
	if !exist {
		return 0
	}
	v, ok := rt.(float64)
	if !ok {
		return 0
	}
	return v
}

// Job is the model represents the requested jobservice job
type Job struct {
	Name       string
	Parameters job.Parameters
	Metadata   *job.Metadata
}
