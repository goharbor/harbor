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
	"strings"

	"github.com/go-openapi/runtime/middleware"

	"github.com/goharbor/harbor/src/common/rbac"
	jobserviceCtl "github.com/goharbor/harbor/src/controller/jobservice"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/schedule"
)

const all = "all"

type scheduleAPI struct {
	BaseAPI
	jobServiceCtl jobserviceCtl.SchedulerController
}

func newScheduleAPI() *scheduleAPI {
	return &scheduleAPI{
		jobServiceCtl: jobserviceCtl.SchedulerCtl,
	}
}

func (s *scheduleAPI) GetSchedulePaused(ctx context.Context, params schedule.GetSchedulePausedParams) middleware.Responder {
	if err := s.RequireAuthenticated(ctx); err != nil {
		return s.SendError(ctx, err)
	}
	if !strings.EqualFold(params.JobType, all) {
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage("job_type can only be 'all'"))
	}
	paused, err := s.jobServiceCtl.Paused(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return schedule.NewGetSchedulePausedOK().WithPayload(&models.SchedulerStatus{
		Paused: paused,
	})
}

func (s *scheduleAPI) ListSchedules(ctx context.Context, params schedule.ListSchedulesParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceJobServiceMonitor); err != nil {
		return s.SendError(ctx, err)
	}
	query, err := s.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return s.SendError(ctx, err)
	}
	count, err := s.jobServiceCtl.Count(ctx, query)
	if err != nil {
		return s.SendError(ctx, err)
	}
	schs, err := s.jobServiceCtl.List(ctx, query)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return schedule.NewListSchedulesOK().
		WithPayload(toScheduleResponse(schs)).
		WithXTotalCount(count).
		WithLink(s.Links(ctx, params.HTTPRequest.URL, count, query.PageNumber, query.PageSize).String())
}
