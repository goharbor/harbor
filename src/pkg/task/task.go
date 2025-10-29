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
	"context"
	"encoding/json"
	"fmt"
	"time"

	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

var (
	// Mgr is a global task manager instance
	Mgr = NewManager()
)

// Manager manages tasks.
// The execution and task managers provide an execution-task model to abstract the interactive with jobservice.
// All of the operations with jobservice should be delegated by them
type Manager interface {
	// Create submits the job to jobservice and creates a corresponding task record.
	// An execution must be created first and the task will be linked to it.
	// The "extraAttrs" can be used to set the customized attributes
	Create(ctx context.Context, executionID int64, job *Job, extraAttrs ...map[string]any) (id int64, err error)
	// Stop the specified task
	Stop(ctx context.Context, id int64) (err error)
	// Get the specified task
	Get(ctx context.Context, id int64) (task *Task, err error)
	// List the tasks according to the query
	// Query the "ExtraAttrs" by setting 'query.Keywords["ExtraAttrs.key"]="value"'
	List(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// Update the extra attributes of the specified task
	UpdateExtraAttrs(ctx context.Context, id int64, extraAttrs map[string]any) (err error)
	// Get the log of the specified task
	GetLog(ctx context.Context, id int64) (log []byte, err error)
	// GetLogByJobID get the log of specified job id
	GetLogByJobID(ctx context.Context, jobID string) (log []byte, err error)
	// Count counts total of tasks according to the query.
	// Query the "ExtraAttrs" by setting 'query.Keywords["ExtraAttrs.key"]="value"'
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Update the status of the specified task
	Update(ctx context.Context, task *Task, props ...string) error
	// UpdateStatusInBatch updates the status of tasks in batch
	UpdateStatusInBatch(ctx context.Context, jobIDs []string, status string, batchSize int) error
	// ExecutionIDsByVendorAndStatus retrieve execution id by vendor type and status
	ExecutionIDsByVendorAndStatus(ctx context.Context, vendorType, status string) ([]int64, error)
	// ListScanTasksByReportUUID lists scan tasks by report uuid, although it's a specific case but it will be
	// more suitable to support multi database in the future.
	ListScanTasksByReportUUID(ctx context.Context, uuid string) (tasks []*Task, err error)
	// RetrieveStatusFromTask retrieve status from task
	RetrieveStatusFromTask(ctx context.Context, reportID string) string
	// IsTaskFinished checks if the scan task is finished by report UUID
	IsTaskFinished(ctx context.Context, reportID string) bool
}

// NewManager creates an instance of the default task manager
func NewManager() Manager {
	return &manager{
		dao:      dao.NewTaskDAO(),
		execDAO:  dao.NewExecutionDAO(),
		jsClient: cjob.GlobalClient,
		coreURL:  config.GetCoreURL(),
	}
}

type manager struct {
	dao      dao.TaskDAO
	execDAO  dao.ExecutionDAO
	jsClient cjob.Client
	coreURL  string
}

func (m *manager) Update(ctx context.Context, task *Task, props ...string) error {
	return m.dao.Update(ctx, &dao.Task{
		ID:     task.ID,
		Status: task.Status,
	}, props...)
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) Create(ctx context.Context, executionID int64, jb *Job, extraAttrs ...map[string]any) (int64, error) {
	// create task record in database
	id, err := m.createTaskRecord(ctx, executionID, extraAttrs...)
	if err != nil {
		return 0, err
	}
	log.Debugf("the database record for task %d created", id)

	// submit job to jobservice
	// As all database operations are in a transaction which is committed until API returns,
	// when the job is submitted to the jobservice and running, the task record may not
	// insert yet, this will cause the status hook handler returning 404, and the jobservice
	// will re-send the status hook again
	jobID, err := m.submitJob(ctx, id, jb)
	if err != nil {
		// failed to submit job to jobservice, delete the task record
		log.Errorf("delete task %d from db due to failed to submit job %v, error: %v", id, jb.Name, err)
		if err := m.dao.Delete(ctx, id); err != nil {
			log.Errorf("failed to delete the task %d: %v", id, err)
		}
		return 0, err
	}

	log.Debugf("the task %d is submitted to jobservice, the job ID is %s", id, jobID)

	// populate the job ID for the task
	if err = m.dao.Update(ctx, &dao.Task{
		ID:    id,
		JobID: jobID,
	}, "JobID"); err != nil {
		log.Errorf("failed to populate the job ID for the task %d: %v", id, err)
	}

	return id, nil
}

func (m *manager) createTaskRecord(ctx context.Context, executionID int64, extraAttrs ...map[string]any) (int64, error) {
	exec, err := m.execDAO.Get(ctx, executionID)
	if err != nil {
		return 0, err
	}
	extras := map[string]any{}
	if len(extraAttrs) > 0 && extraAttrs[0] != nil {
		extras = extraAttrs[0]
	}
	data, err := json.Marshal(extras)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	return m.dao.Create(ctx, &dao.Task{
		VendorType:   exec.VendorType,
		ExecutionID:  executionID,
		Status:       job.PendingStatus.String(),
		StatusCode:   job.PendingStatus.Code(),
		ExtraAttrs:   string(data),
		CreationTime: now,
		UpdateTime:   now,
	})
}

func (m *manager) submitJob(_ context.Context, id int64, jb *Job) (string, error) {
	jobData := &models.JobData{
		Name:       jb.Name,
		StatusHook: fmt.Sprintf("%s/service/notifications/tasks/%d", m.coreURL, id),
	}
	if jb.Parameters != nil {
		jobData.Parameters = models.Parameters(jb.Parameters)
	}
	if jb.Metadata != nil {
		jobData.Metadata = &models.JobMetadata{
			JobKind:       jb.Metadata.JobKind,
			ScheduleDelay: jb.Metadata.ScheduleDelay,
			Cron:          jb.Metadata.Cron,
			IsUnique:      jb.Metadata.IsUnique,
		}
	}

	return m.jsClient.SubmitJob(jobData)
}

func (m *manager) Stop(ctx context.Context, id int64) error {
	task, err := m.dao.Get(ctx, id)
	if err != nil {
		return err
	}

	// when a task is in final status, if it's a periodic or retrying job it will
	// run again in the near future, so we must operate the stop action to these final
	// status jobs as well
	if err = m.jsClient.PostAction(task.JobID, string(job.StopCommand)); err != nil {
		// job not found, update it's status to stop directly
		if err == cjob.ErrJobNotFound {
			now := time.Now()
			err = m.dao.Update(ctx, &dao.Task{
				ID:         task.ID,
				Status:     job.StoppedStatus.String(),
				StatusCode: job.StoppedStatus.Code(),
				UpdateTime: now,
				EndTime:    now,
			}, "Status", "StatusCode", "UpdateTime", "EndTime")
			if err != nil {
				return err
			}
			log.Debugf("got job not found error for task %d, update it's status to stop directly", task.ID)
			// as in this case no status hook will be sent, here refresh the execution status directly
			_, _, err = m.execDAO.RefreshStatus(ctx, task.ExecutionID)
			return err
		}
		return err
	}
	log.Debugf("the stop request for task %d is sent", id)
	return nil
}

func (m *manager) Get(ctx context.Context, id int64) (*Task, error) {
	task, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	t := &Task{}
	t.From(task)
	return t, nil
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*Task, error) {
	tasks, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var ts []*Task
	for _, task := range tasks {
		t := &Task{}
		t.From(task)
		ts = append(ts, t)
	}
	return ts, nil
}

func (m *manager) ListScanTasksByReportUUID(ctx context.Context, uuid string) ([]*Task, error) {
	tasks, err := m.dao.ListScanTasksByReportUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	var ts []*Task
	for _, task := range tasks {
		t := &Task{}
		t.From(task)
		ts = append(ts, t)
	}
	return ts, nil
}

func (m *manager) UpdateExtraAttrs(ctx context.Context, id int64, extraAttrs map[string]any) error {
	data, err := json.Marshal(extraAttrs)
	if err != nil {
		return err
	}
	return m.dao.Update(ctx, &dao.Task{
		ID:         id,
		ExtraAttrs: string(data),
		UpdateTime: time.Time{},
	}, "ExtraAttrs", "UpdateTime")
}

func (m *manager) GetLog(ctx context.Context, id int64) ([]byte, error) {
	task, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return m.jsClient.GetJobLog(task.JobID)
}

func (m *manager) UpdateStatusInBatch(ctx context.Context, jobIDs []string, status string, batchSize int) error {
	return m.dao.UpdateStatusInBatch(ctx, jobIDs, status, batchSize)
}

func (m *manager) ExecutionIDsByVendorAndStatus(ctx context.Context, vendorType, status string) ([]int64, error) {
	return m.dao.ExecutionIDsByVendorAndStatus(ctx, vendorType, status)
}

func (m *manager) GetLogByJobID(_ context.Context, jobID string) (log []byte, err error) {
	return m.jsClient.GetJobLog(jobID)
}

func (m *manager) RetrieveStatusFromTask(ctx context.Context, reportID string) string {
	if len(reportID) == 0 {
		return ""
	}
	tasks, err := m.dao.ListScanTasksByReportUUID(ctx, reportID)
	if err != nil {
		log.Warningf("can not find the task with report UUID %v, error %v", reportID, err)
		return ""
	}
	if len(tasks) > 0 {
		return tasks[0].Status
	}
	return ""
}

func (m *manager) IsTaskFinished(ctx context.Context, reportID string) bool {
	status := m.RetrieveStatusFromTask(ctx, reportID)
	if len(status) == 0 {
		return true
	}
	return job.Status(status).Final()
}
