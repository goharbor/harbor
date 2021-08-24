package scandataexport

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	q2 "github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	"github.com/goharbor/harbor/src/pkg/task"
	"time"
)

const (
	DigestKey                   = "artifact_digest"
	VendorTypeExportDataCleanup = "EXPORT_DATA_CLEANUP"
)

func init() {
	task.SetExecutionSweeperCount(job.ScanDataExport, 50)
}

var Ctl = NewController()

type Controller interface {
	Start(ctx context.Context, criteria export.Criteria) (executionId int64, err error)
	GetExecution(ctx context.Context, executionId int64) (*export.Execution, error)
	GetTask(ctx context.Context, executionId int64) (*task.Task, error)
	StartCleanup(ctx context.Context, trigger string, params TriggerParam, async bool) error
}

func NewController() Controller {
	return &controller{
		execMgr:    task.ExecMgr,
		taskMgr:    task.Mgr,
		cleanupMgr: export.CleanupManager,
	}
}

type controller struct {
	execMgr    task.ExecutionManager
	taskMgr    task.Manager
	exportMgr  export.Manager
	cleanupMgr export.ArtifactCleanupManager
}

func (c *controller) GetTask(ctx context.Context, executionId int64) (*task.Task, error) {
	query := new(q2.Query)

	keywords := make(map[string]interface{})
	keywords["VendorType"] = job.ScanDataExport
	keywords["ExecutionID"] = executionId
	query.Keywords = keywords
	query.Sorts = append(query.Sorts, &q2.Sort{
		Key:  "ID",
		DESC: true,
	})
	tasks, err := c.taskMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.Errorf("No task found for execution Id : %d", executionId)
	}
	// for the export JOB there would be a single instance of the task corresponding to the execution
	// we will hence return the latest instance of the task associated with this execution
	logger.Infof("Returning task instance with ID : %v", tasks[0].ID)
	return tasks[0], nil
}

func (c *controller) GetExecution(ctx context.Context, executionId int64) (*export.Execution, error) {
	exec, err := c.execMgr.Get(ctx, executionId)
	if err != nil {
		logger.Errorf("Error when fetching execution status for ExecutionId: %d error : %v", executionId, err)
		return nil, err
	}

	execStatus := &export.Execution{
		ID:            exec.ID,
		UserID:        exec.VendorID,
		Status:        exec.Status,
		StatusMessage: exec.StatusMessage,
		Trigger:       exec.Trigger,
		StartTime:     exec.StartTime,
		EndTime:       exec.EndTime,
	}
	if digest, ok := exec.ExtraAttrs[DigestKey]; ok {
		execStatus.ExportDataDigest = digest.(string)
	}
	return execStatus, nil
}

func (c *controller) Start(ctx context.Context, criteria export.Criteria) (executionId int64, err error) {
	logger := log.GetLogger(ctx)
	vendorId := int64(ctx.Value("vendorId").(int))
	id, err := c.execMgr.Create(ctx, job.ScanDataExport, vendorId, task.ExecutionTriggerManual)
	logger.Infof("Created an execution record with id : %d for vendorId: %d", id, vendorId)
	if err != nil {
		logger.Errorf("Encountered error when creating job : %v", err)
		return 0, err
	}

	// create a job object and fill with metadata and parameters
	params := make(map[string]interface{})
	params["JobId"] = id
	params["Criteria"] = criteria
	j := &task.Job{
		Name: job.ScanDataExport,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
	}

	_, err = c.taskMgr.Create(ctx, id, j)

	if err != nil {
		logger.Errorf("Unable to create a scan data export job: %v", err)
		c.markError(ctx, id, err)
		return 0, err
	}

	logger.Info("Created job for scan data export successfully")
	return id, nil
}

func (c *controller) StartCleanup(ctx context.Context, trigger string, params TriggerParam, async bool) error {
	execId, err := c.execMgr.Create(ctx, VendorTypeExportDataCleanup, 0, trigger)
	if err != nil {
		return err
	}
	// cleanup job would always be scheduled in async mode in production
	// allowing for sync mode execution only for test mode purposes
	// if there are any trigger settings then pass them to the cleanup manager first
	if (params != TriggerParam{}) {
		settings := export.NewCleanupSettings()
		if params.TimeWindowMinutes > 0 {
			settings.Set(export.TimeWindowKey, params.TimeWindowMinutes)
		}
		if params.PageSize > 0 {
			settings.Set(export.PageSizeKey, params.PageSize)
		}

		c.cleanupMgr.Configure(settings)
	}
	if !async {
		err := c.cleanupMgr.Execute(ctx)
		if err != nil {
			return err
		}
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
		err = c.cleanupMgr.Execute(ctx)
		if err != nil {
			logger.Errorf("Encountered error in scan data artifact cleanup : %v", err)
			return
		}
	}(ctx)

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
