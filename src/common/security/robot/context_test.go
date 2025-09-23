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

package robot

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/rbac/system"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
)

var (
	private = &proModels.Project{
		ProjectID: 1,
		Name:      "testrobot",
		OwnerID:   1,
	}
)

func TestIsAuthenticated(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil)
	assert.False(t, ctx.IsAuthenticated())

	// authenticated
	ctx = NewSecurityContext(&robot.Robot{
		Robot: model.Robot{
			Name:     "test",
			Disabled: false,
		},
	})
	assert.True(t, ctx.IsAuthenticated())
}

func TestGetUsername(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&robot.Robot{
		Robot: model.Robot{
			Name:     "test",
			Disabled: false,
		},
	})
	assert.Equal(t, "test", ctx.GetUsername())
}

func TestGetUser(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&robot.Robot{
		Robot: model.Robot{
			ID:       123,
			Name:     "test",
			Disabled: false,
		},
	})
	assert.Equal(t, "test", ctx.User().Name)
	assert.Equal(t, int64(123), ctx.User().ID)
}

func TestIsSysAdmin(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, non admin
	ctx = NewSecurityContext(&robot.Robot{
		Robot: model.Robot{
			Name:     "test",
			Disabled: false,
		},
	})
	assert.False(t, ctx.IsSysAdmin())
}

func TestIsSolutionUser(t *testing.T) {
	ctx := NewSecurityContext(nil)
	assert.False(t, ctx.IsSolutionUser())
}

func TestHasPullPerm(t *testing.T) {
	robot := &robot.Robot{
		Level: "project",
		Robot: model.Robot{
			Name:        "test_robot_1",
			Description: "desc",
		},
		Permissions: []*robot.Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
						Action:   rbac.ActionPull,
						Effect:   "allow",
					},
				},
			},
		},
	}

	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(private, nil)

	ctx := NewSecurityContext(robot)
	ctx.ctl = ctl
	resource := project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionPull, resource))
}

func TestHasPushPerm(t *testing.T) {
	robot := &robot.Robot{
		Level: "project",
		Robot: model.Robot{
			Name:     "test",
			Disabled: false,
		},
		Permissions: []*robot.Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
						Action:   rbac.ActionPush,
						Effect:   "allow",
					},
				},
			},
		},
	}

	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(private, nil)

	ctx := NewSecurityContext(robot)
	ctx.ctl = ctl
	resource := project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource))
}

func TestHasPushPullPerm(t *testing.T) {
	robot := &robot.Robot{
		Level: "project",
		Robot: model.Robot{
			Name:        "test_robot_3",
			Description: "desc",
		},
		Permissions: []*robot.Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
						Action:   rbac.ActionPush,
						Effect:   "allow",
					},
					{
						Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
						Action:   rbac.ActionPull,
						Effect:   "allow",
					},
				},
			},
		},
	}

	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(private, nil)

	ctx := NewSecurityContext(robot)
	ctx.ctl = ctl
	resource := project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource) && ctx.Can(context.TODO(), rbac.ActionPull, resource))
}

func TestSysAndProPerm(t *testing.T) {
	robot := &robot.Robot{
		Level: "system",
		Robot: model.Robot{
			Name:        "test_robot_4",
			Description: "desc",
		},
		Permissions: []*robot.Permission{
			{
				Kind:      "system",
				Namespace: "/",
				Access: []*types.Policy{
					{
						Resource: rbac.Resource(fmt.Sprintf("system/%s", rbac.ResourceRepository)),
						Action:   rbac.ActionList,
						Effect:   "allow",
					},
					{
						Resource: rbac.Resource(fmt.Sprintf("system/%s", rbac.ResourceGarbageCollection)),
						Action:   rbac.ActionCreate,
						Effect:   "allow",
					},
				},
			},
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
						Action:   rbac.ActionPush,
						Effect:   "allow",
					},
					{
						Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
						Action:   rbac.ActionPull,
						Effect:   "allow",
					},
				},
			},
		},
	}

	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(private, nil)

	ctx := NewSecurityContext(robot)
	ctx.ctl = ctl
	resource := project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource) && ctx.Can(context.TODO(), rbac.ActionPull, resource))

	resource = system.NewNamespace().Resource(rbac.ResourceGarbageCollection)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionCreate, resource))
}

func Test_filterRobotPolicies(t *testing.T) {
	type args struct {
		p        *proModels.Project
		policies []*types.Policy
	}
	tests := []struct {
		name string
		args args
		want []*types.Policy
	}{
		{
			"policies of one project",
			args{
				&proModels.Project{ProjectID: 1},
				[]*types.Policy{
					{Resource: "/project/1/repository", Action: "pull", Effect: "allow"},
				},
			},
			[]*types.Policy{
				{Resource: "/project/1/repository", Action: "pull", Effect: "allow"},
			},
		},
		{
			"policies of multi projects",
			args{
				&proModels.Project{ProjectID: 1},
				[]*types.Policy{
					{Resource: "/project/1/repository", Action: "pull", Effect: "allow"},
					{Resource: "/project/2/repository", Action: "pull", Effect: "allow"},
				},
			},
			[]*types.Policy{
				{Resource: "/project/1/repository", Action: "pull", Effect: "allow"},
			},
		},
		{
			"wildcard permissions should apply to any project and generate auto-pull permissions",
			args{
				&proModels.Project{ProjectID: 1},
				[]*types.Policy{
					{Resource: "/project/*/repository", Action: "push", Effect: "allow"},
					{Resource: "/project/2/repository", Action: "pull", Effect: "allow"},
				},
			},
			[]*types.Policy{
				{Resource: "/project/*/repository", Action: "push", Effect: "allow"},
				{Resource: "/project/*/repository", Action: "pull", Effect: "allow"},
			},
		},
		{
			"wildcard permissions should work for library project specifically",
			args{
				&proModels.Project{ProjectID: 1, Name: "library"},
				[]*types.Policy{
					{Resource: "/project/*/repository", Action: "push", Effect: "allow"},
					{Resource: "/project/*/repository", Action: "delete", Effect: "allow"},
				},
			},
			[]*types.Policy{
				{Resource: "/project/*/repository", Action: "push", Effect: "allow"},
				{Resource: "/project/*/repository", Action: "pull", Effect: "allow"},
				{Resource: "/project/*/repository", Action: "delete", Effect: "allow"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterRobotPolicies(tt.args.p, tt.args.policies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterRobotPolicies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPolicyResource(t *testing.T) {
	type args struct {
		perm *robot.Permission
		poli *types.Policy
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"project resource",
			args{
				&robot.Permission{
					Kind:      "project",
					Namespace: "library",
					Access: []*types.Policy{
						{
							Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
							Action:   rbac.ActionPush,
						},
						{
							Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
							Action:   rbac.ActionPull,
						},
					},
					Scope: fmt.Sprintf("/project/%d", private.ProjectID),
				},
				&types.Policy{Resource: "project", Action: "pull", Effect: "allow"},
			},
			fmt.Sprintf("/project/%d", private.ProjectID),
		},
		{
			"project resource",
			args{
				&robot.Permission{
					Kind:      "project",
					Namespace: "library",
					Access: []*types.Policy{
						{
							Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
							Action:   rbac.ActionPush,
						},
						{
							Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
							Action:   rbac.ActionPull,
						},
					},
					Scope: fmt.Sprintf("/project/%d", private.ProjectID),
				},
				&types.Policy{Resource: "repository", Action: "get", Effect: "allow"},
			},
			fmt.Sprintf("/project/%d/repository", private.ProjectID),
		},
		{
			"system resource",
			args{
				&robot.Permission{
					Kind:      "project",
					Namespace: "library",
					Access: []*types.Policy{
						{
							Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
							Action:   rbac.ActionPush,
						},
						{
							Resource: rbac.Resource(fmt.Sprintf("project/%d/repository", private.ProjectID)),
							Action:   rbac.ActionPull,
						},
					},
					Scope: "/system",
				},
				&types.Policy{Resource: "repository", Action: "get", Effect: "allow"},
			},
			"/system/repository",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPolicyResource(tt.args.perm, tt.args.poli); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPolicyResource() = %v, want %v", got, tt.want)
			}
		})
	}
}
