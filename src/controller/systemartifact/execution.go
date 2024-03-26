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

package systemartifact

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/systemartifact"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	cronTypeDaily = "Daily"
	cronSpec      = "0 0 0 * * *"
)

var (
	sched = scheduler.Sched
)

var Ctl = NewController()

type Controller interface {
	Start(ctx context.Context, async bool, trigger string) error
}

func NewController() Controller {
	return &controller{
		execMgr:           task.ExecMgr,
		taskMgr:           task.Mgr,
		systemArtifactMgr: systemartifact.Mgr,
		makeCtx:           orm.Context,
	}
}

type controller struct {
	execMgr           task.ExecutionManager
	taskMgr           task.Manager
	systemArtifactMgr systemartifact.Manager
	makeCtx           func() context.Context
}

func (c *controller) Start(ctx context.Context, async bool, trigger string) error {
	execID, err := c.execMgr.Create(ctx, job.SystemArtifactCleanupVendorType, 0, trigger)
	if err != nil {
		return err
	}
	// cleanup job would always be scheduled in async mode in production
	// allowing for sync mode execution only for test mode purposes
	// if there are any trigger settings then pass them to the cleanup manager first
	jobParams := job.Parameters{}

	if !async {
		err := c.createCleanupTask(ctx, jobParams, execID)
		if err != nil {
			log.Errorf("failed to create system artifact clean-up task: %v", err)
			return err
		}

		log.Info("Created job for scan data export successfully")
		return nil
	}
	go func(ctx context.Context) {
		err := retry.Retry(func() error {
			_, err := c.execMgr.Get(ctx, execID)
			return err
		})
		if err != nil {
			log.Errorf("failed to get the execution %d for the export data cleanup job", execID)
			return
		}
		err = c.createCleanupTask(ctx, jobParams, execID)
		if err != nil {
			log.Errorf("Encountered error in scan data artifact cleanup : %v", err)
			return
		}
	}(c.makeCtx())

	return nil
}

func (c *controller) createCleanupTask(ctx context.Context, jobParams job.Parameters, execID int64) error {
	j := &task.Job{
		Name: job.SystemArtifactCleanupVendorType,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: jobParams,
	}

	_, err := c.taskMgr.Create(ctx, execID, j)

	if err != nil {
		log.Errorf("Unable to create a scan data export job in clean-up mode : %v", err)
		c.markError(ctx, execID, err)
		return err
	}
	return nil
}

func (c *controller) markError(ctx context.Context, executionID int64, err error) {
	// try to stop the execution first in case that some tasks are already created
	if e := c.execMgr.StopAndWaitWithError(ctx, executionID, 10*time.Second, err); e != nil {
		log.Errorf("failed to stop the execution %d: %v", executionID, e)
	}
}

// ScheduleCleanupTask schedules a system artifact cleanup task
func ScheduleCleanupTask(ctx context.Context) error {
	return scheduleSystemArtifactCleanJob(ctx)
}

func scheduleSystemArtifactCleanJob(ctx context.Context) error {
	schedule, err := getSystemArtifactCleanupSchedule(ctx)
	if err != nil {
		return err
	}
	if schedule != nil {
		log.Debugf("Export data cleanup job already scheduled with ID : %v.", schedule.ID)
		return nil
	}
	scheduleID, err := sched.Schedule(ctx, job.SystemArtifactCleanupVendorType, 0, cronTypeDaily, cronSpec, SystemArtifactCleanupCallback, nil, nil)
	if err != nil {
		log.Errorf("Encountered error when scheduling scan data export cleanup job : %v", err)
		return err
	}
	log.Infof("Scheduled scan data export cleanup job with ID : %v", scheduleID)
	return nil
}

func getSystemArtifactCleanupSchedule(ctx context.Context) (*scheduler.Schedule, error) {
	query := q.New(map[string]interface{}{"vendor_type": job.SystemArtifactCleanupVendorType})
	schedules, err := sched.ListSchedules(ctx, query)
	if err != nil {
		log.Errorf("Unable to check if export data cleanup job is already scheduled : %v", err)
		return nil, err
	}
	if len(schedules) > 0 {
		log.Debugf("Found export data cleanup job with schedule id : %v", schedules[0].ID)
		return schedules[0], nil
	}
	return nil, nil
}
