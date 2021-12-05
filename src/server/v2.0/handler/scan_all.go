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

package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/scan_all"
)

func newScanAllAPI() *scanAllAPI {
	return &scanAllAPI{
		execMgr:    task.ExecMgr,
		scanCtl:    scan.DefaultController,
		scannerCtl: scanner.DefaultController,
		scheduler:  scheduler.Sched,
		makeCtx:    orm.Context,
	}
}

type scanAllAPI struct {
	BaseAPI
	execMgr    task.ExecutionManager
	scanCtl    scan.Controller
	scannerCtl scanner.Controller
	scheduler  scheduler.Scheduler
	makeCtx    func() context.Context
}

func (s *scanAllAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

// StopScanAll stops the execution of scan all artifacts.
func (s *scanAllAPI) StopScanAll(ctx context.Context, params operation.StopScanAllParams) middleware.Responder {
	if err := s.requireAccess(ctx, rbac.ActionStop); err != nil {
		return s.SendError(ctx, err)
	}

	execution, err := s.getLatestScanAllExecution(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}
	if execution == nil {
		message := fmt.Sprintf("no scan all job is found currently")
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage(message))
	}
	go func(ctx context.Context, eid int64) {
		err := s.execMgr.Stop(ctx, eid)
		if err != nil {
			log.Errorf("failed to stop the execution of executionID=%+v", execution.ID)
		}
	}(s.makeCtx(), execution.ID)

	return operation.NewStopScanAllAccepted()
}

func (s *scanAllAPI) CreateScanAllSchedule(ctx context.Context, params operation.CreateScanAllScheduleParams) middleware.Responder {
	if err := s.requireAccess(ctx, rbac.ActionCreate); err != nil {
		return s.SendError(ctx, err)
	}

	req := params.Schedule

	if req.Schedule.Type == ScheduleNone {
		return operation.NewCreateScanAllScheduleCreated()
	}

	if req.Schedule.Type == ScheduleManual {
		execution, err := s.getLatestScanAllExecution(ctx, task.ExecutionTriggerManual)
		if err != nil {
			return s.SendError(ctx, err)
		}

		if execution != nil && execution.IsOnGoing() {
			message := fmt.Sprintf("a previous scan all job aleady exits, its status is %s", execution.Status)
			return s.SendError(ctx, errors.ConflictError(nil).WithMessage(message))
		}

		if _, err := s.scanCtl.ScanAll(ctx, task.ExecutionTriggerManual, true); err != nil {
			return s.SendError(ctx, err)
		}
	} else {
		schedule, err := s.getScanAllSchedule(ctx)
		if err != nil {
			return s.SendError(ctx, err)
		}

		if schedule != nil {
			message := "fail to set schedule for scan all as always had one, please delete it firstly then to re-schedule"
			return s.SendError(ctx, errors.PreconditionFailedError(nil).WithMessage(message))
		}

		if _, err := s.createOrUpdateScanAllSchedule(ctx, req.Schedule.Type, req.Schedule.Cron, nil); err != nil {
			return s.SendError(ctx, err)
		}
	}

	return operation.NewCreateScanAllScheduleCreated()
}

func (s *scanAllAPI) UpdateScanAllSchedule(ctx context.Context, params operation.UpdateScanAllScheduleParams) middleware.Responder {
	if err := s.requireAccess(ctx, rbac.ActionUpdate); err != nil {
		return s.SendError(ctx, err)
	}
	req := params.Schedule

	if req.Schedule.Type == ScheduleManual {
		message := fmt.Sprintf("fail to update scan all schedule as wrong schedule type: %s", req.Schedule.Type)
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage(message))
	}

	schedule, err := s.getScanAllSchedule(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if req.Schedule.Type == ScheduleNone {
		if schedule != nil {
			err = s.scheduler.UnScheduleByID(ctx, schedule.ID)
		}
	} else {
		_, err = s.createOrUpdateScanAllSchedule(ctx, req.Schedule.Type, req.Schedule.Cron, schedule)
	}

	if err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewUpdateScanAllScheduleOK()
}

func (s *scanAllAPI) GetScanAllSchedule(ctx context.Context, params operation.GetScanAllScheduleParams) middleware.Responder {
	if err := s.requireAccess(ctx, rbac.ActionRead); err != nil {
		return s.SendError(ctx, err)
	}
	schedule, err := s.getScanAllSchedule(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewGetScanAllScheduleOK().WithPayload(model.NewSchedule(schedule).ToSwagger())
}

func (s *scanAllAPI) GetLatestScanAllMetrics(ctx context.Context, params operation.GetLatestScanAllMetricsParams) middleware.Responder {
	if err := s.requireAccess(ctx, rbac.ActionRead); err != nil {
		return s.SendError(ctx, err)
	}
	stats, err := s.getMetrics(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewGetLatestScanAllMetricsOK().WithPayload(stats)
}

func (s *scanAllAPI) GetLatestScheduledScanAllMetrics(ctx context.Context, params operation.GetLatestScheduledScanAllMetricsParams) middleware.Responder {
	if err := s.requireAccess(ctx, rbac.ActionRead); err != nil {
		return s.SendError(ctx, err)
	}
	stats, err := s.getMetrics(ctx, task.ExecutionTriggerSchedule)
	if err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewGetLatestScanAllMetricsOK().WithPayload(stats)
}

func (s *scanAllAPI) createOrUpdateScanAllSchedule(ctx context.Context, cronType, cron string, previous *scheduler.Schedule) (int64, error) {
	if previous != nil {
		if cronType == previous.CRONType && cron == previous.CRON {
			return previous.ID, nil
		}

		if err := s.scheduler.UnScheduleByID(ctx, previous.ID); err != nil {
			return 0, err
		}
	}

	return s.scheduler.Schedule(ctx, scan.VendorTypeScanAll, 0, cronType, cron, scan.ScanAllCallback, nil, nil)
}

func (s *scanAllAPI) getScanAllSchedule(ctx context.Context) (*scheduler.Schedule, error) {
	query := q.New(q.KeyWords{"vendor_type": scan.VendorTypeScanAll})
	schedules, err := s.scheduler.ListSchedules(ctx, query.First(q.NewSort("creation_time", true)))
	if err != nil {
		return nil, err
	}

	if len(schedules) > 1 {
		return nil, fmt.Errorf("found more than one scheduled scan all job, please ensure that only one schedule left")
	} else if len(schedules) == 0 {
		return nil, nil
	}

	return schedules[0], nil
}

func (s *scanAllAPI) getMetrics(ctx context.Context, trigger ...string) (*models.Stats, error) {
	execution, err := s.getLatestScanAllExecution(ctx, trigger...)
	if err != nil {
		return nil, err
	}

	sts := &models.Stats{}
	if execution != nil {
		if execution.Metrics != nil {
			metrics := execution.Metrics

			sts.Total = metrics.TaskCount
			sts.Completed = metrics.SuccessTaskCount + metrics.ErrorTaskCount + metrics.StoppedTaskCount
			sts.Metrics = map[string]int64{
				"Pending": metrics.PendingTaskCount,
				"Running": metrics.RunningTaskCount,
				"Success": metrics.SuccessTaskCount,
				"Error":   metrics.ErrorTaskCount,
				"Stopped": metrics.StoppedTaskCount,
			}
		} else {
			sts.Total = 0
			sts.Completed = 0
			sts.Metrics = map[string]int64{
				"Pending": 0,
				"Running": 0,
				"Success": 0,
				"Error":   0,
				"Stopped": 0,
			}
		}

		sts.Ongoing = !job.Status(execution.Status).Final() || sts.Total != sts.Completed
		sts.Trigger = strings.Title(strings.ToLower(execution.Trigger))
	}

	return sts, nil
}

func (s *scanAllAPI) getLatestScanAllExecution(ctx context.Context, trigger ...string) (*task.Execution, error) {
	query := q.New(q.KeyWords{"vendor_type": scan.VendorTypeScanAll})
	if len(trigger) > 0 {
		query.Keywords["trigger"] = trigger[0]
	}

	executions, err := s.execMgr.List(ctx, query.First(q.NewSort("start_time", true)))
	if err != nil {
		return nil, err
	}

	if len(executions) == 0 {
		return nil, nil
	}

	return executions[0], nil
}

func (s *scanAllAPI) requireScanEnabled(ctx context.Context) error {
	kws := make(map[string]interface{})
	kws["is_default"] = true

	query := &q.Query{
		Keywords: kws,
	}

	l, err := s.scannerCtl.ListRegistrations(ctx, query)
	if err != nil {
		return errors.Wrap(err, "check if scan is enabled")
	}

	if len(l) == 0 {
		return errors.PreconditionFailedError(nil).WithMessage("no scanner is configured, it's not possible to scan")
	}

	return nil
}

func (s *scanAllAPI) requireAccess(ctx context.Context, action rbac.Action) error {
	if err := s.RequireSystemAccess(ctx, action, rbac.ResourceScanAll); err != nil {
		return err
	}
	if err := s.requireScanEnabled(ctx); err != nil {
		return err
	}
	return nil
}
