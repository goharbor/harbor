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

package sync

import (
	"context"
	"fmt"
	"time"

	o "github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/mgt"
	"github.com/goharbor/harbor/src/jobservice/period"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

// PolicyLoader is a func template to load schedule policies from js datastore.
type PolicyLoader func() ([]*period.Policy, error)

// Worker is designed to sync the schedules in the database into jobservice datastore.
type Worker struct {
	// context of Worker.
	context *env.Context
	// Indicate whether a new round should be run.
	lastErr error
	// How many rounds have been run.
	round uint8
	// The max number of rounds for repeated runs.
	maxRounds uint8
	// Periodical job scheduler.
	scheduler period.Scheduler
	// Job stats manager.
	manager mgt.Manager
	// Internal addr of core.
	internalCoreAddr string
	// Scheduler from core.
	coreScheduler scheduler.Scheduler
	// Execution manager from core.
	coreExecutionManager task.ExecutionManager
	// Task manager from core
	coreTaskManager task.Manager
	// Loader for loading polices from the js store.
	policyLoader PolicyLoader
}

// New sync worker.
func New(maxRounds uint8) *Worker {
	return &Worker{
		maxRounds: maxRounds,
	}
}

// WithContext set context.
func (w *Worker) WithContext(ctx *env.Context) *Worker {
	w.context = ctx
	return w
}

// UseScheduler refers the period.Scheduler.
func (w *Worker) UseScheduler(scheduler period.Scheduler) *Worker {
	w.scheduler = scheduler
	return w
}

// WithCoreInternalAddr sets the internal addr of core.
func (w *Worker) WithCoreInternalAddr(addr string) *Worker {
	w.internalCoreAddr = addr
	return w
}

// UseManager refers the mgt.Manager.
func (w *Worker) UseManager(mgr mgt.Manager) *Worker {
	w.manager = mgr
	return w
}

// UseCoreScheduler refers the core scheduler.
func (w *Worker) UseCoreScheduler(scheduler scheduler.Scheduler) *Worker {
	w.coreScheduler = scheduler
	return w
}

// UseCoreExecutionManager refers the core execution manager.
func (w *Worker) UseCoreExecutionManager(executionMgr task.ExecutionManager) *Worker {
	w.coreExecutionManager = executionMgr
	return w
}

// UseCoreTaskManager refers the core task manager.
func (w *Worker) UseCoreTaskManager(taskManager task.Manager) *Worker {
	w.coreTaskManager = taskManager
	return w
}

// WithPolicyLoader determines the policy loader func.
func (w *Worker) WithPolicyLoader(loader PolicyLoader) *Worker {
	w.policyLoader = loader
	return w
}

// Start the loop in none-blocking way.
func (w *Worker) Start() error {
	if err := w.validate(); err != nil {
		return err
	}

	w.context.WG.Add(1)
	// Run
	go func() {
		defer func() {
			w.context.WG.Done()
		}()

		ctx := orm.NewContext(w.context.SystemContext, o.NewOrm())
		ctlChan := make(chan struct{}, 1)
		ctlChan <- struct{}{}

		for {
			select {
			case <-ctlChan:
				w.round++
				if w.round == w.maxRounds {
					return
				}

				if err := w.Run(ctx); err == nil {
					return
				}

				// Wait for a while and retry then.
				time.AfterFunc(1*time.Minute, func() {
					ctlChan <- struct{}{}
				})
			case <-w.context.SystemContext.Done():
				logger.Info("Context cancel signal received:sync worker exit")
				return
			}
		}
	}()

	return nil
}

// Run one round.
func (w *Worker) Run(ctx context.Context) error {
	// Start sync schedules.
	logger.Infof("Start to sync schedules in database to jobservice: round[%d].", w.round)

	// Get all the schedules from the database first.
	// Use the default scheduler.
	schedules, err := w.coreScheduler.ListSchedules(ctx, &q.Query{})
	if err != nil {
		// We can not proceed.
		// A non-nil error will cause a follow-up retry later.
		return errors.Wrap(err, "list all the schedules in the database")
	}

	// Exit earlier if no schedules found.
	if len(schedules) == 0 {
		// Log and gracefully exit.
		logger.Info("No schedules found in the database.")
		return nil
	}

	// Get schedule records from the jobservice datastore.
	polices, err := w.policyLoader()
	if err != nil {
		return errors.Wrap(err, "load schedule records from jobservice store")
	}

	// Define a function to get the policy with the specified ID.
	getPolicy := func(jobID string) *period.Policy {
		for _, p := range polices {
			if p.ID == jobID {
				return p
			}
		}

		return nil
	}

	jobHash := make(map[string]struct{})
	restoreCounter := 0
	clearCounter := 0

	// Sync now.
	for _, sch := range schedules {
		// Get the corresponding task.
		t, err := w.getTask(ctx, sch)
		if err != nil {
			// Log and skip
			logger.Error(err)
			w.lastErr = err

			continue
		}

		// Recorded
		jobHash[t.JobID] = struct{}{}

		// Get policy
		p := getPolicy(t.JobID)
		if p == nil {
			// Need to restore this missing schedule.
			if err := w.restore(sch.CRON, t); err != nil {
				// Log and skip
				logger.Error(err)
				w.lastErr = err
			} else {
				restoreCounter++
				logger.Infof("Sync: restore missing schedule: taskID=%d, jobID=%s, cron=%s", t.ID, t.JobID, sch.CRON)
			}
		}
	}

	// Clear the dirty ones.
	for _, p := range polices {
		_, ok := jobHash[p.ID]
		if p.JobName == scheduler.JobNameScheduler && !ok {
			if err := w.scheduler.UnSchedule(p.ID); err != nil {
				logger.Error(err)
			} else {
				clearCounter++
				logger.Infof("Sync: unschedule dirty schedule: %s:%s", p.JobName, p.ID)
			}
		}
	}

	logger.Infof("End sync schedules in database to jobservice: round[%d].", w.round)
	logger.Infof("Found %d schedules, restore %d missing schedules, clear %d dirty schedules", len(schedules), restoreCounter, clearCounter)

	return w.lastErr
}

func (w *Worker) restore(cron string, t *task.Task) error {
	p := &period.Policy{
		ID:         t.JobID,
		JobName:    scheduler.JobNameScheduler,
		CronSpec:   cron,
		WebHookURL: fmt.Sprintf("%s/service/notifications/tasks/%d", w.internalCoreAddr, t.ID),
	}

	// Schedule the policy.
	numericID, err := w.scheduler.Schedule(p)
	if err != nil {
		return errors.Wrap(err, "schedule policy")
	}

	res := &job.Stats{
		Info: &job.StatsInfo{
			JobID:       p.ID,
			JobName:     p.JobName,
			Status:      job.ScheduledStatus.String(),
			JobKind:     job.KindPeriodic,
			CronSpec:    cron,
			WebHookURL:  p.WebHookURL,
			NumericPID:  numericID,
			EnqueueTime: time.Now().Unix(),
			UpdateTime:  time.Now().Unix(),
			RefLink:     fmt.Sprintf("/api/v1/jobs/%s", p.ID),
			Revision:    t.StatusRevision,
		},
	}

	// Keep status synced.
	if res.Info.Revision > 0 {
		res.Info.HookAck = &job.ACK{
			Revision: t.StatusRevision,
			Status:   job.ScheduledStatus.String(),
		}
	}

	// Save the stats.
	return w.manager.SaveJob(res)
}

// validate whether the worker is ready or not.
func (w *Worker) validate() error {
	if w.context == nil {
		return errors.New("missing context")
	}

	if w.manager == nil {
		return errors.New("missing stats manager")
	}

	if w.scheduler == nil {
		return errors.New("missing period scheduler")
	}

	if len(w.internalCoreAddr) == 0 {
		return errors.New("empty internal addr of core")
	}

	if w.coreScheduler == nil {
		return errors.New("missing core scheduler")
	}

	if w.coreExecutionManager == nil {
		return errors.New("missing core execution manager")
	}

	if w.coreTaskManager == nil {
		return errors.New("missing core task manager")
	}

	if w.policyLoader == nil {
		return errors.New("missing policy loader")
	}

	return nil
}

// getTask gets the task associated with the specified schedule.
// Here is an assumption that each schedule has only one execution as well as only one task under the execution.
func (w *Worker) getTask(ctx context.Context, schedule *scheduler.Schedule) (*task.Task, error) {
	// Get associated execution first.
	executions, err := w.coreExecutionManager.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"vendor_type": scheduler.JobNameScheduler,
			"vendor_id":   schedule.ID,
		},
	})

	if err != nil {
		return nil, err
	}

	if len(executions) == 0 {
		return nil, errors.Errorf("no execution found for schedule: %s:%d", schedule.VendorType, schedule.VendorID)
	}

	theOne := executions[0]
	// Now get the execution.
	tasks, err := w.coreTaskManager.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"execution_id": theOne.ID,
		},
	})

	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, errors.Errorf("no task found for execution: %s:%d", schedule.VendorType, theOne.ID)
	}

	return tasks[0], nil
}
