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

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/label"
	pkg_model "github.com/goharbor/harbor/src/pkg/label/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/label"
)

func newLabelAPI() *labelAPI {
	return &labelAPI{
		labelMgr:   label.Mgr,
		projectCtl: project.Ctl,
	}
}

type labelAPI struct {
	BaseAPI
	labelMgr   label.Manager
	projectCtl project.Controller
}

func (lAPI *labelAPI) CreateLabel(ctx context.Context, params operation.CreateLabelParams) middleware.Responder {
	label := &pkg_model.Label{}
	if err := lib.JSONCopy(label, params.Label); err != nil {
		log.Warningf("failed to call JSONCopy on Label when CreateLabel, error: %v", err)
	}

	label.Level = common.LabelLevelUser
	if label.Scope == common.LabelScopeGlobal {
		label.ProjectID = 0
	}

	if err := lAPI.requireAccess(ctx, label, rbac.ActionCreate); err != nil {
		return lAPI.SendError(ctx, err)
	}

	id, err := lAPI.labelMgr.Create(ctx, label)
	if err != nil {
		return lAPI.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateLabelCreated().WithLocation(location)
}

func (lAPI *labelAPI) GetLabelByID(ctx context.Context, params operation.GetLabelByIDParams) middleware.Responder {
	label, err := lAPI.labelMgr.Get(ctx, params.LabelID)
	if err != nil {
		return lAPI.SendError(ctx, err)
	}
	if label == nil || label.Deleted {
		return lAPI.SendError(ctx, errors.New(nil).WithMessagef("label %d not found", params.LabelID).WithCode(errors.NotFoundCode))
	}

	if err := lAPI.requireAccess(ctx, label, rbac.ActionRead); err != nil {
		return lAPI.SendError(ctx, err)
	}

	return operation.NewGetLabelByIDOK().WithPayload(model.NewLabel(label).ToSwagger())
}

func (lAPI *labelAPI) ListLabels(ctx context.Context, params operation.ListLabelsParams) middleware.Responder {
	query, err := lAPI.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return lAPI.SendError(ctx, err)
	}

	scope := lib.StringValue(params.Scope)
	if scope != common.LabelScopeGlobal && scope != common.LabelScopeProject {
		return lAPI.SendError(ctx, errors.New(nil).WithMessagef("invalid scope: %s", scope).WithCode(errors.BadRequestCode))
	}
	query.Keywords["Level"] = common.LabelLevelUser
	query.Keywords["Scope"] = scope
	name := lib.StringValue(params.Name)
	if name != "" {
		query.Keywords["name"] = &q.FuzzyMatchValue{Value: name}
	}
	if scope == common.LabelScopeProject {
		pid := lib.Int64Value(params.ProjectID)
		if pid == 0 {
			return lAPI.SendError(ctx, errors.BadRequestError(nil).WithMessage("must with project ID when to query project labels"))
		}
		if err := lAPI.RequireProjectAccess(ctx, pid, rbac.ActionList, rbac.ResourceLabel); err != nil {
			return lAPI.SendError(ctx, err)
		}
		query.Keywords["ProjectID"] = pid
	}

	results := make([]*models.Label, 0)
	total, err := lAPI.labelMgr.Count(ctx, query)
	if err != nil {
		return lAPI.SendError(ctx, err)
	}
	if total > 0 {
		labels, err := lAPI.labelMgr.List(ctx, query)
		if err != nil {
			return lAPI.SendError(ctx, err)
		}

		for _, l := range labels {
			results = append(results, model.NewLabel(l).ToSwagger())
		}
	}

	return operation.NewListLabelsOK().
		WithXTotalCount(total).
		WithLink(lAPI.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (lAPI *labelAPI) UpdateLabel(ctx context.Context, params operation.UpdateLabelParams) middleware.Responder {
	labelData := &pkg_model.Label{}
	if err := lib.JSONCopy(labelData, params.Label); err != nil {
		log.Warningf("failed to call JSONCopy on Label when UpdateLabel, error: %v", err)
	}

	label, err := lAPI.labelMgr.Get(ctx, params.LabelID)
	if err != nil {
		return lAPI.SendError(ctx, err)
	}
	if label == nil || label.Deleted {
		return lAPI.SendError(ctx, errors.New(nil).WithMessagef("label %d not found", params.LabelID).WithCode(errors.NotFoundCode))
	}

	if err := lAPI.requireAccess(ctx, label, rbac.ActionUpdate); err != nil {
		return lAPI.SendError(ctx, err)
	}

	label.Name = labelData.Name
	label.Description = labelData.Description
	label.Color = labelData.Color

	if err := label.Valid(); err != nil {
		return lAPI.SendError(ctx, err)
	}

	if err := lAPI.labelMgr.Update(ctx, label); err != nil {
		return lAPI.SendError(ctx, err)
	}

	return operation.NewUpdateLabelOK()
}

func (lAPI *labelAPI) DeleteLabel(ctx context.Context, params operation.DeleteLabelParams) middleware.Responder {
	label, err := lAPI.labelMgr.Get(ctx, params.LabelID)
	if err != nil {
		return lAPI.SendError(ctx, err)
	}
	if err := lAPI.requireAccess(ctx, label, rbac.ActionDelete); err != nil {
		return lAPI.SendError(ctx, err)
	}
	id := label.ID
	if err := lAPI.labelMgr.RemoveFromAllArtifacts(ctx, id); err != nil {
		return lAPI.SendError(ctx, err)
	}
	if err := lAPI.labelMgr.Delete(ctx, id); err != nil {
		return lAPI.SendError(ctx, err)
	}

	return operation.NewDeleteLabelOK()
}

func (lAPI *labelAPI) requireAccess(ctx context.Context, label *pkg_model.Label, action rbac.Action, subresources ...rbac.Resource) error {
	switch label.Scope {
	case common.LabelScopeGlobal:
		return lAPI.RequireSystemAccess(ctx, action, rbac.ResourceLabel)
	case common.LabelScopeProject:
		if len(subresources) == 0 {
			subresources = append(subresources, rbac.ResourceLabel)
		}
		return lAPI.RequireProjectAccess(ctx, label.ProjectID, action, subresources...)
	}
	return errors.New("unsupported label scope").WithCode(errors.BadRequestCode)
}
