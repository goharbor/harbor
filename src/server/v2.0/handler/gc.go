package handler

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/controller/gc"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/gc"
	"os"
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

func (g *gcAPI) PostSchedule(ctx context.Context, params operation.PostScheduleParams) middleware.Responder {
	if err := g.parseParam(ctx, params.Schedule.Schedule.Type, params.Schedule.Schedule.Cron, params.Schedule.Parameters); err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewPostScheduleOK()
}

func (g *gcAPI) PutSchedule(ctx context.Context, params operation.PutScheduleParams) middleware.Responder {
	if err := g.parseParam(ctx, params.Schedule.Schedule.Type, params.Schedule.Schedule.Cron, params.Schedule.Parameters); err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewPutScheduleOK()
}

func (g *gcAPI) parseParam(ctx context.Context, scheType string, cron string, parameters map[string]interface{}) error {
	// set the required parameters for GC
	parameters["redis_url_reg"] = os.Getenv("_REDIS_URL_REG")
	parameters["time_window"] = config.GetGCTimeWindow()

	var err error
	switch scheType {
	case model.ScheduleManual:
		err = g.gcCtr.Start(ctx, parameters)
	case model.ScheduleNone:
		err = g.gcCtr.DeleteSchedule(ctx)
	case model.ScheduleHourly, model.ScheduleDaily, model.ScheduleWeekly, model.ScheduleCustom:
		err = g.updateSchedule(ctx, cron, parameters)
	}
	return err
}

func (g *gcAPI) createSchedule(ctx context.Context, cron string, parameters map[string]interface{}) error {
	if cron == "" {
		return errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("empty cron string for gc schedule")
	}
	_, err := g.gcCtr.CreateSchedule(ctx, cron, parameters)
	if err != nil {
		return err
	}
	return nil
}

func (g *gcAPI) updateSchedule(ctx context.Context, cron string, parameters map[string]interface{}) error {
	if err := g.gcCtr.DeleteSchedule(ctx); err != nil {
		return err
	}
	return g.createSchedule(ctx, cron, parameters)
}

func (g *gcAPI) GetSchedule(ctx context.Context, params operation.GetScheduleParams) middleware.Responder {
	schedule, err := g.gcCtr.GetSchedule(ctx)
	if errors.IsNotFoundErr(err) {
		return operation.NewGetScheduleOK().WithPayload(model.NewSchedule(&scheduler.Schedule{}).ToSwagger())
	}
	if err != nil {
		return g.SendError(ctx, err)
	}

	return operation.NewGetScheduleOK().WithPayload(model.NewSchedule(schedule).ToSwagger())
}

func (g *gcAPI) GetGCHistory(ctx context.Context, params operation.GetGCHistoryParams) middleware.Responder {
	query, err := g.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return g.SendError(ctx, err)
	}
	total, err := g.gcCtr.Count(ctx, query)
	if err != nil {
		return g.SendError(ctx, err)
	}
	hs, err := g.gcCtr.History(ctx, query)
	if err != nil {
		return g.SendError(ctx, err)
	}
	var results []*models.GCHistory
	for _, h := range hs {
		res := &model.GCHistory{}
		res.History = h
		results = append(results, res.ToSwagger())
	}
	return operation.NewGetGCHistoryOK().
		WithXTotalCount(total).
		WithLink(g.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (g *gcAPI) GetGCLog(ctx context.Context, params operation.GetGCLogParams) middleware.Responder {
	log, err := g.gcCtr.GetLog(ctx, params.GcID)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewGetGCLogOK().WithPayload(string(log))
}
