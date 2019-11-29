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

package rbac

import (
	"testing"
)

type role struct {
	RoleName string
}

func (r *role) GetRoleName() string {
	return r.RoleName
}

func (r *role) GetPolicies() []*Policy {
	return []*Policy{
		{Resource: "/project", Action: "create"},
		{Resource: "/project", Action: "update"},
	}
}

type userWithRoles struct {
	Username string
	RoleName string
	BaseUser
}

func (u *userWithRoles) GetUserName() string {
	return u.Username
}

func (u *userWithRoles) GetRoles() []Role {
	return []Role{
		&role{RoleName: u.RoleName},
	}
}

type userWithoutRoles struct {
	Username     string
	UserPolicies []*Policy
	BaseUser
}

func (u *userWithoutRoles) GetUserName() string {
	return u.Username
}

func (u *userWithoutRoles) GetPolicies() []*Policy {
	return u.UserPolicies
}

// HasPermission ...
func HasPermission(user User, resource Resource, action Action) bool {
	return NewUserEvaluator(user).HasPermission(resource, action)
}

func TestHasPermissionUserWithRoles(t *testing.T) {
	type args struct {
		user     User
		resource Resource
		action   Action
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "project create for project admin",
			args: args{
				&userWithRoles{Username: "project admin", RoleName: "projectAdmin"},
				"/project",
				"create",
			},
			want: true,
		},
		{
			name: "project update for project admin",
			args: args{
				&userWithRoles{Username: "project admin", RoleName: "projectAdmin"},
				"/project",
				"update",
			},
			want: true,
		},
		{
			name: "project delete for project admin",
			args: args{
				&userWithRoles{Username: "project admin", RoleName: "projectAdmin"},
				"/project",
				"delete",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPermission(tt.args.user, tt.args.resource, tt.args.action); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermissionUsernameEmpty(t *testing.T) {

	type args struct {
		user     User
		resource Resource
		action   Action
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "project create for user without roles",
			args: args{
				&userWithoutRoles{Username: "", UserPolicies: []*Policy{{Resource: "/project", Action: "create"}}},
				"/project",
				"create",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPermission(tt.args.user, tt.args.resource, tt.args.action); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermissionRoleNameEmpty(t *testing.T) {
	type args struct {
		user     User
		resource Resource
		action   Action
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "project create for project admin",
			args: args{
				&userWithRoles{Username: "project admin", RoleName: ""},
				"/project",
				"create",
			},
			want: false,
		},
		{
			name: "project update for project admin",
			args: args{
				&userWithRoles{Username: "project admin", RoleName: ""},
				"/project",
				"update",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPermission(tt.args.user, tt.args.resource, tt.args.action); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermissionResourceKeyMatch(t *testing.T) {

	type args struct {
		user     User
		resource Resource
		action   Action
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "project member create for resource key match",
			args: args{
				&userWithoutRoles{Username: "user1", UserPolicies: []*Policy{{Resource: "/project/1/*", Action: "*"}}},
				"/project/1/member",
				"create",
			},
			want: true,
		},
		{
			name: "project member create for resource key match",
			args: args{
				&userWithoutRoles{Username: "user1", UserPolicies: []*Policy{{Resource: "/project/:id/*", Action: "*"}}},
				"/project/1/member",
				"create",
			},
			want: true,
		},
		{
			name: "project repository create test for resource key match",
			args: args{
				&userWithoutRoles{Username: "user1", UserPolicies: []*Policy{{Resource: "/project/1/*", Action: "create"}}},
				"/project/1/repository",
				"create",
			},
			want: true,
		},
		{
			name: "project repository delete test for resource key match",
			args: args{
				&userWithoutRoles{Username: "user1", UserPolicies: []*Policy{{Resource: "/project/1/*", Action: "create"}}},
				"/project/1/repository",
				"delete",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPermission(tt.args.user, tt.args.resource, tt.args.action); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermissionPolicyDeny(t *testing.T) {

	type args struct {
		user     User
		resource Resource
		action   Action
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "project member create for resource deny",
			args: args{
				&userWithoutRoles{
					Username: "user1",
					UserPolicies: []*Policy{
						{Resource: "/project/1/*", Action: "*"},
						{Resource: "/project/1/member", Action: "create", Effect: EffectDeny},
					},
				},
				"/project/1/member",
				"create",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPermission(tt.args.user, tt.args.resource, tt.args.action); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}
