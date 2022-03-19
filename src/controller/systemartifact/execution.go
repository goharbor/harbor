package systemartifact

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/systemartifact"
	"github.com/goharbor/harbor/src/pkg/task"
	"time"
)

const (
	VendorTypeSystemArtifactCleanup = "SYSTEM_ARTIFACT_CLEANUP"
)

func init() {
	task.SetExecutionSweeperCount(job.SystemArtifactCleanup, 50)
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
	execId, err := c.execMgr.Create(ctx, VendorTypeSystemArtifactCleanup, 0, trigger)
	if err != nil {
		return err
	}
	// cleanup job would always be scheduled in async mode in production
	// allowing for sync mode execution only for test mode purposes
	// if there are any trigger settings then pass them to the cleanup manager first
	jobParams := job.Parameters{}

	if !async {
		err := c.createCleanupTask(ctx, jobParams, execId)
		if err != nil {
			return err
		}

		logger.Info("Created job for scan data export successfully")
		return nil
	}
	go func(ctx context.Context) {
		err := retry.Retry(func() error {
			_, err := c.execMgr.Get(ctx, execId)
			return err
		})
		if err != nil {
			log.Errorf("failed to get the execution %d for the export data cleanup job", execId)
			return
		}
		err = c.createCleanupTask(ctx, jobParams, execId)
		if err != nil {
			logger.Errorf("Encountered error in scan data artifact cleanup : %v", err)
			return
		}
	}(c.makeCtx())

	return nil
}

func (c *controller) createCleanupTask(ctx context.Context, jobParams job.Parameters, execId int64) error {
	j := &task.Job{
		Name: job.SystemArtifactCleanup,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: jobParams,
	}

	_, err := c.taskMgr.Create(ctx, execId, j)

	if err != nil {
		logger.Errorf("Unable to create a scan data export job in clean-up mode : %v", err)
		c.markError(ctx, execId, err)
		return err
	}
	return nil
}

func (c *controller) markError(ctx context.Context, executionID int64, err error) {
	logger := log.GetLogger(ctx)
	// try to stop the execution first in case that some tasks are already created
	if err := c.execMgr.StopAndWait(ctx, executionID, 10*time.Second); err != nil {
		logger.Errorf("failed to stop the execution %d: %v", executionID, err)
	}
	if err := c.execMgr.MarkError(ctx, executionID, err.Error()); err != nil {
		logger.Errorf("failed to mark error for the execution %d: %v", executionID, err)
	}
}
