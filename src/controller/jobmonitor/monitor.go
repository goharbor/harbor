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

package jobmonitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	jobSvc "github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/queuestatus"

	"github.com/goharbor/harbor/src/lib/log"

	"github.com/gocraft/work"

	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/lib/q"
	libRedis "github.com/goharbor/harbor/src/lib/redis"
	jm "github.com/goharbor/harbor/src/pkg/jobmonitor"
	"github.com/goharbor/harbor/src/pkg/task"
	taskDao "github.com/goharbor/harbor/src/pkg/task/dao"
)

const (
	all             = "all"
	batchUpdateSize = 1000
)

// Ctl the controller instance of the worker pool controller
var Ctl = NewMonitorController()

var skippedJobTypes = []string{
	"DEMO",
	"IMAGE_REPLICATE",
	"IMAGE_SCAN_ALL",
	"IMAGE_GC",
	"PURGE_AUDIT",
}

// MonitorController defines the worker pool operations
type MonitorController interface {
	// ListPools lists the worker pools
	ListPools(ctx context.Context) ([]*jm.WorkerPool, error)
	// ListWorkers lists the workers in the pool
	ListWorkers(ctx context.Context, poolID string) ([]*jm.Worker, error)

	// StopRunningJobs stop the running job
	StopRunningJobs(ctx context.Context, jobID string) error
	// StopPendingJobs stop the pending jobs
	StopPendingJobs(ctx context.Context, jobType string) error

	// ListQueues lists job queues
	ListQueues(ctx context.Context) ([]*jm.Queue, error)
	// PauseJobQueues suspend the job queue by type
	PauseJobQueues(ctx context.Context, jobType string) error
	// ResumeJobQueues resume the job queue by type
	ResumeJobQueues(ctx context.Context, jobType string) error
	GetJobLog(ctx context.Context, jobID string) ([]byte, error)
}

type monitorController struct {
	poolManager           jm.PoolManager
	workerManager         jm.WorkerManager
	taskManager           task.Manager
	queueManager          jm.QueueManager
	queueStatusManager    queuestatus.Manager
	monitorClient         func() (jm.JobServiceMonitorClient, error)
	jobServiceRedisClient func() (jm.RedisClient, error)
	executionDAO          taskDao.ExecutionDAO
}

// NewMonitorController ...
func NewMonitorController() MonitorController {
	return &monitorController{
		poolManager:           jm.NewPoolManager(),
		workerManager:         jm.NewWorkerManager(),
		taskManager:           task.NewManager(),
		queueManager:          jm.NewQueueClient(),
		queueStatusManager:    queuestatus.Mgr,
		monitorClient:         jobServiceMonitorClient,
		jobServiceRedisClient: jm.JobServiceRedisClient,
		executionDAO:          taskDao.NewExecutionDAO(),
	}
}

func jobServiceMonitorClient() (jm.JobServiceMonitorClient, error) {
	cfg, err := job.GlobalClient.GetJobServiceConfig()
	if err != nil {
		return nil, err
	}
	config := cfg.RedisPoolConfig
	pool, err := libRedis.GetRedisPool(jm.JobServicePool, config.RedisURL, &libRedis.PoolParam{
		PoolMaxIdle:     0,
		PoolIdleTimeout: time.Duration(config.IdleTimeoutSecond) * time.Second,
	})
	if err != nil {
		log.Errorf("failed to get redis pool: %v", err)
		return nil, err
	}
	return work.NewClient(fmt.Sprintf("{%s}", config.Namespace), pool), nil
}

func (w *monitorController) ListWorkers(ctx context.Context, poolID string) ([]*jm.Worker, error) {
	mClient, err := w.monitorClient()
	if err != nil {
		return nil, err
	}
	return w.workerManager.List(ctx, mClient, poolID)
}

func (w *monitorController) ListPools(ctx context.Context) ([]*jm.WorkerPool, error) {
	mClient, err := w.monitorClient()
	if err != nil {
		return nil, err
	}
	return w.poolManager.List(ctx, mClient)
}

func (w *monitorController) StopRunningJobs(ctx context.Context, jobID string) error {
	if !strings.EqualFold(jobID, all) {
		return w.stopJob(ctx, jobID)
	}
	allRunningJobs, err := w.allRunningJobs(ctx)
	if err != nil {
		log.Errorf("failed to get all running jobs: %v", err)
		return err
	}
	for _, jobID := range allRunningJobs {
		if err := w.stopJob(ctx, jobID); err != nil {
			log.Errorf("failed to stop running job %s: %v", jobID, err)
			return err
		}
	}
	return nil
}

func (w *monitorController) stopJob(ctx context.Context, jobID string) error {
	tasks, err := w.taskManager.List(ctx, &q.Query{Keywords: q.KeyWords{"job_id": jobID}})
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		// the job is not found
		log.Infof("job %s not found, maybe the job is already complete", jobID)
		return nil
	}
	if len(tasks) != 1 {
		return fmt.Errorf("there are more than one task with the same job ID")
	}
	// use local transaction to avoid rollback batch success tasks to previous state when one fail
	if ctx == nil {
		log.Debug("context is nil, skip stop operation")
		return nil
	}
	return orm.WithTransaction(func(ctx context.Context) error {
		return w.taskManager.Stop(ctx, tasks[0].ID)
	})(orm.SetTransactionOpNameToContext(ctx, "tx-stop-job"))
}

func (w *monitorController) allRunningJobs(ctx context.Context) ([]string, error) {
	jobIDs := make([]string, 0)
	wks, err := w.ListWorkers(ctx, all)
	if err != nil {
		log.Errorf("failed to list workers: %v", err)
		return nil, err
	}
	for _, wk := range wks {
		jobIDs = append(jobIDs, wk.JobID)
	}
	return jobIDs, nil
}

func (w *monitorController) StopPendingJobs(ctx context.Context, jobType string) error {
	redisClient, err := w.jobServiceRedisClient()
	if err != nil {
		return err
	}
	if !strings.EqualFold(jobType, all) {
		return w.stopPendingJob(ctx, jobType)
	}

	jobTypes, err := redisClient.AllJobTypes(ctx)
	if err != nil {
		return err
	}
	for _, jobType := range jobTypes {
		if err := w.stopPendingJob(ctx, jobType); err != nil {
			log.Warningf("failed to stop pending jobs of type %s: %v", jobType, err)
			continue
		}
	}
	return nil
}

func (w *monitorController) stopPendingJob(ctx context.Context, jobType string) error {
	redisClient, err := w.jobServiceRedisClient()
	if err != nil {
		return err
	}
	jobIDs, err := redisClient.StopPendingJobs(ctx, jobType)
	if err != nil {
		return err
	}
	go func() {
		if err = w.updateJobStatusInTask(orm.Context(), jobType, jobIDs, jobSvc.StoppedStatus.String()); err != nil {
			log.Errorf("failed to update job status in task: %v", err)
		}
	}()
	return nil
}

func (w *monitorController) updateJobStatusInTask(ctx context.Context, vendorType string, jobIDs []string, status string) error {
	if ctx == nil {
		log.Debug("context is nil, update job status in task")
		return nil
	}
	// Task count could be huge, to avoid query executionID by each task, query with vendor type and status
	// it might include extra executions, but it won't change these executions final status
	pendingExecs, err := w.taskManager.ExecutionIDsByVendorAndStatus(ctx, vendorType, jobSvc.PendingStatus.String())
	if err != nil {
		return err
	}
	if err := w.taskManager.UpdateStatusInBatch(ctx, jobIDs, status, batchUpdateSize); err != nil {
		log.Errorf("failed to update task status in batch: %v", err)
	}
	// Update execution status
	for _, executionID := range pendingExecs {
		if _, _, err := w.executionDAO.RefreshStatus(ctx, executionID); err != nil {
			log.Errorf("failed to refresh execution status: %v", err)
			continue
		}
	}
	return nil
}

func (w *monitorController) ListQueues(ctx context.Context) ([]*jm.Queue, error) {
	mClient, err := w.monitorClient()
	if err != nil {
		return nil, err
	}
	qs, err := mClient.Queues()
	if err != nil {
		return nil, err
	}
	// the original queue doesn't include the paused status, fetch it from the redis
	statusMap, err := w.queueStatusManager.AllJobTypeStatus(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*jm.Queue, 0)
	for _, queue := range qs {
		if skippedUnusedJobType(queue.JobName) {
			continue
		}
		result = append(result, &jm.Queue{
			JobType: queue.JobName,
			Count:   queue.Count,
			Latency: queue.Latency,
			Paused:  statusMap[queue.JobName],
		})
	}
	return result, nil
}

func skippedUnusedJobType(jobType string) bool {
	for _, t := range skippedJobTypes {
		if jobType == t {
			return true
		}
	}
	return false
}

func (w *monitorController) PauseJobQueues(ctx context.Context, jobType string) error {
	redisClient, err := w.jobServiceRedisClient()
	if err != nil {
		return err
	}
	if !strings.EqualFold(jobType, all) {
		return w.pauseQueue(ctx, jobType)
	}

	jobTypes, err := redisClient.AllJobTypes(ctx)
	if err != nil {
		return err
	}
	for _, t := range jobTypes {
		if err := w.pauseQueue(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (w *monitorController) pauseQueue(ctx context.Context, jobType string) error {
	if ctx == nil {
		log.Debug("context is nil, skip pause queue")
		return nil
	}
	redisClient, err := w.jobServiceRedisClient()
	if err != nil {
		return fmt.Errorf("failed to pause queue %v, error: %v", jobType, err)
	}
	err = redisClient.PauseJob(ctx, jobType)
	if err != nil {
		return fmt.Errorf("failed to pause queue %v, error: %v", jobType, err)
	}
	if err := orm.WithTransaction(func(ctx context.Context) error {
		return w.queueStatusManager.UpdateStatus(ctx, jobType, true)
	})(orm.SetTransactionOpNameToContext(ctx, "tx-update-queue-status")); err != nil {
		return fmt.Errorf("failed to pause queue %v, error: %v", jobType, err)
	}
	return nil
}

func (w *monitorController) ResumeJobQueues(ctx context.Context, jobType string) error {
	redisClient, err := w.jobServiceRedisClient()
	if err != nil {
		return err
	}
	if !strings.EqualFold(jobType, all) {
		return w.resumeQueue(ctx, jobType)
	}
	jobTypes, err := redisClient.AllJobTypes(ctx)
	if err != nil {
		return err
	}
	for _, jobType := range jobTypes {
		if err := w.resumeQueue(ctx, jobType); err != nil {
			return err
		}
	}
	return nil
}

func (w *monitorController) resumeQueue(ctx context.Context, jobType string) error {
	if ctx == nil {
		log.Debug("context is nil, skip resume queue")
		return nil
	}
	redisClient, err := w.jobServiceRedisClient()
	if err != nil {
		return fmt.Errorf("failed to resume queue %v, error: %v", jobType, err)
	}
	if err := redisClient.UnpauseJob(ctx, jobType); err != nil {
		return fmt.Errorf("failed to resume queue %v, error: %v", jobType, err)
	}
	if err := orm.WithTransaction(func(ctx context.Context) error {
		return w.queueStatusManager.UpdateStatus(ctx, jobType, false)
	})(orm.SetTransactionOpNameToContext(ctx, "tx-update-queue-status")); err != nil {
		return fmt.Errorf("failed to resume queue %v, error: %v", jobType, err)
	}
	return nil
}

func (w *monitorController) GetJobLog(ctx context.Context, jobID string) ([]byte, error) {
	return w.taskManager.GetLogByJobID(ctx, jobID)
}
