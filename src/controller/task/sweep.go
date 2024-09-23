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

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

var (
	// SweepCtl is the global sweep controller
	SweepCtl = NewSweepController()
)

type SweepParams struct {
	// ExecRetainCounts records the retained execution counts for different vendor type
	ExecRetainCounts map[string]int64
}

const (
	// SchedulerCallback ...
	SchedulerCallback = "EXECUTION_SWEEP_CALLBACK"
	// systemVendorID represents the id for system job.
	systemVendorID = -1

	cronTypeCustom = "Custom"
	// run for every day
	cronSpec = "0 0 0 * * *"
)

func init() {
	err := scheduler.RegisterCallbackFunc(SchedulerCallback, sweepCallback)
	if err != nil {
		log.Fatalf("failed to register execution sweep job callback, error: %v", err)
	}
}

func sweepCallback(ctx context.Context, _ string) error {
	params := &SweepParams{ExecRetainCounts: job.GetExecutionSweeperCount()}
	return SweepCtl.Start(ctx, params, task.ExecutionTriggerSchedule)
}

type SweepController interface {
	Start(ctx context.Context, params *SweepParams, trigger string) error
}

type sweepController struct {
	execMgr task.ExecutionManager
	taskMgr task.Manager
}

func (sc *sweepController) Start(ctx context.Context, params *SweepParams, trigger string) error {
	jobParams := make(map[string]interface{})
	jobParams[task.ExecRetainCounts] = params.ExecRetainCounts

	execID, err := sc.execMgr.Create(ctx, job.ExecSweepVendorType, systemVendorID, trigger, jobParams)
	if err != nil {
		log.Errorf("failed to create execution for %s, error: %v", job.ExecSweepVendorType, err)
		return err
	}

	_, err = sc.taskMgr.Create(ctx, execID, &task.Job{
		Name: job.ExecSweepVendorType,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: jobParams,
	})
	if err != nil {
		log.Errorf("failed to create task for %s, error: %v", job.ExecSweepVendorType, err)
		return err
	}

	return nil
}

func NewSweepController() SweepController {
	return &sweepController{
		execMgr: task.ExecMgr,
		taskMgr: task.Mgr,
	}
}

// ScheduleSweepJob schedules the system execution sweep job.
func ScheduleSweepJob(ctx context.Context) error {
	sched, err := getScheduledSweepJob(ctx)
	if err != nil {
		return err
	}
	// unschedule the job if the cron changed
	if sched != nil {
		if sched.CRON != cronSpec {
			log.Debugf("reschedule the system execution job because the cron changed, old: %s, new: %s", sched.CRON, cronSpec)
			if err = scheduler.Sched.UnScheduleByID(ctx, sched.ID); err != nil {
				return err
			}
		} else {
			log.Debug("skip to schedule the system execution job because the old one existed and cron not changed")
			return nil
		}
	}

	// schedule a job if no schedule found or cron changed
	scheduleID, err := scheduler.Sched.Schedule(ctx, job.ExecSweepVendorType, systemVendorID, cronTypeCustom, cronSpec, SchedulerCallback, nil, nil)
	if err != nil {
		return err
	}

	log.Debugf("scheduled the system execution sweep job, id: %d", scheduleID)
	return nil
}

// getScheduledSweepJob gets sweep job which already scheduled.
func getScheduledSweepJob(ctx context.Context) (*scheduler.Schedule, error) {
	query := q.New(map[string]interface{}{"vendor_type": job.ExecSweepVendorType})
	schedules, err := scheduler.Sched.ListSchedules(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(schedules) > 0 {
		return schedules[0], nil
	}

	return nil, nil
}
