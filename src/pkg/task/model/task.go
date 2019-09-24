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

package model

import (
	"encoding/json"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

// Task is the object mapping to the job of Jobservice, it holds
// status and check in data
type Task struct {
	ID      int64  `json:"id"`
	GroupID int64  `json:"group_id"`
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	// For order the different statuses in one run
	StatusCode int `json:"status_code"`
	// For differentiating the each retry of the same job
	StatusRevision int64 `json:"status_revision"`
	// Currently the message contains the error detail to explain why the job
	// failed to submit to Jobservice
	Message string   `json:"message"`
	Options *Options `json:"options"`
	// One task can contain multiple check in data with "NotOverrideCheckInData" set to true
	// when submitting the job. Each check in data is stored in json string
	CheckInData []string  `json:"check_in_data"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
}

// From converts a dao.Task to a Task
func (t *Task) From(task *dao.Task) error {
	t.ID = task.ID
	t.GroupID = task.GroupID
	t.JobID = task.JobID
	t.Status = task.Status
	t.StatusCode = task.StatusCode
	t.StatusRevision = task.StatusRevision
	t.Message = task.Message
	t.StartTime = task.StartTime
	t.EndTime = task.EndTime
	if len(task.Options) > 0 {
		options := &Options{}
		if err := options.From(task.Options); err != nil {
			return err
		}
		t.Options = options
	}
	return nil
}

// To converts a Task to a dao.Task
func (t *Task) To() (*dao.Task, error) {
	task := &dao.Task{}
	task.ID = t.ID
	task.GroupID = t.GroupID
	task.JobID = t.JobID
	task.Status = t.Status
	task.StatusCode = t.StatusCode
	task.StatusRevision = t.StatusRevision
	task.Message = t.Message
	task.StartTime = t.StartTime
	task.EndTime = t.EndTime
	if t.Options != nil {
		options, err := t.Options.To()
		if err != nil {
			return nil, err
		}
		task.Options = options
	}
	return task, nil
}

// Options for Task
type Options struct {
	// true: store all check in data
	// false: the later data overrides the last one
	AppendCheckInData bool `json:"append_check_in_data"`
}

// From converts string to Option
func (o *Options) From(option string) error {
	return json.Unmarshal([]byte(option), o)
}

// To converts option to string
func (o *Options) To() (string, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Option item for task
type Option func(options *Options) error

// AppendCheckInData returns an Option to specify how to handle the check in data
func AppendCheckInData(append bool) Option {
	return func(options *Options) error {
		options.AppendCheckInData = append
		return nil
	}
}

// Job object submitted to Jobservice
type Job struct {
	Name       string         `json:"name"`
	Parameters job.Parameters `json:"parameters"`
	Metadata   *job.Metadata  `json:"metadata"`
}

// GroupStatus is the overall status of all tasks that belong to a same group
type GroupStatus struct {
	ID      int64     `json:"id"` // group ID
	Status  string    `json:"status"`
	Total   int64     `json:"total"`    // total count of all tasks
	EndTime time.Time `json:"end_time"` // when the status is a final status, the "EndTime" will be set
	Running int64     `json:"running"`  // the count of tasks that in running status
	Stopped int64     `json:"stopped"`
	Error   int64     `json:"error"`
	Success int64     `json:"success"`
}

// IsFinalStatus determines whether the status provided is a final status
func IsFinalStatus(status string) bool {
	switch status {
	case job.StoppedStatus.String(),
		job.SuccessStatus.String(),
		job.ErrorStatus.String():
		return true
	default:
		return false
	}
}
