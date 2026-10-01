package handler

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common/rbac"
	rbacProject "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	projectModels "github.com/goharbor/harbor/src/pkg/project/models"
	robotModel "github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/robot"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	robottesting "github.com/goharbor/harbor/src/testing/controller/robot"
	"github.com/goharbor/harbor/src/testing/mock"
)

func TestValidLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{"project level true",
			"project",
			true,
		},
		{"system level true",
			"system",
			true,
		},
		{"unknown level false",
			"unknown",
			false,
		},
		{"systemproject level false",
			"systemproject",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isValidLevel(tt.level) != tt.expected {
				t.Errorf("name: %s, isValidLevel() = %#v, want %#v", tt.name, tt.level, tt.expected)
			}
		})
	}
}

func TestValidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		expected bool
	}{
		{"duration 0",
			0,
			false,
		},
		{"duration 1",
			1,
			true,
		},
		{"duration -1",
			-1,
			true,
		},
		{"duration -10",
			-10,
			false,
		},
		{"duration 9999",
			9999,
			true,
		},
		{"duration max",
			math.MaxInt32 - 1,
			true,
		},
		{"duration max",
			math.MaxInt32,
			false,
		},
		{"duration 999999999999",
			999999999999,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isValidDuration(tt.duration) != tt.expected {
				t.Errorf("name: %s, isValidLevel() = %#v, want %#v", tt.name, tt.duration, tt.expected)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		rname    string
		expected bool
	}{
		{"rname robotname",
			"robotname",
			true,
		},
		{"rname 123456",
			"123456",
			true,
		},
		{"rname robot123",
			"robot123",
			true,
		},
		{"rname ROBOT",
			"ROBOT",
			false,
		},
		{"rname robot+123",
			"robot+123",
			false,
		},
		{"rname robot$123",
			"robot$123",
			false,
		},
		{"rname robot_test123",
			"robot_test123",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.rname)
			if err != nil && tt.expected {
				t.Errorf("name: %s, validateName() = %#v, want %#v", tt.name, tt.rname, tt.expected)
			}
		})
	}
}

func TestValidateNilPermissionElement(t *testing.T) {
	rAPI := &robotAPI{}
	err := rAPI.validate(-1, robot.LEVELSYSTEM, []*models.RobotPermission{nil})
	assert.Error(t, err)
}

func TestValidateNilAccessElement(t *testing.T) {
	rAPI := &robotAPI{}
	err := rAPI.validate(-1, robot.LEVELSYSTEM, []*models.RobotPermission{
		{
			Kind:   robot.LEVELSYSTEM,
			Access: []*models.Access{nil},
		},
	})
	assert.Error(t, err)
}

func TestContainsAccess(t *testing.T) {
	system := rbac.PoliciesMap["System"]
	systests := []struct {
		name     string
		acc      *models.Access
		expected bool
	}{
		{"System ResourceRegistry push",
			&models.Access{
				Resource: rbac.ResourceRegistry.String(),
				Action:   rbac.ActionPush.String(),
			},
			false,
		},
		{"System ResourceProject delete",
			&models.Access{
				Resource: rbac.ResourceProject.String(),
				Action:   rbac.ActionDelete.String(),
			},
			false,
		},
		{"System ResourceReplicationPolicy delete",
			&models.Access{
				Resource: rbac.ResourceReplicationPolicy.String(),
				Action:   rbac.ActionDelete.String(),
			},
			true,
		},
	}
	for _, tt := range systests {
		t.Run(tt.name, func(t *testing.T) {
			ok := containsAccess(system, tt.acc)
			if ok != tt.expected {
				t.Errorf("name: %s, containsAccess() = %#v, want %#v", tt.name, tt.acc, tt.expected)
			}
		})
	}

	project := rbac.PoliciesMap["Project"]
	protests := []struct {
		name     string
		acc      *models.Access
		expected bool
	}{
		{"Project ResourceLog delete",
			&models.Access{
				Resource: rbac.ResourceLog.String(),
				Action:   rbac.ActionDelete.String(),
			},
			false,
		},
		{"Project ResourceMetadata read",
			&models.Access{
				Resource: rbac.ResourceMetadata.String(),
				Action:   rbac.ActionRead.String(),
			},
			true,
		},
		{"Project ResourceRobot create",
			&models.Access{
				Resource: rbac.ResourceRobot.String(),
				Action:   rbac.ActionCreate.String(),
			},
			false,
		},
	}
	for _, tt := range protests {
		t.Run(tt.name, func(t *testing.T) {
			ok := containsAccess(project, tt.acc)
			if ok != tt.expected {
				t.Errorf("name: %s, containsAccess() = %#v, want %#v", tt.name, tt.acc, tt.expected)
			}
		})
	}
}

func TestValidPermissionScope(t *testing.T) {
	tests := []struct {
		name          string
		creatingPerms []*models.RobotPermission
		creatorPerms  []*robot.Permission
		expected      bool
	}{
		{
			name: "Project - subset",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "project",
					Namespace: "testSubset",
					Access: []*models.Access{
						{Resource: "repository", Action: "pull", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "project",
					Namespace: "testSubset",
					Access: []*types.Policy{
						{Resource: "repository", Action: "pull", Effect: "allow"},
						{Resource: "repository", Action: "push", Effect: "allow"},
					},
				},
			},
			expected: true,
		},
		{
			name: "Project - not Subset",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "project",
					Namespace: "testNotSubset",
					Access: []*models.Access{
						{Resource: "repository", Action: "push", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "project",
					Namespace: "testNotSubset",
					Access: []*types.Policy{
						{Resource: "repository", Action: "pull", Effect: "allow"},
					},
				},
			},
			expected: false,
		},
		{
			name: "Project - equal",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "project",
					Namespace: "library",
					Access: []*models.Access{
						{Resource: "repository", Action: "pull", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "project",
					Namespace: "library",
					Access: []*types.Policy{
						{Resource: "repository", Action: "pull", Effect: "allow"},
					},
				},
			},
			expected: true,
		},
		{
			name: "Project - different",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "project",
					Namespace: "library",
					Access: []*models.Access{
						{Resource: "repository", Action: "pull", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "project",
					Namespace: "other",
					Access: []*types.Policy{
						{Resource: "repository", Action: "pull", Effect: "allow"},
					},
				},
			},
			expected: false,
		},
		{
			name: "Project - empty creator",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "project",
					Namespace: "library",
					Access: []*models.Access{
						{Resource: "repository", Action: "pull", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{},
			expected:     false,
		},
		{
			name:          "Project - empty creating",
			creatingPerms: []*models.RobotPermission{},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "project",
					Namespace: "library",
					Access: []*types.Policy{
						{Resource: "repository", Action: "pull", Effect: "allow"},
					},
				},
			},
			expected: true,
		},
		{
			name: "System - subset",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "system",
					Namespace: "admin",
					Access: []*models.Access{
						{Resource: "user", Action: "create", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "system",
					Namespace: "admin",
					Access: []*types.Policy{
						{Resource: "user", Action: "create", Effect: "allow"},
						{Resource: "user", Action: "delete", Effect: "allow"},
					},
				},
			},
			expected: true,
		},
		{
			name: "System - not subset",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "system",
					Namespace: "admin",
					Access: []*models.Access{
						{Resource: "user", Action: "delete", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "system",
					Namespace: "admin",
					Access: []*types.Policy{
						{Resource: "user", Action: "create", Effect: "allow"},
					},
				},
			},
			expected: false,
		},
		{
			name: "System - subset project",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "project",
					Namespace: "test1",
					Access: []*models.Access{
						{Resource: "user", Action: "delete", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "system",
					Namespace: "/",
					Access: []*types.Policy{
						{Resource: "robot", Action: "create", Effect: "allow"},
					},
				},
				{
					Kind:      "project",
					Namespace: "test1",
					Access: []*types.Policy{
						{Resource: "user", Action: "create", Effect: "allow"},
						{Resource: "user", Action: "delete", Effect: "allow"},
					},
				},
			},
			expected: true,
		},
		{
			name: "System - cover all",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "project",
					Namespace: "test1",
					Access: []*models.Access{
						{Resource: "user", Action: "delete", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "system",
					Namespace: "/",
					Access: []*types.Policy{
						{Resource: "robot", Action: "create", Effect: "allow"},
					},
				},
				{
					Kind:      "project",
					Namespace: "*",
					Access: []*types.Policy{
						{Resource: "user", Action: "create", Effect: "allow"},
						{Resource: "user", Action: "delete", Effect: "allow"},
					},
				},
			},
			expected: true,
		},
		{
			name: "System - cover all 2",
			creatingPerms: []*models.RobotPermission{
				{
					Kind:      "project",
					Namespace: "test1",
					Access: []*models.Access{
						{Resource: "user", Action: "update", Effect: "allow"},
					},
				},
			},
			creatorPerms: []*robot.Permission{
				{
					Kind:      "system",
					Namespace: "/",
					Access: []*types.Policy{
						{Resource: "robot", Action: "create", Effect: "allow"},
					},
				},
				{
					Kind:      "project",
					Namespace: "*",
					Access: []*types.Policy{
						{Resource: "user", Action: "create", Effect: "allow"},
						{Resource: "user", Action: "delete", Effect: "allow"},
					},
				},
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPermissionScope(tt.creatingPerms, tt.creatorPerms)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListRobotNonStringQueryValue(t *testing.T) {
	queries := []string{
		"Level=~system",
		"Level=[1~2]",
		"Level={a b}",
		"Level=(a b)",
		"Level=project,ProjectID=~1",
		"Level=project,ProjectID=[1~2]",
	}

	secCtx := &securitytesting.Context{}
	secCtx.On("IsAuthenticated").Return(true)
	ctx := security.NewContext(context.Background(), secCtx)

	for _, q := range queries {
		t.Run(q, func(t *testing.T) {
			responder := (&robotAPI{}).ListRobot(ctx, operation.ListRobotParams{Q: &q})

			rec := httptest.NewRecorder()
			responder.WriteResponse(rec, nil)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// mapRobotToHumanResource
// ---------------------------------------------------------------------------

func TestMapRobotToHumanResource_ProjectMapsToSelf(t *testing.T) {
	assert.Equal(t, rbac.ResourceSelf, mapRobotToHumanResource(rbac.ResourceProject))
}

func TestMapRobotToHumanResource_RepositoryPassthrough(t *testing.T) {
	assert.Equal(t, rbac.ResourceRepository, mapRobotToHumanResource(rbac.ResourceRepository))
}

func TestMapRobotToHumanResource_UnknownPassthrough(t *testing.T) {
	unknown := rbac.Resource("unknown")
	assert.Equal(t, unknown, mapRobotToHumanResource(unknown))
}

// ---------------------------------------------------------------------------
// validateNoEscalation
// ---------------------------------------------------------------------------

func robotPerm(ns, resource, action string) []*models.RobotPermission {
	return []*models.RobotPermission{{
		Namespace: ns,
		Access:    []*models.Access{{Resource: resource, Action: action}},
	}}
}

// stubProject stubs projectCtlMock.GetByName to return a project with the given ID.
// perm.Namespace is always a string so HasProjectPermission always calls GetByName.
func stubProject(projectID int64) {
	mock.OnAnything(projectCtlMock, "GetByName").
		Return(&projectModels.Project{ProjectID: projectID}, nil).Once()
}

func TestValidateNoEscalation_CallerHasPermission(t *testing.T) {
	stubProject(1)
	sc := &securitytesting.Context{}
	sc.On("Can", testifymock.Anything, rbac.ActionPull,
		rbacProject.NewNamespace(1).Resource(rbac.ResourceRepository)).Return(true)

	err := (&robotAPI{}).validateNoEscalation(newCtxWithSecurity(sc), robotPerm("testproject", "repository", "pull"))
	assert.NoError(t, err)
}

func TestValidateNoEscalation_CallerLacksPermission(t *testing.T) {
	stubProject(1)
	sc := &securitytesting.Context{}
	sc.On("Can", testifymock.Anything, rbac.ActionPush,
		rbacProject.NewNamespace(1).Resource(rbac.ResourceRepository)).Return(false)

	err := (&robotAPI{}).validateNoEscalation(newCtxWithSecurity(sc), robotPerm("testproject", "repository", "push"))
	assert.Error(t, err)
	assert.Equal(t, errors.ForbiddenCode, errors.ErrCode(err))
}

func TestValidateNoEscalation_EmptyPermissions(t *testing.T) {
	sc := &securitytesting.Context{}
	err := (&robotAPI{}).validateNoEscalation(newCtxWithSecurity(sc), nil)
	assert.NoError(t, err)
	sc.AssertNotCalled(t, "Can")
}

func TestValidateNoEscalation_ProjectResourceMappedToSelf(t *testing.T) {
	stubProject(1)
	sc := &securitytesting.Context{}
	// "project" maps to ResourceSelf ("") via mapRobotToHumanResource.
	sc.On("Can", testifymock.Anything, rbac.ActionRead,
		rbacProject.NewNamespace(1).Resource(rbac.ResourceSelf)).Return(true)

	err := (&robotAPI{}).validateNoEscalation(newCtxWithSecurity(sc), robotPerm("testproject", "project", "read"))
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// RefreshSec escalation guard (default case)
// ---------------------------------------------------------------------------

// RefreshSec hands the caller control of the robot, so it runs the same
// no-escalation switch as create/update. A security context that is neither a
// human (*local) nor a robot (*robot) must not fall through that switch and skip
// the check: the default case rejects it before the secret is refreshed. Without
// the default arm, an unknown context would reach robotCtl.Update unchecked.
//
// The human- and robot-caller escalation logic itself is covered directly by the
// TestValidateNoEscalation_* and TestValidPermissionScope cases above; this test
// covers the previously-unguarded default path.
func TestRefreshSec_UnknownSecurityContextRejected(t *testing.T) {
	sc := &securitytesting.Context{}
	sc.On("IsAuthenticated").Return(true)
	sc.On("IsSysAdmin").Return(false)
	// Caller may update robots, so requireAccess(ActionUpdate) passes...
	sc.On("Can", testifymock.Anything, rbac.ActionUpdate,
		projectResource(rbac.ResourceRobot)).Return(true)

	rc := &robottesting.Controller{}
	rc.On("Get", testifymock.Anything, int64(5), testifymock.Anything).Return(&robot.Robot{
		Robot: robotModel.Robot{ProjectID: testProjectID},
		Level: robot.LEVELPROJECT,
		Permissions: []*robot.Permission{{
			Kind:      "project",
			Namespace: "1",
			Access:    []*types.Policy{{Resource: rbac.ResourceRepository, Action: rbac.ActionPush}},
		}},
	}, nil)

	rAPI := &robotAPI{robotCtl: rc}
	params := operation.RefreshSecParams{RobotID: 5, RobotSec: &models.RobotSec{}}
	resp := rAPI.RefreshSec(newCtxWithSecurity(sc), params)

	rr := httptest.NewRecorder()
	resp.WriteResponse(rr, runtime.JSONProducer())
	assert.NotEqual(t, http.StatusOK, rr.Code, "unknown security context must be rejected")
	// The secret must never be refreshed for a context that skipped the escalation check.
	rc.AssertNotCalled(t, "Update", testifymock.Anything, testifymock.Anything, testifymock.Anything)
}
