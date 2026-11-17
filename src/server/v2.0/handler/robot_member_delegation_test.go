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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common"
	commonModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security/local"
	robotSecurity "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/controller/member"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/controller/role"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	projectModels "github.com/goharbor/harbor/src/pkg/project/models"
	robotModels "github.com/goharbor/harbor/src/pkg/robot/model"
	roleManager "github.com/goharbor/harbor/src/pkg/role"
	roleModels "github.com/goharbor/harbor/src/pkg/role/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	memberop "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/member"
	robotop "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/robot"
	projectMock "github.com/goharbor/harbor/src/testing/controller/project"
	robotMock "github.com/goharbor/harbor/src/testing/controller/robot"
	projectManagerMock "github.com/goharbor/harbor/src/testing/pkg/project"
	roleManagerMock "github.com/goharbor/harbor/src/testing/pkg/role"
)

func setupMemberDelegationProject(t *testing.T) *projectMock.Controller {
	t.Helper()
	pc := &projectMock.Controller{}
	oldProject, oldBase, oldRole := project.Ctl, baseProjectCtl, role.Ctl
	project.Ctl, baseProjectCtl = pc, pc
	t.Cleanup(func() {
		project.Ctl, baseProjectCtl, role.Ctl = oldProject, oldBase, oldRole
	})
	p := &projectModels.Project{ProjectID: testProjectID, Name: "library"}
	pc.On("Get", testifymock.Anything, testProjectID, testifymock.Anything).Return(p, nil)
	pc.On("Get", testifymock.Anything, "library", testifymock.Anything).Return(p, nil)
	pc.On("GetByName", testifymock.Anything, "library").Return(p, nil)
	rc := &stubRoleCtl{}
	rc.On("Get", testifymock.Anything, int64(100), testifymock.Anything).Return(&role.Role{
		Role: roleModels.Role{ID: 100, Name: "member-manager"},
		Permissions: []*role.Permission{{Access: []*types.Policy{
			{Resource: rbac.ResourceMember, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceMember, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceMember, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionUpdate},
		}}},
	}, nil)
	role.Ctl = rc
	return pc
}

func TestHumanRobotMemberDelegation(t *testing.T) {
	for _, action := range []string{"create", "update"} {
		for _, endpoint := range []string{"create robot", "update robot", "refresh secret"} {
			for _, caller := range []struct {
				name     string
				sysAdmin bool
				roles    []int
				allowed  bool
			}{
				{"restricted custom role", false, []int{100}, false},
				{"administrator in another project", false, []int{100}, false},
				{"project administrator", false, []int{100, common.RoleProjectAdmin}, true},
				{"system administrator", true, nil, true},
			} {
				t.Run(action+"/"+endpoint+"/"+caller.name, func(t *testing.T) {
					pc := setupMemberDelegationProject(t)
					user := &commonModels.User{UserID: 7, Username: "delegator", SysAdminFlag: caller.sysAdmin}
					pc.On("ListRoles", testifymock.Anything, testProjectID, user).Return(caller.roles, nil)
					if caller.name == "administrator in another project" {
						pc.On("ListRoles", testifymock.Anything, testProjectID+1, user).Return([]int{common.RoleProjectAdmin}, nil)
					}
					ctx := newCtxWithSecurity(local.NewSecurityContext(user))
					permissions := robotPerm("library", "member", action)
					permissions[0].Kind = robot.LEVELPROJECT
					if endpoint == "create robot" && caller.allowed {
						// Successful creation calls SetProject with a concrete DB manager.
						// Check its authorization here; rejection cases exercise the handler.
						api := &robotAPI{}
						assert.NoError(t, api.requireAccess(ctx, &robot.Robot{Level: robot.LEVELPROJECT, ProjectID: testProjectID}, rbac.ActionCreate))
						assert.NoError(t, api.validateNoEscalation(ctx, permissions))
						return
					}
					r := &robot.Robot{
						Robot: robotModels.Robot{ID: 7, Name: "member-updater", ProjectID: testProjectID, Duration: -1, Salt: "salt"},
						Level: robot.LEVELPROJECT, Editable: true,
						Permissions: []*robot.Permission{{
							Kind: robot.LEVELPROJECT, Namespace: "library", Scope: fmt.Sprintf("/project/%d", testProjectID),
							Access: policyAccess(rbac.ResourceMember, rbac.Action(action)),
						}},
					}
					rc := &robotMock.Controller{}
					if endpoint != "create robot" {
						rc.On("Get", testifymock.Anything, int64(7), testifymock.Anything).Return(r, nil).Once()
					}
					if caller.allowed {
						rc.On("Update", testifymock.Anything, r, testifymock.Anything).Return(nil).Once()
					}
					api := &robotAPI{robotCtl: rc}
					rr := httptest.NewRecorder()
					switch endpoint {
					case "create robot":
						api.CreateRobot(ctx, robotop.CreateRobotParams{
							HTTPRequest: httptest.NewRequest(http.MethodPost, "/api/v2.0/robots", nil),
							Robot:       &models.RobotCreate{Name: r.Name, Level: r.Level, Duration: -1, Permissions: permissions},
						}).WriteResponse(rr, runtime.JSONProducer())
					case "update robot":
						api.UpdateRobot(ctx, robotop.UpdateRobotParams{RobotID: 7,
							Robot: &models.Robot{Name: r.Name, Level: r.Level, Permissions: permissions},
						}).WriteResponse(rr, runtime.JSONProducer())
					case "refresh secret":
						api.RefreshSec(ctx, robotop.RefreshSecParams{RobotID: 7, RobotSec: &models.RobotSec{Secret: "NewTestSecret1"}}).
							WriteResponse(rr, runtime.JSONProducer())
					}
					want := http.StatusForbidden
					if caller.allowed {
						want = http.StatusOK
					} else {
						assert.Contains(t, rr.Body.String(), "only system or project administrators")
						rc.AssertNotCalled(t, "Create", testifymock.Anything, testifymock.Anything)
						rc.AssertNotCalled(t, "Update", testifymock.Anything, testifymock.Anything, testifymock.Anything)
					}
					assert.Equal(t, want, rr.Code, rr.Body.String())
					rc.AssertExpectations(t)
				})
			}
		}
	}
}

func TestRobotUpdateCannotExpandMemberPermissions(t *testing.T) {
	for _, hasMemberUpdate := range []bool{false, true} {
		t.Run(fmt.Sprintf("has-member-update-%t", hasMemberUpdate), func(t *testing.T) {
			setupMemberDelegationProject(t)
			access := policyAccess(rbac.ResourceRobot, rbac.ActionUpdate)
			if hasMemberUpdate {
				access = append(access, policyAccess(rbac.ResourceMember, rbac.ActionUpdate)...)
			}
			// Stored grants may use explicit allow while the update request omits effect.
			for _, policy := range access {
				policy.Effect = types.EffectAllow
			}
			ctx := newCtxWithSecurity(robotSecurity.NewSecurityContext(&robot.Robot{
				Robot:       robotModels.Robot{ID: 1, Name: "robot$library+updater", ProjectID: testProjectID},
				Level:       robot.LEVELPROJECT,
				Permissions: []*robot.Permission{{Kind: robot.LEVELPROJECT, Namespace: "library", Scope: fmt.Sprintf("/project/%d", testProjectID), Access: access}},
			}))
			r := &robot.Robot{Robot: robotModels.Robot{ID: 2, Name: "target", ProjectID: testProjectID, Duration: -1}, Level: robot.LEVELPROJECT}
			permissions := robotPerm("library", "member", "update")
			permissions[0].Kind = robot.LEVELPROJECT
			rc := &robotMock.Controller{}
			if hasMemberUpdate {
				rc.On("Update", testifymock.Anything, r, testifymock.Anything).Return(nil).Once()
			}
			err := (&robotAPI{robotCtl: rc}).updateV2Robot(ctx, robotop.UpdateRobotParams{
				RobotID: 2, Robot: &models.Robot{Name: r.Name, Level: r.Level, Permissions: permissions},
			}, r)
			if hasMemberUpdate {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, errors.DENIED, errors.ErrCode(err))
				rc.AssertNotCalled(t, "Update", testifymock.Anything, testifymock.Anything, testifymock.Anything)
			}
			rc.AssertExpectations(t)
		})
	}
}

func TestHumanWithRobotNameStillCannotEscalate(t *testing.T) {
	pc := setupMemberDelegationProject(t)
	user := &commonModels.User{UserID: 7, Username: "robot$library+member-updater"}
	pc.On("ListRoles", testifymock.Anything, testProjectID, user).Return([]int{100}, nil)
	ctx := newCtxWithSecurity(local.NewSecurityContext(user))
	xFalse := false
	rr := httptest.NewRecorder()
	(&memberAPI{}).UpdateProjectMember(ctx, memberop.UpdateProjectMemberParams{
		ProjectNameOrID: fmt.Sprint(testProjectID), XIsResourceName: &xFalse, Mid: 2,
		Role: &models.RoleRequest{RoleID: common.RoleLimitedGuest},
	}).WriteResponse(rr, runtime.JSONProducer())
	assert.Equal(t, http.StatusForbidden, rr.Code, rr.Body.String())
	assert.Contains(t, rr.Body.String(), "escalation not allowed")
}

func TestRobotMemberDelegationRequiresAdminForEveryProject(t *testing.T) {
	pc := setupMemberDelegationProject(t)
	user := &commonModels.User{UserID: 7, Username: "delegator"}
	pc.On("ListRoles", testifymock.Anything, testProjectID, user).Return([]int{common.RoleProjectAdmin}, nil)
	pc.On("GetByName", testifymock.Anything, "other").Return(&projectModels.Project{ProjectID: testProjectID + 1}, nil)
	pc.On("ListRoles", testifymock.Anything, testProjectID+1, user).Return([]int{100}, nil)
	ctx := newCtxWithSecurity(local.NewSecurityContext(user))
	permissions := append(robotPerm("library", "member", "update"), robotPerm("other", "member", "create")...)
	err := (&robotAPI{}).validateNoEscalation(ctx, permissions)
	assert.Equal(t, errors.ForbiddenCode, errors.ErrCode(err))
	assert.Contains(t, err.Error(), "project other")
}

func TestRobotMemberDelegationRoleLookupFailure(t *testing.T) {
	pc := setupMemberDelegationProject(t)
	user := &commonModels.User{UserID: 7, Username: "delegator"}
	lookupErr := fmt.Errorf("role lookup failed")
	pc.On("ListRoles", testifymock.Anything, testProjectID, user).Return(nil, lookupErr)
	err := (&robotAPI{}).validateNoEscalation(newCtxWithSecurity(local.NewSecurityContext(user)), robotPerm("library", "member", "update"))
	assert.ErrorIs(t, err, lookupErr)
}

func TestRobotMemberReadDelegationDoesNotRequireAdmin(t *testing.T) {
	pc := setupMemberDelegationProject(t)
	user := &commonModels.User{UserID: 7, Username: "delegator"}
	pc.On("ListRoles", testifymock.Anything, testProjectID, user).Return([]int{100}, nil)
	err := (&robotAPI{}).validateNoEscalation(newCtxWithSecurity(local.NewSecurityContext(user)), robotPerm("library", "member", "read"))
	assert.NoError(t, err)
}

func TestRobotMemberDelegationWildcardRequiresSystemAdmin(t *testing.T) {
	for _, sysAdmin := range []bool{false, true} {
		t.Run(fmt.Sprint(sysAdmin), func(t *testing.T) {
			ctx := newCtxWithSecurity(local.NewSecurityContext(&commonModels.User{Username: "delegator", SysAdminFlag: sysAdmin}))
			err := (&robotAPI{}).requireMemberDelegation(ctx, "*")
			if sysAdmin {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, errors.ForbiddenCode, errors.ErrCode(err))
			}
		})
	}
}

func TestRobotMemberAssignmentRequiresSecurityContext(t *testing.T) {
	err := (&memberAPI{}).checkNoEscalation(context.Background(), testProjectID, common.RoleLimitedGuest)
	assert.Equal(t, errors.UnAuthorizedCode, errors.ErrCode(err))
}

func TestRobotMemberUpdateStillValidatesRole(t *testing.T) {
	setupMemberDelegationProject(t)
	oldProject, oldRole := pkg.ProjectMgr, roleManager.Mgr
	t.Cleanup(func() { pkg.ProjectMgr, roleManager.Mgr = oldProject, oldRole })
	pm := &projectManagerMock.Manager{}
	pm.On("Get", testifymock.Anything, testProjectID).Return(&projectModels.Project{ProjectID: testProjectID}, nil)
	rm := &roleManagerMock.Manager{}
	rm.On("Get", testifymock.Anything, int64(100)).Return(nil, errors.NotFoundError(nil))
	pkg.ProjectMgr, roleManager.Mgr = pm, rm
	ctx := newCtxWithSecurity(robotSecurity.NewSecurityContext(&robot.Robot{
		Robot:       robotModels.Robot{Name: "robot$library+member-updater", ProjectID: testProjectID},
		Level:       robot.LEVELPROJECT,
		Permissions: []*robot.Permission{{Scope: fmt.Sprintf("/project/%d", testProjectID), Access: policyAccess(rbac.ResourceMember, rbac.ActionUpdate)}},
	}))
	api := &memberAPI{ctl: member.NewController()}
	for _, roleID := range []int64{0, -1, 100} {
		t.Run(fmt.Sprint(roleID), func(t *testing.T) {
			xFalse := false
			rr := httptest.NewRecorder()
			api.UpdateProjectMember(ctx, memberop.UpdateProjectMemberParams{
				ProjectNameOrID: fmt.Sprint(testProjectID), XIsResourceName: &xFalse, Mid: 2,
				Role: &models.RoleRequest{RoleID: roleID},
			}).WriteResponse(rr, runtime.JSONProducer())
			assert.Equal(t, http.StatusBadRequest, rr.Code, rr.Body.String())
			assert.Contains(t, rr.Body.String(), "invalid role")
		})
	}
}
