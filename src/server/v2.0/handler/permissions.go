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
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/permissions"
)

type permissionsAPI struct {
	BaseAPI
	uc user.Controller
}

func newPermissionsAPIAPI() *permissionsAPI {
	return &permissionsAPI{
		uc: user.Ctl,
	}
}

// GetPermissions returns the available permission catalog.
// Accessible to all authenticated users — the catalog is read-only metadata
// (action names such as repository:pull) used by project-level UI components
// (robot accounts, webhooks, role management). System-level permissions are
// only included in the response for system admins.
func (p *permissionsAPI) GetPermissions(ctx context.Context, _ permissions.GetPermissionsParams) middleware.Responder {
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return p.SendError(ctx, errors.UnauthorizedError(errors.New("security context not found")))
	}
	if !secCtx.IsAuthenticated() {
		return p.SendError(ctx, errors.UnauthorizedError(nil).WithMessage(secCtx.GetUsername()))
	}

	isSystemAdmin := secCtx.IsSysAdmin()

	provider := rbac.GetPermissionProvider()
	sysPermissions := make([]*types.Policy, 0)
	proPermissions := provider.GetPermissions(rbac.ScopeProject)
	if isSystemAdmin {
		sysPermissions = provider.GetPermissions(rbac.ScopeSystem)
	}
	rolePermissions := provider.GetPermissions(rbac.ScopeRole)

	return permissions.NewGetPermissionsOK().WithPayload(p.convertPermissions(sysPermissions, proPermissions, rolePermissions))
}

func (p *permissionsAPI) convertPermissions(system, project, role []*types.Policy) *models.Permissions {
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

	if len(role) > 0 {
		var rolePermission []*models.Permission
		for _, item := range role {
			rolePermission = append(rolePermission, &models.Permission{
				Resource: item.Resource.String(),
				Action:   item.Action.String(),
			})
		}
		res.Role = rolePermission
	}

	return res
}
