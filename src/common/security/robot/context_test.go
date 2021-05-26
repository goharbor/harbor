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
	"github.com/goharbor/harbor/src/common/rbac/project"
	"reflect"
	"testing"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/assert"
)

var (
	private = &proModels.Project{
		Name:    "testrobot",
		OwnerID: 1,
	}
)

func TestIsAuthenticated(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, false, nil)
	assert.False(t, ctx.IsAuthenticated())

	// authenticated
	ctx = NewSecurityContext(&model.Robot{
		Name:     "test",
		Disabled: false,
	}, false, nil)
	assert.True(t, ctx.IsAuthenticated())
}

func TestGetUsername(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, false, nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&model.Robot{
		Name:     "test",
		Disabled: false,
	}, false, nil)
	assert.Equal(t, "test", ctx.GetUsername())
}

func TestIsSysAdmin(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, false, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, non admin
	ctx = NewSecurityContext(&model.Robot{
		Name:     "test",
		Disabled: false,
	}, false, nil)
	assert.False(t, ctx.IsSysAdmin())
}

func TestIsSolutionUser(t *testing.T) {
	ctx := NewSecurityContext(nil, false, nil)
	assert.False(t, ctx.IsSolutionUser())
}

func TestHasPullPerm(t *testing.T) {
	policies := []*types.Policy{
		{
			Resource: rbac.Resource(fmt.Sprintf("/project/%d/repository", private.ProjectID)),
			Action:   rbac.ActionPull,
		},
	}
	robot := &model.Robot{
		Name:        "test_robot_1",
		Description: "desc",
	}

	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(private, nil)

	ctx := NewSecurityContext(robot, false, policies)
	ctx.ctl = ctl
	resource := project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionPull, resource))
}

func TestHasPushPerm(t *testing.T) {
	policies := []*types.Policy{
		{
			Resource: rbac.Resource(fmt.Sprintf("/project/%d/repository", private.ProjectID)),
			Action:   rbac.ActionPush,
		},
	}
	robot := &model.Robot{
		Name:        "test_robot_2",
		Description: "desc",
	}

	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(private, nil)

	ctx := NewSecurityContext(robot, false, policies)
	ctx.ctl = ctl
	resource := project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource))
}

func TestHasPushPullPerm(t *testing.T) {
	policies := []*types.Policy{
		{
			Resource: rbac.Resource(fmt.Sprintf("/project/%d/repository", private.ProjectID)),
			Action:   rbac.ActionPush,
		},
		{
			Resource: rbac.Resource(fmt.Sprintf("/project/%d/repository", private.ProjectID)),
			Action:   rbac.ActionPull,
		},
	}
	robot := &model.Robot{
		Name:        "test_robot_3",
		Description: "desc",
	}

	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(private, nil)

	ctx := NewSecurityContext(robot, false, policies)
	ctx.ctl = ctl
	resource := project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource) && ctx.Can(context.TODO(), rbac.ActionPull, resource))
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterRobotPolicies(tt.args.p, tt.args.policies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterRobotPolicies() = %v, want %v", got, tt.want)
			}
		})
	}
}
