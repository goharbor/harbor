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
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/member"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/permissions"
)

type permissionsAPI struct {
	BaseAPI
	uc user.Controller
	mc member.Controller
}

func newPermissionsAPIAPI() *permissionsAPI {
	return &permissionsAPI{
		uc: user.Ctl,
		mc: member.NewController(),
	}
}

func (p *permissionsAPI) GetPermissions(ctx context.Context, _ permissions.GetPermissionsParams) middleware.Responder {
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return p.SendError(ctx, errors.UnauthorizedError(errors.New("security context not found")))
	}
	if !secCtx.IsAuthenticated() {
		return p.SendError(ctx, errors.UnauthorizedError(nil).WithMessage(secCtx.GetUsername()))
	}

	var isSystemAdmin bool
	var isProjectAdmin bool

	if secCtx.IsSysAdmin() {
		isSystemAdmin = true
	} else {
		if sc, ok := secCtx.(*local.SecurityContext); ok {
			user := sc.User()
			var err error
			isProjectAdmin, err = p.mc.IsProjectAdmin(ctx, *user)
			if err != nil {
				return p.SendError(ctx, err)
			}
		}
	}
	if !isSystemAdmin && !isProjectAdmin {
		return p.SendError(ctx, errors.ForbiddenError(errors.New("only admins(system and project) can access permissions")))
	}

	sysPermissions := make([]*types.Policy, 0)
	proPermissions := rbac.PoliciesMap["Project"]
	if isSystemAdmin {
		// project admin cannot see the system level permissions
		sysPermissions = rbac.PoliciesMap["System"]
	}

	return permissions.NewGetPermissionsOK().WithPayload(p.convertPermissions(sysPermissions, proPermissions))
}

func (p *permissionsAPI) convertPermissions(system, project []*types.Policy) *models.Permissions {
	res := &models.Permissions{}
	if len(system) > 0 {
		var sysPermission []*models.Permission
		for _, item := range system {
			sysPermission = append(sysPermission, &models.Permission{
				Resource: item.Resource.String(),
				Action:   item.Action.String(),
			})
		}
		res.System = sysPermission
	}

	if len(project) > 0 {
		var proPermission []*models.Permission
		for _, item := range project {
			proPermission = append(proPermission, &models.Permission{
				Resource: item.Resource.String(),
				Action:   item.Action.String(),
			})
		}
		res.Project = proPermission
	}

	return res
}
