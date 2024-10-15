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
	"os"
	"strings"

	"github.com/go-openapi/runtime/middleware"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/gc"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/gc"
)

type gcAPI struct {
	BaseAPI
	gcCtr gc.Controller
}

func newGCAPI() *gcAPI {
	return &gcAPI{
		gcCtr: gc.NewController(),
	}
}

func (g *gcAPI) Prepare(_ context.Context, _ string, _ interface{}) middleware.Responder {
	return nil
}

func (g *gcAPI) CreateGCSchedule(ctx context.Context, params operation.CreateGCScheduleParams) middleware.Responder {
	if err := g.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceGarbageCollection); err != nil {
		return g.SendError(ctx, err)
	}
	id, err := g.kick(ctx, params.Schedule.Schedule.Type, params.Schedule.Schedule.Cron, params.Schedule.Parameters)
	if err != nil {
		return g.SendError(ctx, err)
	}
	// replace the /api/v2.0/system/gc/schedule/{id} to /api/v2.0/system/gc/{id}
	lastSlashIndex := strings.LastIndex(params.HTTPRequest.URL.Path, "/")
	if lastSlashIndex != -1 {
		location := fmt.Sprintf("%s/%d", params.HTTPRequest.URL.Path[:lastSlashIndex], id)
		return operation.NewCreateGCScheduleCreated().WithLocation(location)
	}
	return operation.NewCreateGCScheduleCreated()
}

func (g *gcAPI) UpdateGCSchedule(ctx context.Context, params operation.UpdateGCScheduleParams) middleware.Responder {
	if err := g.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceGarbageCollection); err != nil {
		return g.SendError(ctx, err)
	}
	_, err := g.kick(ctx, params.Schedule.Schedule.Type, params.Schedule.Schedule.Cron, params.Schedule.Parameters)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewUpdateGCScheduleOK()
}

func (g *gcAPI) kick(ctx context.Context, scheType string, cron string, parameters map[string]interface{}) (int64, error) {
	if parameters == nil {
		parameters = make(map[string]interface{})
	}
	// set the required parameters for GC
	parameters["redis_url_reg"] = os.Getenv("_REDIS_URL_REG")
	parameters["time_window"] = config.GetGCTimeWindow()

	var err error
	var id int64
	switch scheType {
	case ScheduleManual:
		policy := gc.Policy{
			ExtraAttrs: parameters,
		}
		if dryRun, ok := parameters["dry_run"].(bool); ok {
			policy.DryRun = dryRun
		}
		if deleteUntagged, ok := parameters["delete_untagged"].(bool); ok {
			policy.DeleteUntagged = deleteUntagged
		}
		if workers, ok := parameters["workers"].(json.Number); ok {
			wInt, err := workers.Int64()
			if err != nil {
				return 0, errors.BadRequestError(fmt.Errorf("workers should be integer format"))
			}
			if !validateWorkers(int(wInt)) {
				return 0, errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("Error: Invalid number of workers:%s. Workers must be greater than 0 and less than or equal to 5.", workers)
			}
			policy.Workers = int(wInt)
		}

		id, err = g.gcCtr.Start(ctx, policy, task.ExecutionTriggerManual)
	case ScheduleNone:
		err = g.gcCtr.DeleteSchedule(ctx)
	case ScheduleHourly, ScheduleDaily, ScheduleWeekly, ScheduleCustom:
		policy := gc.Policy{
			ExtraAttrs: parameters,
		}
		if dryRun, ok := parameters["dry_run"].(bool); ok {
			policy.DryRun = dryRun
		}
		if deleteUntagged, ok := parameters["delete_untagged"].(bool); ok {
			policy.DeleteUntagged = deleteUntagged
		}
		if workers, ok := parameters["workers"].(json.Number); ok {
			wInt, err := workers.Int64()
			if err != nil {
				return 0, errors.BadRequestError(fmt.Errorf("workers should be integer format"))
			}
			if !validateWorkers(int(wInt)) {
				return 0, errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("Error: Invalid number of workers:%s. Workers must be greater than 0 and less than or equal to 5.", workers)
			}
			policy.Workers = int(wInt)
		}
		err = g.updateSchedule(ctx, scheType, cron, policy)
	}
	return id, err
}

func (g *gcAPI) createSchedule(ctx context.Context, cronType, cron string, policy gc.Policy) error {
	_, err := g.gcCtr.CreateSchedule(ctx, cronType, cron, policy)
	if err != nil {
		return err
	}
	return nil
}

func (g *gcAPI) updateSchedule(ctx context.Context, cronType, cron string, policy gc.Policy) error {
	if err := utils.ValidateCronString(cron); err != nil {
		return errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("invalid cron string for scheduled gc: %s, error: %v", cron, err)
	}
	if err := g.gcCtr.DeleteSchedule(ctx); err != nil {
		return err
	}
	return g.createSchedule(ctx, cronType, cron, policy)
}

func (g *gcAPI) GetGCSchedule(ctx context.Context, _ operation.GetGCScheduleParams) middleware.Responder {
	if err := g.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceGarbageCollection); err != nil {
		return g.SendError(ctx, err)
	}
	schedule, err := g.gcCtr.GetSchedule(ctx)
	if errors.IsNotFoundErr(err) {
		return operation.NewGetGCScheduleOK()
	}
	if err != nil {
		return g.SendError(ctx, err)
	}

	return operation.NewGetGCScheduleOK().WithPayload(model.NewGCSchedule(schedule).ToSwagger())
}

func (g *gcAPI) GetGCHistory(ctx context.Context, params operation.GetGCHistoryParams) middleware.Responder {
	if err := g.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceGarbageCollection); err != nil {
		return g.SendError(ctx, err)
	}
	query, err := g.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return g.SendError(ctx, err)
	}
	total, err := g.gcCtr.ExecutionCount(ctx, query)
	if err != nil {
		return g.SendError(ctx, err)
	}
	execs, err := g.gcCtr.ListExecutions(ctx, query)
	if err != nil {
		return g.SendError(ctx, err)
	}

	var hs []*model.GCHistory
	for _, exec := range execs {
		extraAttrsString, err := json.Marshal(exec.ExtraAttrs)
		if err != nil {
			return g.SendError(ctx, err)
		}
		hs = append(hs, &model.GCHistory{
			ID:         exec.ID,
			Name:       job.GarbageCollectionVendorType,
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

	var results []*models.GCHistory
	for _, h := range hs {
		results = append(results, h.ToSwagger())
	}

	return operation.NewGetGCHistoryOK().
		WithXTotalCount(total).
		WithLink(g.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (g *gcAPI) GetGC(ctx context.Context, params operation.GetGCParams) middleware.Responder {
	if err := g.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceGarbageCollection); err != nil {
		return g.SendError(ctx, err)
	}
	exec, err := g.gcCtr.GetExecution(ctx, params.GCID)
	if err != nil {
		return g.SendError(ctx, err)
	}

	extraAttrsString, err := json.Marshal(exec.ExtraAttrs)
	if err != nil {
		return g.SendError(ctx, err)
	}

	res := &model.GCHistory{
		ID:         exec.ID,
		Name:       job.GarbageCollectionVendorType,
		Kind:       exec.Trigger,
		Parameters: string(extraAttrsString),
		Status:     exec.Status,
		Schedule: &model.ScheduleParam{
			Type: exec.Trigger,
		},
		CreationTime: exec.StartTime,
		UpdateTime:   exec.UpdateTime,
	}

	return operation.NewGetGCOK().WithPayload(res.ToSwagger())
}

func (g *gcAPI) GetGCLog(ctx context.Context, params operation.GetGCLogParams) middleware.Responder {
	if err := g.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceGarbageCollection); err != nil {
		return g.SendError(ctx, err)
	}
	tasks, err := g.gcCtr.ListTasks(ctx, q.New(q.KeyWords{
		"ExecutionID": params.GCID,
	}))
	if err != nil {
		return g.SendError(ctx, err)
	}
	if len(tasks) == 0 {
		return g.SendError(ctx, errors.New(nil).WithCode(errors.NotFoundCode).WithMessagef("garbage collection %d log is not found", params.GCID))
	}
	log, err := g.gcCtr.GetTaskLog(ctx, tasks[0].ID)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewGetGCLogOK().WithPayload(string(log))
}

func (g *gcAPI) StopGC(ctx context.Context, params operation.StopGCParams) middleware.Responder {
	if err := g.RequireSystemAccess(ctx, rbac.ActionStop, rbac.ResourceGarbageCollection); err != nil {
		return g.SendError(ctx, err)
	}

	if err := g.gcCtr.Stop(ctx, params.GCID); err != nil {
		return g.SendError(ctx, err)
	}

	return operation.NewStopGCOK()
}

func validateWorkers(workers int) bool {
	if workers <= 0 || workers > 5 {
		return false
	}
	return true
}
