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

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/quota"
)

func newQuotaAPI() *quotaAPI {
	return &quotaAPI{
		quotaCtl: quota.Ctl,
	}
}

type quotaAPI struct {
	BaseAPI
	quotaCtl quota.Controller
}

func (qa *quotaAPI) GetQuota(ctx context.Context, params operation.GetQuotaParams) middleware.Responder {
	if err := qa.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceQuota); err != nil {
		return qa.SendError(ctx, err)
	}

	quota, err := qa.quotaCtl.Get(ctx, params.ID, quota.WithReferenceObject())
	if err != nil {
		return qa.SendError(ctx, err)

	}
	return operation.NewGetQuotaOK().WithPayload(model.NewQuota(quota).ToSwagger(ctx))
}

func (qa *quotaAPI) ListQuotas(ctx context.Context, params operation.ListQuotasParams) middleware.Responder {
	if err := qa.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceQuota); err != nil {
		return qa.SendError(ctx, err)
	}

	query := &q.Query{
		Keywords: q.KeyWords{
			"reference":    lib.StringValue(params.Reference),
			"reference_id": lib.StringValue(params.ReferenceID),
		},
		PageNumber: *params.Page,
		PageSize:   *params.PageSize,
		Sorting:    lib.StringValue(params.Sort),
	}

	total, err := qa.quotaCtl.Count(ctx, query)
	if err != nil {
		return qa.SendError(ctx, err)
	}

	quotas, err := qa.quotaCtl.List(ctx, query, quota.WithReferenceObject())
	if err != nil {
		return qa.SendError(ctx, err)
	}

	payload := make([]*models.Quota, len(quotas))
	for i, quota := range quotas {
		payload[i] = model.NewQuota(quota).ToSwagger(ctx)
	}

	return operation.NewListQuotasOK().
		WithXTotalCount(total).
		WithLink(qa.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(payload)
}

func (qa *quotaAPI) UpdateQuota(ctx context.Context, params operation.UpdateQuotaParams) middleware.Responder {
	if err := qa.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceQuota); err != nil {
		return qa.SendError(ctx, err)
	}

	if params.Hard == nil || len(params.Hard.Hard) == 0 {
		return qa.SendError(ctx, errors.BadRequestError(nil).WithMessage("hard required in body"))
	}

	q, err := qa.quotaCtl.Get(ctx, params.ID)
	if err != nil {
		return qa.SendError(ctx, err)
	}

	if err := quota.Validate(ctx, q.Reference, params.Hard.Hard); err != nil {
		return qa.SendError(ctx, errors.BadRequestError(nil).WithMessage(err.Error()))
	}

	q.SetHard(params.Hard.Hard)

	if err := qa.quotaCtl.Update(ctx, q); err != nil {
		return qa.SendError(ctx, err)
	}

	return operation.NewUpdateQuotaOK()
}
