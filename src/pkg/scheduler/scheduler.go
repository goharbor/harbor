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

package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task"
	cronlib "github.com/robfig/cron"
)

var (
	// Sched is an instance of the default scheduler that can be used globally
	Sched = New()
)

// Schedule describes the detail information about the created schedule
type Schedule struct {
	ID           int64                  `json:"id"`
	VendorType   string                 `json:"vendor_type"`
	VendorID     int64                  `json:"vendor_id"`
	CRONType     string                 `json:"cron_type"`
	CRON         string                 `json:"cron"`
	ExtraAttrs   map[string]interface{} `json:"extra_attrs"`
	Status       string                 `json:"status"` // status of the underlying task(jobservice job)
	CreationTime time.Time              `json:"creation_time"`
	UpdateTime   time.Time              `json:"update_time"`
	// we can extend this model to include more information(e.g. how many times the schedule already
	// runs; when will the schedule runs next time)
}

// Scheduler provides the capability to run a periodic task, a callback function
// needs to be registered before using the scheduler
type Scheduler interface {
	// Schedule creates a task which calls the specified callback function periodically
	// The callback function needs to be registered first
	// The "vendorType" specifies the type of vendor (e.g. replication, scan, gc, retention, etc.),
	// and the "vendorID" specifies the ID of vendor if needed(e.g. policy ID for replication and retention).
	// The "callbackFuncParams" is passed to the callback function as encoded json string, so the callback
	// function must decode it before using
	// The customized attributes can be put into the "extraAttrs"
	Schedule(ctx context.Context, vendorType string, vendorID int64, cronType string,
		cron string, callbackFuncName string, callbackFuncParams interface{}, extraAttrs map[string]interface{}) (int64, error)
	// UnScheduleByID the schedule specified by ID
	UnScheduleByID(ctx context.Context, id int64) error
	// UnScheduleByVendor the schedule specified by vendor
	UnScheduleByVendor(ctx context.Context, vendorType string, vendorID int64) error
	// GetSchedule gets the schedule specified by ID
	GetSchedule(ctx context.Context, id int64) (*Schedule, error)
	// ListSchedules according to the query
	ListSchedules(ctx context.Context, query *q.Query) ([]*Schedule, error)
}

// New returns an instance of the default scheduler
func New() Scheduler {
	return &scheduler{
		dao:     &dao{},
		execMgr: task.ExecMgr,
		taskMgr: task.Mgr,
	}
}

type scheduler struct {
	dao     DAO
	execMgr task.ExecutionManager
	taskMgr task.Manager
}

func (s *scheduler) Schedule(ctx context.Context, vendorType string, vendorID int64, cronType string,
	cron string, callbackFuncName string, callbackFuncParams interface{}, extraAttrs map[string]interface{}) (int64, error) {
	if len(vendorType) == 0 {
		return 0, fmt.Errorf("empty vendor type")
	}
	if _, err := cronlib.Parse(cron); err != nil {
		return 0, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("invalid cron %s: %v", cron, err)
	}
	if !callbackFuncExist(callbackFuncName) {
		return 0, fmt.Errorf("callback function %s not found", callbackFuncName)
	}

	now := time.Now()
	sched := &schedule{
		VendorType:       vendorType,
		VendorID:         vendorID,
		CRONType:         cronType,
		CRON:             cron,
		CallbackFuncName: callbackFuncName,
		CreationTime:     now,
		UpdateTime:       now,
	}

	paramsData, err := json.Marshal(callbackFuncParams)
	if err != nil {
		return 0, err
	}
	sched.CallbackFuncParam = string(paramsData)
	params := map[string]interface{}{}
	if len(paramsData) > 0 {
		err = json.Unmarshal(paramsData, &params)
		if err != nil {
			log.Debugf("current paramsData is not a json string")
		}
	}
	extrasData, err := json.Marshal(extraAttrs)
	if err != nil {
		return 0, err
	}
	sched.ExtraAttrs = string(extrasData)

	// create schedule record
	// when checkin hook comes, the database record must exist,
	// so the database record must be created first before submitting job
	id, err := s.dao.Create(ctx, sched)
	if err != nil {
		return 0, err
	}

	execID, err := s.execMgr.Create(ctx, JobNameScheduler, id, task.ExecutionTriggerManual, params)
	if err != nil {
		return 0, err
	}

	taskID, err := s.taskMgr.Create(ctx, execID, &task.Job{
		Name: JobNameScheduler,
		Metadata: &job.Metadata{
			JobKind: job.KindPeriodic,
			Cron:    cron,
		},
	})
	if err != nil {
		return 0, err
	}
	// make sure the created task is stopped if got any error in the following steps
	defer func() {
		if err == nil {
			return
		}
		if err := s.taskMgr.Stop(ctx, taskID); err != nil {
			log.Errorf("failed to stop the task %d: %v", taskID, err)
		}
	}()
	// when task manager creating a task, it creates the task database record first and
	// then submits the job to jobservice. If the submitting failed, it doesn't return
	// any error. So we check the task status to make sure the job is submitted to jobservice
	// successfully here
	task, err := s.taskMgr.Get(ctx, taskID)
	if err != nil {
		return 0, err
	}
	if task.Status == job.ErrorStatus.String() {
		// assign the error to "err" to trigger the defer function to clean up the created task
		err = fmt.Errorf("failed to create the schedule: the task status is %s", job.ErrorStatus.String())
		return 0, err
	}

	return id, nil
}

func (s *scheduler) UnScheduleByID(ctx context.Context, id int64) error {
	executions, err := s.execMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": JobNameScheduler,
			"VendorID":   id,
		},
	})
	if err != nil {
		return err
	}
	if len(executions) > 0 {
		executionID := executions[0].ID
		// stop the execution
		if err = s.execMgr.StopAndWait(ctx, executionID, 10*time.Second); err != nil {
			return err
		}
		// delete execution
		if err = s.execMgr.Delete(ctx, executionID); err != nil {
			return err
		}
	}

	// delete schedule record
	return s.dao.Delete(ctx, id)
}

func (s *scheduler) UnScheduleByVendor(ctx context.Context, vendorType string, vendorID int64) error {
	q := &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": vendorType,
		},
	}
	if vendorID > 0 {
		q.Keywords["VendorID"] = vendorID
	}
	schedules, err := s.dao.List(ctx, q)
	if err != nil {
		return err
	}
	for _, schedule := range schedules {
		if err = s.UnScheduleByID(ctx, schedule.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *scheduler) GetSchedule(ctx context.Context, id int64) (*Schedule, error) {
	schedule, err := s.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.convertSchedule(ctx, schedule)
}

func (s *scheduler) ListSchedules(ctx context.Context, query *q.Query) ([]*Schedule, error) {
	schedules, err := s.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var scheds []*Schedule
	for _, schedule := range schedules {
		sched, err := s.convertSchedule(ctx, schedule)
		if err != nil {
			return nil, err
		}
		scheds = append(scheds, sched)
	}
	return scheds, nil
}

func (s *scheduler) convertSchedule(ctx context.Context, schedule *schedule) (*Schedule, error) {
	schd := &Schedule{
		ID:           schedule.ID,
		VendorType:   schedule.VendorType,
		VendorID:     schedule.VendorID,
		CRONType:     schedule.CRONType,
		CRON:         schedule.CRON,
		CreationTime: schedule.CreationTime,
		UpdateTime:   schedule.UpdateTime,
	}
	if len(schedule.ExtraAttrs) > 0 {
		extras := map[string]interface{}{}
		if err := json.Unmarshal([]byte(schedule.ExtraAttrs), &extras); err != nil {
			log.Errorf("failed to unmarshal the extra attributes of schedule %d: %v", schedule.ID, err)
			return nil, err
		}
		schd.ExtraAttrs = extras
	}

	executions, err := s.execMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": JobNameScheduler,
			"VendorID":   schedule.ID,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(executions) == 0 {
		// if no execution found for the schedule, mark it's status as error
		schd.Status = job.ErrorStatus.String()
	} else {
		schd.Status = executions[0].Status
	}
	return schd, nil
}
