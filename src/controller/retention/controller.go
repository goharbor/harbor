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

package retention

import (
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

func init() {
	// keep only the latest created 50 retention execution records
	task.SetExecutionSweeperCount(job.Retention, 50)
}

// go:generate mockery -name Controller -case snake

// Controller to handle the requests related with retention
type Controller interface {
	GetRetention(ctx context.Context, id int64) (*policy.Metadata, error)

	CreateRetention(ctx context.Context, p *policy.Metadata) (int64, error)

	UpdateRetention(ctx context.Context, p *policy.Metadata) error

	DeleteRetention(ctx context.Context, id int64) error

	TriggerRetentionExec(ctx context.Context, policyID int64, trigger string, dryRun bool) (int64, error)

	OperateRetentionExec(ctx context.Context, eid int64, action string) error

	GetRetentionExec(ctx context.Context, eid int64) (*retention.Execution, error)

	ListRetentionExecs(ctx context.Context, policyID int64, query *q.Query) ([]*retention.Execution, error)

	GetTotalOfRetentionExecs(ctx context.Context, policyID int64) (int64, error)

	ListRetentionExecTasks(ctx context.Context, executionID int64, query *q.Query) ([]*retention.Task, error)

	GetTotalOfRetentionExecTasks(ctx context.Context, executionID int64) (int64, error)

	GetRetentionExecTaskLog(ctx context.Context, taskID int64) ([]byte, error)

	GetRetentionExecTask(ctx context.Context, taskID int64) (*retention.Task, error)
}

var (
	// Ctl is a global retention controller instance
	Ctl = NewController()
)

// defaultController ...
type defaultController struct {
	manager        retention.Manager
	execMgr        task.ExecutionManager
	taskMgr        task.Manager
	launcher       retention.Launcher
	projectManager project.Manager
	repositoryMgr  repository.Manager
	scheduler      scheduler.Scheduler
}

const (
	// SchedulerCallback ...
	SchedulerCallback   = "RETENTION"
	schedulerVendorType = job.Retention
)

// TriggerParam ...
type TriggerParam struct {
	PolicyID int64
	Trigger  string
}

// GetRetention Get Retention
func (r *defaultController) GetRetention(ctx context.Context, id int64) (*policy.Metadata, error) {
	return r.manager.GetPolicy(ctx, id)
}

// CreateRetention Create Retention
func (r *defaultController) CreateRetention(ctx context.Context, p *policy.Metadata) (int64, error) {
	id, err := r.manager.CreatePolicy(ctx, p)
	if err != nil {
		return 0, err
	}

	if p.Trigger.Kind == policy.TriggerKindSchedule {
		cron, ok := p.Trigger.Settings[policy.TriggerSettingsCron]
		if ok && len(cron.(string)) > 0 {
			extras := make(map[string]interface{})
			if _, err = r.scheduler.Schedule(ctx, schedulerVendorType, id, "", cron.(string), SchedulerCallback, TriggerParam{
				PolicyID: id,
				Trigger:  retention.ExecutionTriggerSchedule,
			}, extras); err != nil {
				return 0, err
			}
		}
	}

	return id, nil
}

// UpdateRetention Update Retention
func (r *defaultController) UpdateRetention(ctx context.Context, p *policy.Metadata) error {
	p0, err := r.manager.GetPolicy(ctx, p.ID)
	if err != nil {
		return err
	}
	needUn := false
	needSch := false

	if p0.Trigger.Kind != p.Trigger.Kind {
		if p0.Trigger.Kind == policy.TriggerKindSchedule {
			needUn = true
		}

		if p.Trigger.Kind == policy.TriggerKindSchedule {
			needSch = true
		}
	} else {
		switch p.Trigger.Kind {
		case policy.TriggerKindSchedule:
			if p0.Trigger.Settings["cron"] != p.Trigger.Settings["cron"] {
				// unschedule old
				if len(p0.Trigger.Settings[policy.TriggerSettingsCron].(string)) > 0 {
					needUn = true
				}
				// schedule new
				if len(p.Trigger.Settings[policy.TriggerSettingsCron].(string)) > 0 {
					// valid cron
					needSch = true
				}
			}
		case "":

		default:
			return fmt.Errorf("not support Trigger %s", p.Trigger.Kind)
		}
	}
	if err = r.manager.UpdatePolicy(ctx, p); err != nil {
		return err
	}
	if needUn {
		err = r.scheduler.UnScheduleByVendor(ctx, schedulerVendorType, p.ID)
		if err != nil {
			return err
		}
	}
	if needSch {
		extras := make(map[string]interface{})
		_, err := r.scheduler.Schedule(ctx, schedulerVendorType, p.ID, "", p.Trigger.Settings[policy.TriggerSettingsCron].(string), SchedulerCallback, TriggerParam{
			PolicyID: p.ID,
			Trigger:  retention.ExecutionTriggerSchedule,
		}, extras)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteRetention Delete Retention
func (r *defaultController) DeleteRetention(ctx context.Context, id int64) error {
	p, err := r.manager.GetPolicy(ctx, id)
	if err != nil {
		return err
	}
	if p.Trigger.Kind == policy.TriggerKindSchedule && len(p.Trigger.Settings[policy.TriggerSettingsCron].(string)) > 0 {
		err = r.scheduler.UnScheduleByVendor(ctx, schedulerVendorType, id)
		if err != nil {
			return err
		}
	}

	err = r.deleteExecs(ctx, id)
	if err != nil {
		return err
	}
	return r.manager.DeletePolicy(ctx, id)
}

// deleteExecs delete executions
func (r *defaultController) deleteExecs(ctx context.Context, vendorID int64) error {
	executions, err := r.execMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": job.Retention,
			"VendorID":   vendorID,
		},
	})

	if err != nil {
		return err
	}

	for _, execution := range executions {
		if err = r.execMgr.Delete(ctx, execution.ID); err != nil {
			return err
		}
	}

	return nil
}

// TriggerRetentionExec Trigger Retention Execution
func (r *defaultController) TriggerRetentionExec(ctx context.Context, policyID int64, trigger string, dryRun bool) (int64, error) {
	p, err := r.manager.GetPolicy(ctx, policyID)
	if err != nil {
		return 0, err
	}

	id, err := r.execMgr.Create(ctx, job.Retention, policyID, trigger,
		map[string]interface{}{
			"dry_run": dryRun,
		},
	)
	if num, err := r.launcher.Launch(ctx, p, id, dryRun); err != nil {
		if err1 := r.execMgr.StopAndWait(ctx, id, 10*time.Second); err1 != nil {
			logger.Errorf("failed to stop the retention execution %d: %v", id, err1)
		}
		if err1 := r.execMgr.MarkError(ctx, id, err.Error()); err1 != nil {
			logger.Errorf("failed to mark error for the retention execution %d: %v", id, err1)
		}
		return 0, err
	} else if num == 0 {
		// no candidates, mark the execution as done directly
		if err := r.execMgr.MarkDone(ctx, id, "no resources for retention"); err != nil {
			logger.Errorf("failed to mark done for the execution %d: %v", id, err)
		}
	}
	return id, err
}

// OperateRetentionExec Operate Retention Execution
func (r *defaultController) OperateRetentionExec(ctx context.Context, eid int64, action string) error {
	e, err := r.execMgr.Get(ctx, eid)
	if err != nil {
		return err
	}
	if e == nil {
		return fmt.Errorf("execution %d not found", eid)
	}
	switch action {
	case "stop":
		return r.launcher.Stop(ctx, eid)
	default:
		return fmt.Errorf("not support action %s", action)
	}
}

// GetRetentionExec Get Retention Execution
func (r *defaultController) GetRetentionExec(ctx context.Context, executionID int64) (*retention.Execution, error) {
	e, err := r.execMgr.Get(ctx, executionID)
	if err != nil {
		return nil, err
	}

	return convertExecution(e), nil
}

// ListRetentionExecs List Retention Executions
func (r *defaultController) ListRetentionExecs(ctx context.Context, policyID int64, query *q.Query) ([]*retention.Execution, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.Retention
	query.Keywords["VendorID"] = policyID
	execs, err := r.execMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var executions []*retention.Execution
	for _, exec := range execs {
		executions = append(executions, convertExecution(exec))
	}
	return executions, nil
}

func convertExecution(exec *task.Execution) *retention.Execution {
	return &retention.Execution{
		ID:        exec.ID,
		PolicyID:  exec.VendorID,
		StartTime: exec.StartTime,
		EndTime:   exec.EndTime,
		Status:    exec.Status,
		Trigger:   exec.Trigger,
		DryRun:    exec.ExtraAttrs["dry_run"].(bool),
	}
}

// GetTotalOfRetentionExecs Count Retention Executions
func (r *defaultController) GetTotalOfRetentionExecs(ctx context.Context, policyID int64) (int64, error) {
	return r.execMgr.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": job.Retention,
			"VendorID":   policyID,
		},
	})
}

// ListRetentionExecTasks List Retention Execution Histories
func (r *defaultController) ListRetentionExecTasks(ctx context.Context, executionID int64, query *q.Query) ([]*retention.Task, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.Retention
	query.Keywords["ExecutionID"] = executionID
	tks, err := r.taskMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var tasks []*retention.Task
	for _, tk := range tks {
		tasks = append(tasks, convertTask(tk))
	}
	return tasks, nil
}

func convertTask(t *task.Task) *retention.Task {
	return &retention.Task{
		ID:          t.ID,
		ExecutionID: t.ExecutionID,
		Repository:  t.GetStringFromExtraAttrs("repository"),
		JobID:       t.JobID,
		Status:      t.Status,
		StatusCode:  job.Status(t.Status).Code(),
		StartTime:   t.StartTime,
		EndTime:     t.EndTime,
		Total:       int(t.GetNumFromExtraAttrs("total")),
		Retained:    int(t.GetNumFromExtraAttrs("retained")),
	}
}

// GetTotalOfRetentionExecTasks Count Retention Execution Histories
func (r *defaultController) GetTotalOfRetentionExecTasks(ctx context.Context, executionID int64) (int64, error) {
	return r.taskMgr.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType":  job.Retention,
			"ExecutionID": executionID,
		},
	})
}

// GetRetentionExecTaskLog Get Retention Execution Task Log
func (r *defaultController) GetRetentionExecTaskLog(ctx context.Context, taskID int64) ([]byte, error) {
	return r.taskMgr.GetLog(ctx, taskID)
}

// GetRetentionExecTask Get Retention Execution Task
func (r *defaultController) GetRetentionExecTask(ctx context.Context, taskID int64) (*retention.Task, error) {
	t, err := r.taskMgr.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return convertTask(t), nil
}

// UpdateTaskInfo Update task info
func (r *defaultController) UpdateTaskInfo(ctx context.Context, taskID int64, total int, retained int) error {
	t, err := r.taskMgr.Get(ctx, taskID)
	if err != nil {
		return err
	}

	t.ExtraAttrs["total"] = total
	t.ExtraAttrs["retained"] = retained

	return r.taskMgr.UpdateExtraAttrs(ctx, taskID, t.ExtraAttrs)
}

// NewController ...
func NewController() Controller {
	retentionMgr := retention.NewManager()
	retentionLauncher := retention.NewLauncher(pkg.ProjectMgr, pkg.RepositoryMgr, retentionMgr, task.ExecMgr, task.Mgr)
	return &defaultController{
		manager:        retentionMgr,
		execMgr:        task.ExecMgr,
		taskMgr:        task.Mgr,
		launcher:       retentionLauncher,
		projectManager: pkg.ProjectMgr,
		repositoryMgr:  pkg.RepositoryMgr,
		scheduler:      scheduler.Sched,
	}
}
