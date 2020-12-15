package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/controller/gc"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/gc"
	"os"
	"strings"
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

func (g *gcAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	if err := g.RequireSysAdmin(ctx); err != nil {
		return g.SendError(ctx, err)
	}
	return nil
}

func (g *gcAPI) CreateGCSchedule(ctx context.Context, params operation.CreateGCScheduleParams) middleware.Responder {
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
	_, err := g.kick(ctx, params.Schedule.Schedule.Type, params.Schedule.Schedule.Cron, params.Schedule.Parameters)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewUpdateGCScheduleOK()
}

func (g *gcAPI) kick(ctx context.Context, scheType string, cron string, parameters map[string]interface{}) (int64, error) {
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
		err = g.updateSchedule(ctx, scheType, cron, policy)
	}
	return id, err
}

func (g *gcAPI) createSchedule(ctx context.Context, cronType, cron string, policy gc.Policy) error {
	if cron == "" {
		return errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("empty cron string for gc schedule")
	}
	_, err := g.gcCtr.CreateSchedule(ctx, cronType, cron, policy)
	if err != nil {
		return err
	}
	return nil
}

func (g *gcAPI) updateSchedule(ctx context.Context, cronType, cron string, policy gc.Policy) error {
	if err := g.gcCtr.DeleteSchedule(ctx); err != nil {
		return err
	}
	return g.createSchedule(ctx, cronType, cron, policy)
}

func (g *gcAPI) GetGCSchedule(ctx context.Context, params operation.GetGCScheduleParams) middleware.Responder {
	schedule, err := g.gcCtr.GetSchedule(ctx)
	if errors.IsNotFoundErr(err) {
		return operation.NewGetGCScheduleOK().WithPayload(model.NewSchedule(&scheduler.Schedule{}).ToSwagger())
	}
	if err != nil {
		return g.SendError(ctx, err)
	}

	return operation.NewGetGCScheduleOK().WithPayload(model.NewSchedule(schedule).ToSwagger())
}

func (g *gcAPI) GetGCHistory(ctx context.Context, params operation.GetGCHistoryParams) middleware.Responder {
	query, err := g.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
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
			Name:       gc.GCVendorType,
			Kind:       exec.Trigger,
			Parameters: string(extraAttrsString),
			Schedule: &model.ScheduleParam{
				Type: exec.Trigger,
			},
			Status:       exec.Status,
			CreationTime: exec.StartTime,
			UpdateTime:   exec.EndTime,
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
	exec, err := g.gcCtr.GetExecution(ctx, params.GcID)
	if err != nil {
		return g.SendError(ctx, err)
	}

	extraAttrsString, err := json.Marshal(exec.ExtraAttrs)
	if err != nil {
		return g.SendError(ctx, err)
	}

	res := &model.GCHistory{
		ID:         exec.ID,
		Name:       gc.GCVendorType,
		Kind:       exec.Trigger,
		Parameters: string(extraAttrsString),
		Status:     exec.Status,
		Schedule: &model.ScheduleParam{
			Type: exec.Trigger,
		},
		CreationTime: exec.StartTime,
		UpdateTime:   exec.EndTime,
	}

	return operation.NewGetGCOK().WithPayload(res.ToSwagger())
}

func (g *gcAPI) GetGCLog(ctx context.Context, params operation.GetGCLogParams) middleware.Responder {
	log, err := g.gcCtr.GetTaskLog(ctx, params.GcID)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewGetGCLogOK().WithPayload(string(log))
}
