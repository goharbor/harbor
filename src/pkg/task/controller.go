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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/manager"
	"github.com/goharbor/harbor/src/pkg/task/model"
)

var (
	// Ctl is a global variable for the default task controller implementation
	Ctl = NewController()
)

// Controller is an interface for task management. A task is corresponding to a job
// in Jobservice.
// All Jobservice job related operations should be covered by the controller to reduce
// the complexity
type Controller interface {
	// Submit creates the task record in database and submits a job to the Jobservice
	Submit(job *model.Job, options ...model.Option) (int64, error)
	Get(id int64) (*model.Task, error)
	Stop(id int64) error
	Delete(id int64) error
	GetLog(id int64) ([]byte, error)
	// Calculate the status of a group of tasks specified by the id.
	CalculateTaskGroupStatus(groupID int64) (*model.GroupStatus, error)
}

// NewController creates a new task controller
func NewController() Controller {
	return &controller{
		mgr: manager.New(),
	}
}

type controller struct {
	mgr manager.Manager
}

func (c *controller) Submit(jb *model.Job, options ...model.Option) (int64, error) {
	if jb == nil {
		return 0, errors.New("empty job")
	}
	// create task record
	now := time.Now()
	task := &model.Task{
		Status:    job.PendingStatus.String(),
		StartTime: now,
		Options:   &model.Options{},
	}
	for _, option := range options {
		if err := option(task.Options); err != nil {
			return 0, err
		}
	}
	id, err := c.mgr.Create(task)
	if err != nil {
		return 0, err
	}
	task.ID = id

	// submit job to Jobservice
	j := &models.JobData{
		Name:       jb.Name,
		StatusHook: fmt.Sprintf("%s/service/notifications/tasks/%d", getInternalCoreURL(), id),
	}
	if jb.Parameters != nil {
		j.Parameters = models.Parameters(jb.Parameters)
	}
	if jb.Metadata != nil {
		j.Metadata = &models.JobMetadata{
			JobKind:       jb.Metadata.JobKind,
			ScheduleDelay: jb.Metadata.ScheduleDelay,
			Cron:          jb.Metadata.Cron,
			IsUnique:      jb.Metadata.IsUnique,
		}
	}
	cols := []string{}
	jobID, err := cjob.GlobalClient.SubmitJob(j)
	if err != nil {
		// failed to submit the job, mark the task failure
		task.Status = job.ErrorStatus.String()
		task.StatusCode = job.ErrorStatus.Code()
		task.Message = err.Error()
		task.EndTime = now
		cols = append(cols, "Status", "StatusCode", "Message", "EndTime")
	} else {
		// success to submit the job, update the job ID
		task.JobID = jobID
		cols = append(cols, "JobID")
	}
	if err = c.mgr.Update(task, cols...); err != nil {
		log.Errorf("failed update the task %d: %v", id, err)
	}
	return id, nil
}

func (c *controller) Get(id int64) (*model.Task, error) {
	return c.mgr.Get(id)
}

func (c *controller) Stop(id int64) error {
	task, err := c.mgr.Get(id)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task %d not found", id)
	}
	// TODO
	// if we cannot determine the job is in real final status or just in retrying stage
	// we cannot skip. Issue: https://github.com/goharbor/harbor/issues/9114
	if model.IsFinalStatus(task.Status) {
		log.Debugf("the task %d is in final status: %s, skip the stop action", id, task.Status)
		return nil
	}
	if err := cjob.GlobalClient.PostAction(task.JobID, string(job.StopCommand)); err != nil {
		e, ok := err.(*cjob.StatusBehindError)
		if ok {
			status := e.Status()
			task.Status = status
			task.StatusCode = job.Status(status).Code()
			task.EndTime = time.Now()
			if err := c.mgr.Update(task, "Status", "StatusCode", "EndTime"); err != nil {
				return err
			}
			log.Debugf("got status behind error for task %d, update it's status to %s directly", id, status)
			return nil
		}
		// the job records in Jobservice are swept periodically, will get not found error
		// when trying to stop the swept jobs. The logic here sets the status to stopped
		// directly when got this error
		if isJobNotFoundError(err) {
			status := job.StoppedStatus
			task.Status = status.String()
			task.StatusCode = status.Code()
			task.EndTime = time.Now()
			if err := c.mgr.Update(task, "Status", "StatusCode", "EndTime"); err != nil {
				return err
			}
			log.Debugf("got not found error for task %d, update it's status to %s directly", id, status.String())
			return nil
		}
		return err
	}
	log.Debugf("the stop request for task %d sent", id)
	return nil
}

func (c *controller) Delete(id int64) error {
	task, err := c.Get(id)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task %d not found", id)
	}
	if !model.IsFinalStatus(task.Status) {
		return fmt.Errorf("task %d is %s, isn't the final status, stop it first", id, task.Status)
	}
	return c.mgr.Delete(id)
}

func (c *controller) GetLog(id int64) ([]byte, error) {
	task, err := c.Get(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task %d not found", id)
	}
	return cjob.GlobalClient.GetJobLog(task.JobID)
}

func (c *controller) CalculateTaskGroupStatus(groupID int64) (*model.GroupStatus, error) {
	return c.mgr.CalculateTaskGroupStatus(groupID)
}

func isJobNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "object is not found")
}

func getInternalCoreURL() string {
	// the "RUN_MODE" is set to "TEST" when running unit test cases
	// to avoid the external dependency
	if os.Getenv("RUN_MODE") == "TEST" {
		return "http://127.0.0.1"
	}
	return config.InternalCoreURL()
}
