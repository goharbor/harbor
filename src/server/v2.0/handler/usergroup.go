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
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	ugCtl "github.com/goharbor/harbor/src/controller/usergroup"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/usergroup"
	"strings"
)

type userGroupAPI struct {
	BaseAPI
	ctl ugCtl.Controller
}

func newUserGroupAPI() *userGroupAPI {
	return &userGroupAPI{ctl: ugCtl.Ctl}
}

func (u *userGroupAPI) CreateUserGroup(ctx context.Context, params operation.CreateUserGroupParams) middleware.Responder {
	if err := u.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceUserGroup); err != nil {
		return u.SendError(ctx, err)
	}
	if params.Usergroup == nil {
		return operation.NewCreateUserGroupBadRequest()
	}
	if len(params.Usergroup.GroupName) == 0 {
		return operation.NewCreateUserGroupBadRequest()
	}
	ug := model.UserGroup{
		GroupName:   params.Usergroup.GroupName,
		GroupType:   int(params.Usergroup.GroupType),
		LdapGroupDN: params.Usergroup.LdapGroupDn,
	}
	id, err := u.ctl.Create(ctx, ug)
	if err != nil {
		return u.SendError(ctx, err)
	}
	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateUserGroupCreated().WithLocation(location)
}

func (u *userGroupAPI) DeleteUserGroup(ctx context.Context, params operation.DeleteUserGroupParams) middleware.Responder {
	if err := u.RequireSystemAccess(ctx, rbac.ActionDelete, rbac.ResourceUserGroup); err != nil {
		return u.SendError(ctx, err)
	}
	if params.GroupID <= 0 {
		return u.SendError(ctx, errors.BadRequestError(nil).WithMessage("the group id should be provided"))
	}
	err := u.ctl.Delete(ctx, int(params.GroupID))
	if err != nil {
		return u.SendError(ctx, err)
	}
	return operation.NewDeleteUserGroupOK()
}

func (u *userGroupAPI) GetUserGroup(ctx context.Context, params operation.GetUserGroupParams) middleware.Responder {
	if err := u.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceUserGroup); err != nil {
		return u.SendError(ctx, err)
	}
	if params.GroupID <= 0 {
		return u.SendError(ctx, errors.BadRequestError(nil).WithMessage("the group id should be provided"))
	}
	ug, err := u.ctl.Get(ctx, int(params.GroupID))
	if err != nil {
		return u.SendError(ctx, err)
	}
	if ug == nil {
		return u.SendError(ctx, errors.NotFoundError(nil).WithMessage("the user group with id %v is not found", params.GroupID))
	}
	userGroup := &models.UserGroup{
		GroupName:   ug.GroupName,
		GroupType:   int64(ug.GroupType),
		LdapGroupDn: ug.LdapGroupDN,
	}
	return operation.NewGetUserGroupOK().WithPayload(userGroup)
}

func (u *userGroupAPI) ListUserGroups(ctx context.Context, params operation.ListUserGroupsParams) middleware.Responder {
	if err := u.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceUserGroup); err != nil {
		return u.SendError(ctx, err)
	}
	authMode, err := config.AuthMode(ctx)
	if err != nil {
		return u.SendError(ctx, err)
	}
	query, err := u.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return u.SendError(ctx, err)
	}
	if params.GroupName != nil && len(*params.GroupName) > 0 {
		query.Keywords["GroupName"] = &q.FuzzyMatchValue{Value: *params.GroupName}
	}
	switch authMode {
	case common.LDAPAuth:
		query.Keywords["GroupType"] = common.LDAPGroupType
		if params.LdapGroupDn != nil && len(*params.LdapGroupDn) > 0 {
			query.Keywords["LdapGroupDN"] = *params.LdapGroupDn
		}
	case common.HTTPAuth:
		query.Keywords["GroupType"] = common.HTTPGroupType
	}

	total, err := u.ctl.Count(ctx, query)
	if err != nil {
		return u.SendError(ctx, err)
	}
	if total == 0 {
		return operation.NewListUserGroupsOK().WithXTotalCount(0).WithPayload([]*models.UserGroup{})
	}
	ug, err := u.ctl.List(ctx, query)
	if err != nil {
		return u.SendError(ctx, err)
	}
	return operation.NewListUserGroupsOK().
		WithXTotalCount(total).
		WithPayload(getUserGroupResp(ug)).
		WithLink(u.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}
func getUserGroupResp(ug []*model.UserGroup) []*models.UserGroup {
	result := make([]*models.UserGroup, 0)
	for _, u := range ug {
		ug := &models.UserGroup{
			GroupName:   u.GroupName,
			GroupType:   int64(u.GroupType),
			LdapGroupDn: u.LdapGroupDN,
			ID:          int64(u.ID),
		}
		result = append(result, ug)
	}
	return result
}
func getUserGroupSearchItem(ug []*model.UserGroup) []*models.UserGroupSearchItem {
	result := make([]*models.UserGroupSearchItem, 0)
	for _, u := range ug {
		ug := &models.UserGroupSearchItem{
			GroupName: u.GroupName,
			GroupType: int64(u.GroupType),
			ID:        int64(u.ID),
		}
		result = append(result, ug)
	}
	return result
}
func (u *userGroupAPI) UpdateUserGroup(ctx context.Context, params operation.UpdateUserGroupParams) middleware.Responder {
	if err := u.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceUserGroup); err != nil {
		return u.SendError(ctx, err)
	}
	if params.GroupID <= 0 {
		return operation.NewUpdateUserGroupBadRequest()
	}
	if params.Usergroup == nil || len(params.Usergroup.GroupName) == 0 {
		return operation.NewUpdateUserGroupBadRequest()
	}
	err := u.ctl.Update(ctx, int(params.GroupID), params.Usergroup.GroupName)
	if err != nil {
		return u.SendError(ctx, err)
	}
	return operation.NewUpdateUserGroupOK()
}

func (u *userGroupAPI) SearchUserGroups(ctx context.Context, params operation.SearchUserGroupsParams) middleware.Responder {
	if err := u.RequireAuthenticated(ctx); err != nil {
		return u.SendError(ctx, err)
	}
	query, err := u.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return u.SendError(ctx, err)
	}
	if len(params.Groupname) == 0 {
		return u.SendError(ctx, errors.BadRequestError(nil).WithMessage("need to provide groupname to search user group"))
	}
	query.Keywords["GroupName"] = &q.FuzzyMatchValue{Value: params.Groupname}
	total, err := u.ctl.Count(ctx, query)
	if err != nil {
		return u.SendError(ctx, err)
	}
	if total == 0 {
		return operation.NewSearchUserGroupsOK().WithXTotalCount(0).WithPayload([]*models.UserGroupSearchItem{})
	}
	ug, err := u.ctl.List(ctx, query)
	if err != nil {
		return u.SendError(ctx, err)
	}
	return operation.NewSearchUserGroupsOK().WithXTotalCount(total).
		WithPayload(getUserGroupSearchItem(ug)).
		WithLink(u.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}
