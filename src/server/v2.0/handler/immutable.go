package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
	handler_model "github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/immutable"
	"strings"
)

func newImmutableAPI() *immutableAPI {
	return &immutableAPI{
		immuCtl:    immutable.Ctr,
		projectCtr: project.Ctl,
	}
}

type immutableAPI struct {
	BaseAPI
	immuCtl    immutable.Controller
	projectCtr project.Controller
}

func (ia *immutableAPI) CreateImmuRule(ctx context.Context, params operation.CreateImmuRuleParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := ia.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceImmutableTag); err != nil {
		return ia.SendError(ctx, err)
	}

	metadata := model.Metadata{}
	lib.JSONCopy(&metadata, params.ImmutableRule)

	projectID, err := ia.getProjectID(ctx, projectNameOrID)
	if err != nil {
		return ia.SendError(ctx, err)
	}
	metadata.ProjectID = projectID

	id, err := ia.immuCtl.CreateImmutableRule(ctx, &metadata)
	if err != nil {
		return ia.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateImmuRuleCreated().WithLocation(location)
}

func (ia *immutableAPI) DeleteImmuRule(ctx context.Context, params operation.DeleteImmuRuleParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := ia.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete, rbac.ResourceImmutableTag); err != nil {
		return ia.SendError(ctx, err)
	}

	if err := ia.immuCtl.DeleteImmutableRule(ctx, params.ImmutableRuleID); err != nil {
		return ia.SendError(ctx, err)
	}

	return operation.NewDeleteImmuRuleOK()
}

func (ia *immutableAPI) UpdateImmuRule(ctx context.Context, params operation.UpdateImmuRuleParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := ia.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceImmutableTag); err != nil {
		return ia.SendError(ctx, err)
	}

	metadata := model.Metadata{}
	lib.JSONCopy(&metadata, params.ImmutableRule)

	projectID, err := ia.getProjectID(ctx, projectNameOrID)
	if err != nil {
		return ia.SendError(ctx, err)
	}
	metadata.ProjectID = projectID

	if err := ia.immuCtl.UpdateImmutableRule(ctx, projectID, &metadata); err != nil {
		return ia.SendError(ctx, err)
	}

	return operation.NewUpdateImmuRuleOK()
}

func (ia *immutableAPI) ListImmuRules(ctx context.Context, params operation.ListImmuRulesParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := ia.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceImmutableTag); err != nil {
		return ia.SendError(ctx, err)
	}

	query, err := ia.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return ia.SendError(ctx, err)
	}

	projectID, err := ia.getProjectID(ctx, projectNameOrID)
	if err != nil {
		return ia.SendError(ctx, err)
	}
	query.Keywords["ProjectID"] = projectID

	total, err := ia.immuCtl.Count(ctx, query)
	if err != nil {
		return ia.SendError(ctx, err)
	}

	rules, err := ia.immuCtl.ListImmutableRules(ctx, query)
	if err != nil {
		return ia.SendError(ctx, err)
	}

	var results []*models.ImmutableRule
	for _, r := range rules {
		results = append(results, handler_model.NewImmutableRule(r).ToSwagger())
	}

	return operation.NewListImmuRulesOK().
		WithXTotalCount(total).
		WithLink(ia.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (ia *immutableAPI) getProjectID(ctx context.Context, projectNameOrID interface{}) (int64, error) {
	projectName, ok := projectNameOrID.(string)
	if ok {
		p, err := ia.projectCtr.Get(ctx, projectName, project.Metadata(false))
		if err != nil {
			return 0, err
		}
		return p.ProjectID, nil
	}
	projectID, ok := projectNameOrID.(int64)
	if ok {
		return projectID, nil
	}
	return 0, errors.New("unknown project identifier type")
}
