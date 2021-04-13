//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/member"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	memberModels "github.com/goharbor/harbor/src/pkg/member/models"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/member"
)

type memberAPI struct {
	BaseAPI
	ctl member.Controller
}

func newMemberAPI() *memberAPI {
	return &memberAPI{ctl: member.NewController()}
}

func (m *memberAPI) CreateProjectMember(ctx context.Context, params operation.CreateProjectMemberParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := m.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceMember); err != nil {
		return m.SendError(ctx, err)
	}
	if params.ProjectMember == nil {
		return m.SendError(ctx, errors.BadRequestError(nil).WithMessage("the project member should provide"))
	}
	req, err := toMemberReq(params.ProjectMember)
	if err != nil {
		return m.SendError(ctx, err)
	}
	id, err := m.ctl.Create(ctx, projectNameOrID, *req)
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewCreateProjectMemberCreated().
		WithLocation(fmt.Sprintf("/api/v2.0/projects/%v/members/%d", projectNameOrID, id))
}

func toMemberReq(memberReq *models.ProjectMember) (*member.Request, error) {
	data, err := json.Marshal(memberReq)
	if err != nil {
		return nil, err
	}
	var result member.Request
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *memberAPI) DeleteProjectMember(ctx context.Context, params operation.DeleteProjectMemberParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := m.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete, rbac.ResourceMember); err != nil {
		return m.SendError(ctx, err)
	}
	if params.Mid == 0 {
		return m.SendError(ctx, errors.BadRequestError(nil).WithMessage("the project member id is required."))
	}
	err := m.ctl.Delete(ctx, projectNameOrID, int(params.Mid))
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewDeleteProjectMemberOK()
}

func (m *memberAPI) GetProjectMember(ctx context.Context, params operation.GetProjectMemberParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := m.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceMember); err != nil {
		return m.SendError(ctx, err)
	}

	if params.Mid == 0 {
		return m.SendError(ctx, errors.BadRequestError(nil).WithMessage("the member id can not be empty!"))
	}

	member, err := m.ctl.Get(ctx, projectNameOrID, int(params.Mid))
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewGetProjectMemberOK().WithPayload(toProjectMemberResp(member))
}

func (m *memberAPI) ListProjectMembers(ctx context.Context, params operation.ListProjectMembersParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := m.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceMember); err != nil {
		return m.SendError(ctx, err)
	}
	entityName := lib.StringValue(params.Entityname)
	query, err := m.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return m.SendError(ctx, err)
	}
	total, err := m.ctl.Count(ctx, projectNameOrID, query)
	if err != nil {
		return m.SendError(ctx, err)
	}
	if total == 0 {
		return operation.NewListProjectMembersOK().
			WithXTotalCount(0).
			WithPayload([]*models.ProjectMemberEntity{})
	}
	members, err := m.ctl.List(ctx, projectNameOrID, entityName, query)
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewListProjectMembersOK().
		WithXTotalCount(int64(total)).
		WithLink(m.Links(ctx, params.HTTPRequest.URL, int64(total), query.PageNumber, query.PageSize).String()).
		WithPayload(toProjectMemberRespList(members))
}

func toProjectMemberRespList(members []*memberModels.Member) []*models.ProjectMemberEntity {
	result := make([]*models.ProjectMemberEntity, 0)
	for _, mem := range members {
		result = append(result, toProjectMemberResp(mem))
	}
	return result
}

func toProjectMemberResp(member *memberModels.Member) *models.ProjectMemberEntity {
	return &models.ProjectMemberEntity{
		ProjectID:  member.ProjectID,
		ID:         int64(member.ID),
		EntityName: member.Entityname,
		EntityID:   int64(member.EntityID),
		EntityType: member.EntityType,
		RoleID:     int64(member.Role),
		RoleName:   member.Rolename,
	}
}

func (m *memberAPI) UpdateProjectMember(ctx context.Context, params operation.UpdateProjectMemberParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := m.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceMember); err != nil {
		return m.SendError(ctx, err)
	}
	if params.Role == nil {
		return m.SendError(ctx, errors.BadRequestError(nil).WithMessage("role can not be empty!"))
	}
	if params.Mid == 0 {
		return m.SendError(ctx, errors.BadRequestError(nil).WithMessage("member id can not be empty!"))
	}

	err := m.ctl.UpdateRole(ctx, projectNameOrID, int(params.Mid), int(params.Role.RoleID))
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewUpdateProjectMemberOK()
}
