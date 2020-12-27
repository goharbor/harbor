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
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"time"
)

// go:generate mockery -name APIController -case snake

// APIController to handle the requests related with retention
type APIController interface {
	GetRetention(id int64) (*policy.Metadata, error)

	CreateRetention(p *policy.Metadata) (int64, error)

	UpdateRetention(p *policy.Metadata) error

	DeleteRetention(id int64) error

	TriggerRetentionExec(policyID int64, trigger string, dryRun bool) (int64, error)

	OperateRetentionExec(eid int64, action string) error

	GetRetentionExec(eid int64) (*Execution, error)

	ListRetentionExecs(policyID int64, query *q.Query) ([]*Execution, error)

	GetTotalOfRetentionExecs(policyID int64) (int64, error)

	ListRetentionExecTasks(executionID int64, query *q.Query) ([]*Task, error)

	GetTotalOfRetentionExecTasks(executionID int64) (int64, error)

	GetRetentionExecTaskLog(taskID int64) ([]byte, error)

	GetRetentionExecTask(taskID int64) (*Task, error)
}

// DefaultAPIController ...
type DefaultAPIController struct {
	manager        Manager
	execMgr        task.ExecutionManager
	taskMgr        task.Manager
	launcher       Launcher
	projectManager project.Manager
	repositoryMgr  repository.Manager
	scheduler      scheduler.Scheduler
}

const (
	// SchedulerCallback ...
	SchedulerCallback   = "RETENTION"
	schedulerVendorType = "RETENTION"
)

// TriggerParam ...
type TriggerParam struct {
	PolicyID int64
	Trigger  string
}

// GetRetention Get Retention
func (r *DefaultAPIController) GetRetention(id int64) (*policy.Metadata, error) {
	return r.manager.GetPolicy(id)
}

// CreateRetention Create Retention
func (r *DefaultAPIController) CreateRetention(p *policy.Metadata) (int64, error) {
	id, err := r.manager.CreatePolicy(p)
	if err != nil {
		return 0, err
	}

	if p.Trigger.Kind == policy.TriggerKindSchedule {
		cron, ok := p.Trigger.Settings[policy.TriggerSettingsCron]
		if ok && len(cron.(string)) > 0 {
			extras := make(map[string]interface{})
			if _, err = r.scheduler.Schedule(orm.Context(), schedulerVendorType, id, "", cron.(string), SchedulerCallback, TriggerParam{
				PolicyID: id,
				Trigger:  ExecutionTriggerSchedule,
			}, extras); err != nil {
				return 0, err
			}
		}
	}

	return id, nil
}

// UpdateRetention Update Retention
func (r *DefaultAPIController) UpdateRetention(p *policy.Metadata) error {
	p0, err := r.manager.GetPolicy(p.ID)
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
	if err = r.manager.UpdatePolicy(p); err != nil {
		return err
	}
	if needUn {
		err = r.scheduler.UnScheduleByVendor(orm.Context(), schedulerVendorType, p.ID)
		if err != nil {
			return err
		}
	}
	if needSch {
		extras := make(map[string]interface{})
		_, err := r.scheduler.Schedule(orm.Context(), schedulerVendorType, p.ID, "", p.Trigger.Settings[policy.TriggerSettingsCron].(string), SchedulerCallback, TriggerParam{
			PolicyID: p.ID,
			Trigger:  ExecutionTriggerSchedule,
		}, extras)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteRetention Delete Retention
func (r *DefaultAPIController) DeleteRetention(id int64) error {
	p, err := r.manager.GetPolicy(id)
	if err != nil {
		return err
	}
	if p.Trigger.Kind == policy.TriggerKindSchedule && len(p.Trigger.Settings[policy.TriggerSettingsCron].(string)) > 0 {
		err = r.scheduler.UnScheduleByVendor(orm.Context(), schedulerVendorType, id)
		if err != nil {
			return err
		}
	}

	ctx := orm.Context()
	err = r.deleteExecs(ctx, id)
	if err != nil {
		return err
	}
	return r.manager.DeletePolicyAndExec(id)
}

// deleteExecs delete executions
func (r *DefaultAPIController) deleteExecs(ctx context.Context, vendorID int64) error {
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
func (r *DefaultAPIController) TriggerRetentionExec(policyID int64, trigger string, dryRun bool) (int64, error) {
	p, err := r.manager.GetPolicy(policyID)
	if err != nil {
		return 0, err
	}

	ctx := orm.Context()
	id, err := r.execMgr.Create(ctx, job.Retention, policyID, trigger,
		map[string]interface{}{
			"dry_run": dryRun,
		},
	)
	if _, err = r.launcher.Launch(p, id, dryRun); err != nil {
		if err1 := r.execMgr.StopAndWait(ctx, id, 10*time.Second); err1 != nil {
			logger.Errorf("failed to stop the retention execution %d: %v", id, err1)
		}
		if err1 := r.execMgr.MarkError(ctx, id, err.Error()); err1 != nil {
			logger.Errorf("failed to mark error for the retention execution %d: %v", id, err1)
		}
		return 0, err
	}
	return id, err

}

// OperateRetentionExec Operate Retention Execution
func (r *DefaultAPIController) OperateRetentionExec(eid int64, action string) error {
	ctx := orm.Context()
	e, err := r.execMgr.Get(ctx, eid)
	if err != nil {
		return err
	}
	if e == nil {
		return fmt.Errorf("execution %d not found", eid)
	}
	switch action {
	case "stop":
		return r.launcher.Stop(eid)
	default:
		return fmt.Errorf("not support action %s", action)
	}
}

// GetRetentionExec Get Retention Execution
func (r *DefaultAPIController) GetRetentionExec(executionID int64) (*Execution, error) {
	ctx := orm.Context()
	e, err := r.execMgr.Get(ctx, executionID)
	if err != nil {
		return nil, err
	}

	return convertExecution(e), nil
}

// ListRetentionExecs List Retention Executions
func (r *DefaultAPIController) ListRetentionExecs(policyID int64, query *q.Query) ([]*Execution, error) {
	ctx := orm.Context()
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.Retention
	query.Keywords["VendorID"] = policyID
	execs, err := r.execMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var executions []*Execution
	for _, exec := range execs {
		executions = append(executions, convertExecution(exec))
	}
	return executions, nil
}

func convertExecution(exec *task.Execution) *Execution {
	return &Execution{
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
func (r *DefaultAPIController) GetTotalOfRetentionExecs(policyID int64) (int64, error) {
	ctx := orm.Context()
	return r.execMgr.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": job.Retention,
			"VendorID":   policyID,
		},
	})
}

// ListRetentionExecTasks List Retention Execution Histories
func (r *DefaultAPIController) ListRetentionExecTasks(executionID int64, query *q.Query) ([]*Task, error) {
	ctx := orm.Context()
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.Retention
	query.Keywords["ExecutionID"] = executionID
	tks, err := r.taskMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var tasks []*Task
	for _, tk := range tks {
		tasks = append(tasks, convertTask(tk))
	}
	return tasks, nil
}

func convertTask(t *task.Task) *Task {
	return &Task{
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
func (r *DefaultAPIController) GetTotalOfRetentionExecTasks(executionID int64) (int64, error) {
	ctx := orm.Context()
	return r.taskMgr.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"VendorType":  job.Retention,
			"ExecutionID": executionID,
		},
	})
}

// GetRetentionExecTaskLog Get Retention Execution Task Log
func (r *DefaultAPIController) GetRetentionExecTaskLog(taskID int64) ([]byte, error) {
	ctx := orm.Context()
	return r.taskMgr.GetLog(ctx, taskID)
}

// GetRetentionExecTask Get Retention Execution Task
func (r *DefaultAPIController) GetRetentionExecTask(taskID int64) (*Task, error) {
	ctx := orm.Context()
	t, err := r.taskMgr.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return convertTask(t), nil
}

// UpdateTaskInfo Update task info
func (r *DefaultAPIController) UpdateTaskInfo(taskID int64, total int, retained int) error {
	ctx := orm.Context()
	t, err := r.taskMgr.Get(ctx, taskID)
	if err != nil {
		return err
	}

	t.ExtraAttrs["total"] = total
	t.ExtraAttrs["retained"] = retained

	return r.taskMgr.UpdateExtraAttrs(ctx, taskID, t.ExtraAttrs)
}

// NewAPIController ...
func NewAPIController(retentionMgr Manager, projectManager project.Manager, repositoryMgr repository.Manager, scheduler scheduler.Scheduler, retentionLauncher Launcher, execMgr task.ExecutionManager, taskMgr task.Manager) APIController {
	return &DefaultAPIController{
		manager:        retentionMgr,
		execMgr:        execMgr,
		taskMgr:        taskMgr,
		launcher:       retentionLauncher,
		projectManager: projectManager,
		repositoryMgr:  repositoryMgr,
		scheduler:      scheduler,
	}
}
