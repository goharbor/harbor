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
	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
	handler_model "github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/immutable"
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
	if err := lib.JSONCopy(&metadata, params.ImmutableRule); err != nil {
		log.Warningf("failed to call JSONCopy into Metadata of the immutable rule when CreateImmuRule, error: %v", err)
	}

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

	projectID, err := ia.getProjectID(ctx, projectNameOrID)
	if err != nil {
		return ia.SendError(ctx, err)
	}

	if err := ia.requireRuleAccess(ctx, projectID, params.ImmutableRuleID); err != nil {
		return ia.SendError(ctx, err)
	}

	if err := ia.immuCtl.DeleteImmutableRule(ctx, params.ImmutableRuleID); err != nil {
		return ia.SendError(ctx, err)
	}

	return operation.NewDeleteImmuRuleOK()
}

func (ia *immutableAPI) UpdateImmuRule(ctx context.Context, params operation.UpdateImmuRuleParams) middleware.Responder {
	if params.ImmutableRuleID != params.ImmutableRule.ID {
		return ia.SendError(ctx, errors.BadRequestError(fmt.Errorf("the immutable_rule_id doesn't match the id in the payload body of ImmutableRule")))
	}
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := ia.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceImmutableTag); err != nil {
		return ia.SendError(ctx, err)
	}

	metadata := model.Metadata{}
	if err := lib.JSONCopy(&metadata, params.ImmutableRule); err != nil {
		log.Warningf("failed to call JSONCopy into Metadata of the immutable rule when UpdateImmuRule, error: %v", err)
	}

	projectID, err := ia.getProjectID(ctx, projectNameOrID)
	if err != nil {
		return ia.SendError(ctx, err)
	}
	metadata.ProjectID = projectID

	if err = ia.requireRuleAccess(ctx, projectID, metadata.ID); err != nil {
		return ia.SendError(ctx, err)
	}

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

// requireRuleAccess checks whether the project has the permission to the
// immutable rule.
func (ia *immutableAPI) requireRuleAccess(ctx context.Context, projectID, metadataID int64) error {
	rule, err := ia.immuCtl.GetImmutableRule(ctx, metadataID)
	if err != nil {
		return err
	}
	// if input project id does not equal projectID in db return err
	if rule.ProjectID != projectID {
		return errors.NotFoundError(errors.Errorf("project id %d does not match", projectID))
	}

	return nil
}
