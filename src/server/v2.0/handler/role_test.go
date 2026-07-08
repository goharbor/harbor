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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	roleCtl "github.com/goharbor/harbor/src/controller/role"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/role"
	securityMock "github.com/goharbor/harbor/src/testing/common/security"
)

// ---------------------------------------------------------------------------
// checkSysAdmin
// ---------------------------------------------------------------------------

func TestCheckSysAdmin_SysAdmin(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsAuthenticated").Return(true)
	sc.On("IsSysAdmin").Return(true)

	err := (&roleAPI{}).checkSysAdmin(newCtxWithSecurity(sc))
	assert.NoError(t, err)
}

func TestCheckSysAdmin_NonSysAdminForbidden(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsAuthenticated").Return(true)
	sc.On("IsSysAdmin").Return(false)

	err := (&roleAPI{}).checkSysAdmin(newCtxWithSecurity(sc))
	assert.Error(t, err)
	assert.Equal(t, errors.ForbiddenCode, errors.ErrCode(err))
}

func TestCheckSysAdmin_UnauthenticatedUnauthorized(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsAuthenticated").Return(false)

	err := (&roleAPI{}).checkSysAdmin(newCtxWithSecurity(sc))
	assert.Error(t, err)
	assert.Equal(t, errors.UnAuthorizedCode, errors.ErrCode(err))
}

// ---------------------------------------------------------------------------
// CreateRole authorization gate (escalation prevention)
// ---------------------------------------------------------------------------

// TestCreateRole_NonSysAdminForbidden proves that role definition is restricted
// to system admins: a non-sysadmin attempting to create a role with broad
// permissions is rejected at the checkSysAdmin gate before any role is created,
// so there is no privilege-escalation path through this handler. roleCtl is
// intentionally nil — the request must never reach it.
func TestCreateRole_NonSysAdminForbidden(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsAuthenticated").Return(true)
	sc.On("IsSysAdmin").Return(false)

	api := &roleAPI{}
	params := operation.CreateRoleParams{
		Role: &models.RoleCreate{
			Name: "escalated-role",
			Permissions: []*models.RolePermission{
				{
					Kind:      roleCtl.LEVELROLE,
					Namespace: "*",
					Access: []*models.Access{
						{Resource: "repository", Action: "push"},
					},
				},
			},
		},
	}

	resp := api.CreateRole(newCtxWithSecurity(sc), params)

	rr := httptest.NewRecorder()
	resp.WriteResponse(rr, runtime.JSONProducer())
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// ---------------------------------------------------------------------------
// validate
// ---------------------------------------------------------------------------

func TestValidate_EmptyPermissions(t *testing.T) {
	err := (&roleAPI{}).validate(nil)
	assert.Error(t, err)
	assert.Equal(t, errors.BadRequestCode, errors.ErrCode(err))
}

func TestValidate_EmptyAccess(t *testing.T) {
	err := (&roleAPI{}).validate([]*models.RolePermission{
		{Kind: roleCtl.LEVELROLE, Access: []*models.Access{}},
	})
	assert.Error(t, err)
	assert.Equal(t, errors.BadRequestCode, errors.ErrCode(err))
}

func TestValidate_WrongKind(t *testing.T) {
	err := (&roleAPI{}).validate([]*models.RolePermission{
		{Kind: "system", Access: []*models.Access{{Resource: "member", Action: "create"}}},
	})
	assert.Error(t, err)
	assert.Equal(t, errors.BadRequestCode, errors.ErrCode(err))
}

func TestValidate_UnknownResourceAction(t *testing.T) {
	err := (&roleAPI{}).validate([]*models.RolePermission{
		{Kind: roleCtl.LEVELROLE, Access: []*models.Access{{Resource: "notaresource", Action: "notanaction"}}},
	})
	assert.Error(t, err)
	assert.Equal(t, errors.BadRequestCode, errors.ErrCode(err))
}

func TestValidate_ValidPermission(t *testing.T) {
	// "member:create" is a known ScopeRole permission (src/common/rbac/const.go)
	err := (&roleAPI{}).validate([]*models.RolePermission{
		{Kind: roleCtl.LEVELROLE, Access: []*models.Access{{Resource: "member", Action: "create"}}},
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// containsRoleAccess
// ---------------------------------------------------------------------------

func TestContainsRoleAccess_Present(t *testing.T) {
	policies := []*types.Policy{
		{Resource: "member", Action: "create"},
		{Resource: "member", Action: "delete"},
	}
	assert.True(t, containsRoleAccess(policies, &models.Access{Resource: "member", Action: "create"}))
}

func TestContainsRoleAccess_Absent(t *testing.T) {
	policies := []*types.Policy{
		{Resource: "member", Action: "create"},
	}
	assert.False(t, containsRoleAccess(policies, &models.Access{Resource: "member", Action: "delete"}))
}

func TestContainsRoleAccess_EmptyPolicies(t *testing.T) {
	assert.False(t, containsRoleAccess(nil, &models.Access{Resource: "member", Action: "create"}))
}

// ---------------------------------------------------------------------------
// isValidRolePermissionScope
// ---------------------------------------------------------------------------

func rolePermissions(kind, ns string, accesses ...[2]string) []*models.RolePermission {
	var access []*models.Access
	for _, a := range accesses {
		access = append(access, &models.Access{Resource: a[0], Action: a[1]})
	}
	return []*models.RolePermission{{Kind: kind, Namespace: ns, Access: access}}
}

func creatorPerms(kind, ns string, accesses ...[2]string) []*roleCtl.Permission {
	var policies []*types.Policy
	for _, a := range accesses {
		policies = append(policies, &types.Policy{Resource: types.Resource(a[0]), Action: types.Action(a[1])})
	}
	return []*roleCtl.Permission{{Kind: kind, Namespace: ns, Access: policies}}
}

func TestIsValidRolePermissionScope_Subset(t *testing.T) {
	creating := rolePermissions(roleCtl.LEVELROLE, "*", [2]string{"repository", "pull"})
	creator := creatorPerms(roleCtl.LEVELROLE, "*", [2]string{"repository", "pull"}, [2]string{"repository", "push"})
	assert.True(t, isValidRolePermissionScope(creating, creator))
}

func TestIsValidRolePermissionScope_Equal(t *testing.T) {
	creating := rolePermissions(roleCtl.LEVELROLE, "*", [2]string{"repository", "pull"}, [2]string{"repository", "push"})
	creator := creatorPerms(roleCtl.LEVELROLE, "*", [2]string{"repository", "pull"}, [2]string{"repository", "push"})
	assert.True(t, isValidRolePermissionScope(creating, creator))
}

func TestIsValidRolePermissionScope_ExcessPerm(t *testing.T) {
	creating := rolePermissions(roleCtl.LEVELROLE, "*", [2]string{"repository", "pull"}, [2]string{"repository", "push"})
	creator := creatorPerms(roleCtl.LEVELROLE, "*", [2]string{"repository", "pull"})
	assert.False(t, isValidRolePermissionScope(creating, creator))
}

func TestIsValidRolePermissionScope_EmptyCreating(t *testing.T) {
	creator := creatorPerms(roleCtl.LEVELROLE, "*", [2]string{"repository", "pull"})
	assert.True(t, isValidRolePermissionScope(nil, creator))
}

func TestIsValidRolePermissionScope_ScopeNotInCreator(t *testing.T) {
	creating := rolePermissions(roleCtl.LEVELROLE, "myns", [2]string{"repository", "pull"})
	creator := creatorPerms(roleCtl.LEVELROLE, "otherns", [2]string{"repository", "pull"})
	assert.False(t, isValidRolePermissionScope(creating, creator))
}

func TestIsValidRolePermissionScope_WildcardNamespaceMatches(t *testing.T) {
	creating := rolePermissions(roleCtl.LEVELROLE, "specific", [2]string{"repository", "pull"})
	creator := creatorPerms(roleCtl.LEVELROLE, "*", [2]string{"repository", "pull"})
	assert.True(t, isValidRolePermissionScope(creating, creator))
}

// compile-time check: securityMock import used
var _ = testifymock.Anything
