package systemartifact

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/systemartifact"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	VendorTypeSystemArtifactCleanup = "SYSTEM_ARTIFACT_CLEANUP"
	cronTypeDaily                   = "Daily"
	cronSpec                        = "0 0 0 * * *"
)

var (
	sched = scheduler.Sched
)

func init() {
	task.SetExecutionSweeperCount(VendorTypeSystemArtifactCleanup, 50)
}

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
	execID, err := c.execMgr.Create(ctx, VendorTypeSystemArtifactCleanup, 0, trigger)
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

		logger.Info("Created job for scan data export successfully")
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
			logger.Errorf("Encountered error in scan data artifact cleanup : %v", err)
			return
		}
	}(c.makeCtx())

	return nil
}

func (c *controller) createCleanupTask(ctx context.Context, jobParams job.Parameters, execID int64) error {
	j := &task.Job{
		Name: job.SystemArtifactCleanup,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: jobParams,
	}

	_, err := c.taskMgr.Create(ctx, execID, j)

	if err != nil {
		logger.Errorf("Unable to create a scan data export job in clean-up mode : %v", err)
		c.markError(ctx, execID, err)
		return err
	}
	return nil
}

func (c *controller) markError(ctx context.Context, executionID int64, err error) {
	// try to stop the execution first in case that some tasks are already created
	if err := c.execMgr.StopAndWait(ctx, executionID, 10*time.Second); err != nil {
		logger.Errorf("failed to stop the execution %d: %v", executionID, err)
	}
	if err := c.execMgr.MarkError(ctx, executionID, err.Error()); err != nil {
		logger.Errorf("failed to mark error for the execution %d: %v", executionID, err)
	}
}

// ScheduleCleanupTask schedules a system artifact cleanup task
func ScheduleCleanupTask(ctx context.Context) {
	scheduleSystemArtifactCleanJob(ctx)
}

func scheduleSystemArtifactCleanJob(ctx context.Context) {
	schedule, err := getSystemArtifactCleanupSchedule(ctx)
	if err != nil {
		return
	}
	if schedule != nil {
		logger.Debugf(" Export data cleanup job already scheduled with ID : %v.", schedule.ID)
		return
	}
	scheduleID, err := sched.Schedule(ctx, VendorTypeSystemArtifactCleanup, 0, cronTypeDaily, cronSpec, SystemArtifactCleanupCallback, nil, nil)
	if err != nil {
		log.Errorf("Encountered error when scheduling scan data export cleanup job : %v", err)
		return
	}
	log.Infof("Scheduled scan data export cleanup job with ID : %v", scheduleID)
}

func getSystemArtifactCleanupSchedule(ctx context.Context) (*scheduler.Schedule, error) {
	query := q.New(map[string]interface{}{"vendor_type": VendorTypeSystemArtifactCleanup})
	schedules, err := sched.ListSchedules(ctx, query)
	if err != nil {
		logger.Errorf("Unable to check if export data cleanup job is already scheduled : %v", err)
		return nil, err
	}
	if len(schedules) > 0 {
		logger.Infof("Found export data cleanup job with schedule id : %v", schedules[0].ID)
		return schedules[0], nil
	}
	return nil, nil
}
