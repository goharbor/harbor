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
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/role"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	pkg "github.com/goharbor/harbor/src/pkg/role/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/role"
)

func newRoleAPI() *roleAPI {
	return &roleAPI{
		roleCtl: role.Ctl,
	}
}

type roleAPI struct {
	BaseAPI
	roleCtl role.Controller
}

func (rAPI *roleAPI) CreateRole(ctx context.Context, params operation.CreateRoleParams) middleware.Responder {
	if err := rAPI.checkSysAdmin(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := validateRoleName(params.Role.Name); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.validate(params.Role.Permissions); err != nil {
		return rAPI.SendError(ctx, err)
	}

	sc, err := rAPI.GetSecurityContext(ctx)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	r := &role.Role{
		Role: pkg.Role{
			Name:     params.Role.Name,
			RoleMask: params.Role.RoleMask,
			RoleCode: params.Role.RoleCode,
		},
	}

	switch s := sc.(type) {
	case *local.SecurityContext:
		if s.User() == nil {
			return rAPI.SendError(ctx, errors.New(nil).WithMessage("invalid security context: empty role account"))
		}
		r.CreatedBy = s.User().Username
	default:
		return rAPI.SendError(ctx, errors.New(nil).WithMessage("invalid security context"))
	}
	r.Description = params.Role.Description

	if err := lib.JSONCopy(&r.Permissions, params.Role.Permissions); err != nil {
		log.Warningf("failed to call JSONCopy on role permission when CreateRole, error: %v", err)
	}

	rid, err := rAPI.roleCtl.Create(ctx, r)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	created, err := rAPI.roleCtl.Get(ctx, rid, nil)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), created.ID)
	return operation.NewCreateRoleCreated().WithLocation(location).WithPayload(&models.RoleCreated{
		ID:   created.ID,
		Name: created.Name,
	})
}

func (rAPI *roleAPI) DeleteRole(ctx context.Context, params operation.DeleteRoleParams) middleware.Responder {
	if err := rAPI.checkSysAdmin(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	r, err := rAPI.roleCtl.Get(ctx, params.RoleID, nil)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.roleCtl.Delete(ctx, params.RoleID); err != nil {
		// for the version 1 role account, has to ignore the no permission error.
		if !r.Editable && errors.IsNotFoundErr(err) {
			return operation.NewDeleteRoleOK()
		}
		return rAPI.SendError(ctx, err)
	}
	return operation.NewDeleteRoleOK()
}

func (rAPI *roleAPI) ListRole(ctx context.Context, params operation.ListRoleParams) middleware.Responder {
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	query, err := rAPI.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	query.Keywords["Visible"] = true

	total, err := rAPI.roleCtl.Count(ctx, query)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	roles, err := rAPI.roleCtl.List(ctx, query, &role.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	var results []*models.Role
	for _, r := range roles {
		results = append(results, model.NewRole(r).ToSwagger())
	}

	return operation.NewListRoleOK().
		WithXTotalCount(total).
		WithLink(rAPI.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (rAPI *roleAPI) GetRoleByID(ctx context.Context, params operation.GetRoleByIDParams) middleware.Responder {
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	r, err := rAPI.roleCtl.Get(ctx, params.RoleID, &role.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	return operation.NewGetRoleByIDOK().WithPayload(model.NewRole(r).ToSwagger())
}

// checkSysAdmin returns UnauthorizedError for unauthenticated callers and
// ForbiddenError for callers who are not sysadmins.
func (rAPI *roleAPI) checkSysAdmin(ctx context.Context) error {
	sc, err := rAPI.GetSecurityContext(ctx)
	if err != nil {
		return err
	}
	if !sc.IsAuthenticated() {
		return errors.UnauthorizedError(nil)
	}
	if !sc.IsSysAdmin() {
		return errors.ForbiddenError(nil).WithMessage("only sysadmins can manage roles")
	}
	return nil
}

func (rAPI *roleAPI) UpdateRole(ctx context.Context, params operation.UpdateRoleParams) middleware.Responder {
	var err error
	if err := rAPI.checkSysAdmin(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}
	r, err := rAPI.roleCtl.Get(ctx, params.RoleID, &role.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	if !r.Editable {
		err = errors.DeniedError(nil).WithMessage("editing of legacy role is not allowed")
	} else {
		err = rAPI.updateV2Role(ctx, params, r)
	}
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	return operation.NewUpdateRoleOK()
}

// validate checks that every requested permission is a project-role-scoped
// access drawn from the role permission catalog (rbac.ScopeRole). Role creation
// and update are restricted to system admins (see checkSysAdmin), so there is no
// privilege escalation to guard against here: a sysadmin already holds every
// permission, and non-admins never reach this path
// (see TestCreateRole_NonSysAdminForbidden). Escalation is enforced separately on
// the paths where non-admin callers assign permissions (validateNoEscalation).
func (rAPI *roleAPI) validate(permissions []*models.RolePermission) error {
	if len(permissions) == 0 {
		return errors.New(nil).WithMessage("bad request empty permission").WithCode(errors.BadRequestCode)
	}

	for _, perm := range permissions {
		if len(perm.Access) == 0 {
			return errors.New(nil).WithMessage("bad request empty access").WithCode(errors.BadRequestCode)
		}
	}

	provider := rbac.GetPermissionProvider()
	// to validate the access scope
	for _, perm := range permissions {
		if perm.Kind == role.LEVELROLE {
			polices := provider.GetPermissions(rbac.ScopeRole)

			for _, acc := range perm.Access {
				if !containsAccess(polices, acc) {
					return errors.New(nil).WithMessagef("bad request permission: %s:%s", acc.Resource, acc.Action).WithCode(errors.BadRequestCode)
				}
			}
		} else {
			return errors.New(nil).WithMessagef("bad request permission level: %s", perm.Kind).WithCode(errors.BadRequestCode)
		}
	}

	return nil
}

func (rAPI *roleAPI) updateV2Role(ctx context.Context, params operation.UpdateRoleParams, r *role.Role) error {
	if err := rAPI.validate(params.Role.Permissions); err != nil {
		return err
	}

	if len(params.Role.Permissions) != 0 {
		if err := lib.JSONCopy(&r.Permissions, params.Role.Permissions); err != nil {
			log.Warningf("failed to call JSONCopy on role permission when updateV2Role, error: %v", err)
		}
	}
	r.Description = params.Role.Description

	if err := rAPI.roleCtl.Update(ctx, r, &role.Option{
		WithPermission: true,
	}); err != nil {
		return err
	}
	return nil
}

// validateName validates the role name, especially '+' cannot be a valid character
// validateRoleName enforces the server-side role name contract: non-empty, at
// most 255 characters (matching the role.name column and the UI), and free of the
// '+' reserved character.
func validateRoleName(name string) error {
	if name == "" {
		return errors.New(nil).WithMessage("role name cannot be empty").WithCode(errors.BadRequestCode)
	}
	if len(name) > 255 {
		return errors.New(nil).WithMessage("role name cannot exceed 255 characters").WithCode(errors.BadRequestCode)
	}
	if strings.Contains(name, "+") {
		return errors.New(nil).WithMessage("role name cannot contain the '+' character").WithCode(errors.BadRequestCode)
	}
	return nil
}

func containsRoleAccess(policies []*types.Policy, item *models.Access) bool {
	for _, po := range policies {
		if po.Resource.String() == item.Resource && po.Action.String() == item.Action {
			return true
		}
	}
	return false
}

// isValidPermissionScope checks if permission slice A is a subset of permission slice B
func isValidRolePermissionScope(creating []*models.RolePermission, creator []*role.Permission) bool {
	creatorMap := make(map[string]*role.Permission)
	for _, creatorPerm := range creator {
		key := fmt.Sprintf("%s:%s", creatorPerm.Kind, creatorPerm.Namespace)
		creatorMap[key] = creatorPerm
	}

	hasLessThanOrEqualAccess := func(creating []*models.Access, creator []*types.Policy) bool {
		creatorMap := make(map[string]*types.Policy)
		for _, creatorP := range creator {
			key := fmt.Sprintf("%s:%s:%s", creatorP.Resource, creatorP.Action, creatorP.Effect)
			creatorMap[key] = creatorP
		}
		for _, creatingP := range creating {
			key := fmt.Sprintf("%s:%s:%s", creatingP.Resource, creatingP.Action, creatingP.Effect)
			if _, found := creatorMap[key]; !found {
				return false
			}
		}
		return true
	}

	for _, pCreating := range creating {
		key := fmt.Sprintf("%s:%s", pCreating.Kind, pCreating.Namespace)
		creatorPerm, found := creatorMap[key]
		if !found {
			allProjects := fmt.Sprintf("%s:*", pCreating.Kind)
			if creatorPerm, found = creatorMap[allProjects]; !found {
				return false
			}
		}
		if !hasLessThanOrEqualAccess(pCreating.Access, creatorPerm.Access) {
			return false
		}
	}
	return true
}
