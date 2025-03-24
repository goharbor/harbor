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
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/jobservice"
	pg "github.com/goharbor/harbor/src/controller/purge"
	"github.com/goharbor/harbor/src/controller/task"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	taskPkg "github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/purge"
)

type purgeAPI struct {
	BaseAPI
	purgeCtr     pg.Controller
	schedulerCtl jobservice.SchedulerController
	taskCtl      task.Controller
	executionCtl task.ExecutionController
}

func newPurgeAPI() *purgeAPI {
	return &purgeAPI{
		purgeCtr:     pg.Ctrl,
		schedulerCtl: jobservice.SchedulerCtl,
		taskCtl:      task.Ctl,
		executionCtl: task.ExecutionCtl,
	}
}

func (p *purgeAPI) CreatePurgeSchedule(ctx context.Context, params purge.CreatePurgeScheduleParams) middleware.Responder {
	if err := p.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourcePurgeAuditLog); err != nil {
		return p.SendError(ctx, err)
	}
	if err := verifyCreateRequest(params); err != nil {
		return p.SendError(ctx, err)
	}
	id, err := p.kick(ctx, job.PurgeAuditVendorType, params.Schedule.Schedule.Type, params.Schedule.Schedule.Cron, params.Schedule.Parameters)
	if err != nil {
		return p.SendError(ctx, err)
	}
	location := path.Join(params.HTTPRequest.URL.Path, fmt.Sprintf("../%d", id))
	return purge.NewCreatePurgeScheduleCreated().WithLocation(location)
}

func verifyCreateRequest(params purge.CreatePurgeScheduleParams) error {
	if params.Schedule == nil || params.Schedule.Schedule == nil {
		return errors.BadRequestError(fmt.Errorf("schedule cann't be empty"))
	}
	if len(params.Schedule.Parameters) == 0 {
		return errors.BadRequestError(fmt.Errorf("schedule parameter cann't be empty"))
	}
	if _, exist := params.Schedule.Parameters[common.PurgeAuditRetentionHour]; !exist {
		return errors.BadRequestError(fmt.Errorf("audit_retention_hour should provide"))
	}
	if _, err := retentionHour(params.Schedule.Parameters); err != nil {
		return err
	}
	if _, exist := params.Schedule.Parameters[common.PurgeAuditIncludeEventTypes]; !exist {
		return errors.BadRequestError(fmt.Errorf("include_event_types should provide"))
	}
	return nil
}

func retentionHour(m map[string]interface{}) (int, error) {
	if ret, ok := m[common.PurgeAuditRetentionHour]; ok {
		if rh, ok := ret.(json.Number); ok {
			ret, err := rh.Int64()
			if err != nil {
				return 0, errors.BadRequestError(fmt.Errorf("audit_retention_hour should be integer format"))
			}
			if int(ret) > common.MaxAuditRetentionHour {
				return 0, errors.BadRequestError(fmt.Errorf("audit_retention_hour should be less than %d", common.MaxAuditRetentionHour))
			}
			return int(ret), nil
		}
	}
	return 0, nil
}

func (p *purgeAPI) kick(ctx context.Context, vendorType string, scheType string, cron string, parameters map[string]interface{}) (int64, error) {
	if parameters == nil {
		parameters = make(map[string]interface{})
	}
	var err error
	var id int64

	policy := pg.JobPolicy{
		ExtraAttrs: parameters,
	}
	if dryRun, ok := parameters[common.PurgeAuditDryRun].(bool); ok {
		policy.DryRun = dryRun
	}
	if includeEventTypes, ok := parameters[common.PurgeAuditIncludeEventTypes].(string); ok {
		policy.IncludeEventTypes = includeEventTypes
	}
	retHour, err := retentionHour(parameters)
	if err != nil {
		return 0, err
	}
	policy.RetentionHour = retHour

	switch scheType {
	case ScheduleManual:
		id, err = p.purgeCtr.Start(ctx, policy, taskPkg.ExecutionTriggerManual)
	case ScheduleNone:
		// delete the schedule of purge
		err = p.schedulerCtl.Delete(ctx, vendorType)
	case ScheduleHourly, ScheduleDaily, ScheduleWeekly, ScheduleCustom:
		err = p.updateSchedule(ctx, vendorType, scheType, cron, policy, parameters)
	}
	return id, err
}

func (p *purgeAPI) updateSchedule(ctx context.Context, vendorType, cronType, cron string, policy pg.JobPolicy, extraParams map[string]interface{}) error {
	if err := utils.ValidateCronString(cron); err != nil {
		return errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("invalid cron string for scheduled log rotation purge: %s, error: %v", cron, err)
	}
	if err := p.schedulerCtl.Delete(ctx, vendorType); err != nil {
		return err
	}
	return p.createSchedule(ctx, vendorType, cronType, cron, policy, extraParams)
}

func (p *purgeAPI) GetPurgeHistory(ctx context.Context, params purge.GetPurgeHistoryParams) middleware.Responder {
	if err := p.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourcePurgeAuditLog); err != nil {
		return p.SendError(ctx, err)
	}
	query, err := p.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	query.Keywords["VendorType"] = job.PurgeAuditVendorType
	if err != nil {
		return p.SendError(ctx, err)
	}
	total, err := p.executionCtl.Count(ctx, query)
	if err != nil {
		return p.SendError(ctx, err)
	}
	execs, err := p.executionCtl.List(ctx, query)
	if err != nil {
		p.SendError(ctx, err)
	}

	var hs []*model.ExecHistory
	for _, exec := range execs {
		extraAttrsString, err := json.Marshal(exec.ExtraAttrs)
		if err != nil {
			return p.SendError(ctx, err)
		}
		hs = append(hs, &model.ExecHistory{
			ID:         exec.ID,
			Name:       job.PurgeAuditVendorType,
			Kind:       exec.Trigger,
			Parameters: string(extraAttrsString),
			Schedule: &model.ScheduleParam{
				Type: exec.Trigger,
			},
			Status:       exec.Status,
			CreationTime: exec.StartTime,
			UpdateTime:   exec.UpdateTime,
		})
	}
	var results []*models.ExecHistory
	for _, h := range hs {
		results = append(results, h.ToSwagger())
	}

	return purge.NewGetPurgeHistoryOK().
		WithXTotalCount(total).
		WithLink(p.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (p *purgeAPI) GetPurgeJob(ctx context.Context, params purge.GetPurgeJobParams) middleware.Responder {
	if err := p.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourcePurgeAuditLog); err != nil {
		return p.SendError(ctx, err)
	}

	exec, err := p.executionCtl.Get(ctx, params.PurgeID)
	if exec.VendorType != job.PurgeAuditVendorType {
		return p.SendError(ctx, fmt.Errorf("purge job with id %d not found", params.PurgeID))
	}
	if err != nil {
		return p.SendError(ctx, err)
	}

	extraAttrsString, err := json.Marshal(exec.ExtraAttrs)
	if err != nil {
		return p.SendError(ctx, err)
	}

	res := &model.ExecHistory{
		ID:         exec.ID,
		Name:       job.PurgeAuditVendorType,
		Kind:       exec.Trigger,
		Parameters: string(extraAttrsString),
		Status:     exec.Status,
		Schedule: &model.ScheduleParam{
			Type: exec.Trigger,
		},
		CreationTime: exec.StartTime,
		UpdateTime:   exec.UpdateTime,
	}

	return purge.NewGetPurgeJobOK().WithPayload(res.ToSwagger())
}

func (p *purgeAPI) GetPurgeJobLog(ctx context.Context, params purge.GetPurgeJobLogParams) middleware.Responder {
	if err := p.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourcePurgeAuditLog); err != nil {
		return p.SendError(ctx, err)
	}
	tasks, err := p.taskCtl.List(ctx, q.New(q.KeyWords{
		"ExecutionID": params.PurgeID,
		"VendorType":  job.PurgeAuditVendorType,
	}))
	if err != nil {
		return p.SendError(ctx, err)
	}
	if len(tasks) == 0 {
		return p.SendError(ctx,
			errors.New(nil).WithCode(errors.NotFoundCode).
				WithMessagef("purge job with execution ID: %d taskLog is not found", params.PurgeID))
	}
	taskLog, err := p.taskCtl.GetLog(ctx, tasks[0].ID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	return purge.NewGetPurgeJobLogOK().WithPayload(string(taskLog))
}

func (p *purgeAPI) GetPurgeSchedule(ctx context.Context, _ purge.GetPurgeScheduleParams) middleware.Responder {
	if err := p.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourcePurgeAuditLog); err != nil {
		return p.SendError(ctx, err)
	}
	sch, err := p.schedulerCtl.Get(ctx, job.PurgeAuditVendorType)
	if errors.IsNotFoundErr(err) {
		return purge.NewGetPurgeScheduleOK()
	}
	if err != nil {
		return p.SendError(ctx, err)
	}
	execHistory := &models.ExecHistory{
		ID:            sch.ID,
		JobName:       "",
		JobKind:       sch.CRON,
		JobParameters: pg.String(sch.ExtraAttrs),
		Deleted:       false,
		JobStatus:     sch.Status,
		Schedule: &models.ScheduleObj{
			Cron:              sch.CRON,
			Type:              sch.CRONType,
			NextScheduledTime: strfmt.DateTime(utils.NextSchedule(sch.CRON, time.Now())),
		},
		CreationTime: strfmt.DateTime(sch.CreationTime),
		UpdateTime:   strfmt.DateTime(sch.UpdateTime),
	}
	return purge.NewGetPurgeScheduleOK().WithPayload(execHistory)
}

func (p *purgeAPI) UpdatePurgeSchedule(ctx context.Context, params purge.UpdatePurgeScheduleParams) middleware.Responder {
	if err := p.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourcePurgeAuditLog); err != nil {
		return p.SendError(ctx, err)
	}
	if err := verifyUpdateRequest(params); err != nil {
		return p.SendError(ctx, err)
	}
	_, err := p.kick(ctx, job.PurgeAuditVendorType, params.Schedule.Schedule.Type, params.Schedule.Schedule.Cron, params.Schedule.Parameters)
	if err != nil {
		return p.SendError(ctx, err)
	}
	return purge.NewUpdatePurgeScheduleOK()
}

func verifyUpdateRequest(params purge.UpdatePurgeScheduleParams) error {
	if params.Schedule == nil || params.Schedule.Schedule == nil {
		return errors.BadRequestError(fmt.Errorf("schedule cann't be empty"))
	}
	if len(params.Schedule.Parameters) == 0 {
		return errors.BadRequestError(fmt.Errorf("schedule parameter cann't be empty"))
	}
	if _, exist := params.Schedule.Parameters[common.PurgeAuditRetentionHour]; !exist {
		return errors.BadRequestError(fmt.Errorf("audit_retention_hour should provide"))
	}
	if _, err := retentionHour(params.Schedule.Parameters); err != nil {
		return err
	}
	if _, exist := params.Schedule.Parameters[common.PurgeAuditIncludeEventTypes]; !exist {
		return errors.BadRequestError(fmt.Errorf("include_event_types should provide"))
	}
	return nil
}

func (p *purgeAPI) createSchedule(ctx context.Context, vendorType string, cronType string, cron string, policy pg.JobPolicy, extraParam map[string]interface{}) error {
	_, err := p.schedulerCtl.Create(ctx, vendorType, cronType, cron, pg.SchedulerCallback, policy, extraParam)
	if err != nil {
		return err
	}
	return nil
}

func (p *purgeAPI) StopPurge(ctx context.Context, params purge.StopPurgeParams) middleware.Responder {
	if err := p.RequireSystemAccess(ctx, rbac.ActionStop, rbac.ResourcePurgeAuditLog); err != nil {
		return p.SendError(ctx, err)
	}
	if err := p.purgeCtr.Stop(ctx, params.PurgeID); err != nil {
		return p.SendError(ctx, err)
	}
	return purge.NewStopPurgeOK()
}
