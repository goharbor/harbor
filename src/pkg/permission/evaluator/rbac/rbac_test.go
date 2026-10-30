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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/permission/policy"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// fakeRole is a membership in project 1 held under the store's role id 1.
type fakeRole struct{ name string }

func (r fakeRole) GetRoleName() string          { return r.name }
func (r fakeRole) GetPolicies() []*types.Policy { return nil }
func (r fakeRole) PolicyObject(resource types.Resource) (types.Resource, bool) {
	rest, err := resource.RelativeTo("/project/1")
	if err != nil {
		return "", false
	}
	return types.Resource(policy.Object(rest.String())), true
}

type fakeUser struct {
	name     string
	policies []*types.Policy
	roles    []types.RBACRole
}

func (u fakeUser) GetUserName() string          { return u.name }
func (u fakeUser) GetPolicies() []*types.Policy { return u.policies }
func (u fakeUser) GetRoles() []types.RBACRole   { return u.roles }

func pull(resource string) []*types.Policy {
	return []*types.Policy{{Resource: types.Resource(resource), Action: "pull"}}
}

// A direct grant written against a pattern only matches when the per-request
// enforcer registered Harbor's matcher under the name the model calls.
func TestDirectGrantsUseHarborKeyMatch2(t *testing.T) {
	e := New(fakeUser{name: "alice", policies: pull("/project/:pid/repository")})
	assert.True(t, e.HasPermission(context.TODO(), "/project/1/repository", "pull"))
	assert.False(t, e.HasPermission(context.TODO(), "/project/1/repository", "push"))
}

// casbin v3 parses a text policy line as CSV, so names holding quotes or
// commas only survive when the rules are loaded as arrays.
func TestDirectGrantsNamesWithCSVSpecialChars(t *testing.T) {
	for _, name := range []string{`team"ops`, "a,b", "a#b"} {
		e := New(fakeUser{name: "user " + name, policies: pull("/project/1/repository")})
		assert.True(t, e.HasPermission(context.TODO(), "/project/1/repository", "pull"), name)
		assert.False(t, e.HasPermission(context.TODO(), "/project/1/repository", "push"), name)
	}
}

// Role grants are answered by the store, in the role's own object space.
func TestRoleGrantsComeFromTheStore(t *testing.T) {
	store, err := policy.NewInMemory([]policy.Grant{{RoleID: 1, Resource: "repository", Action: "pull"}})
	require.NoError(t, err)
	defer store.Close()

	e := &Evaluator{rbacUser: fakeUser{name: "bob", roles: []types.RBACRole{fakeRole{name: policy.Subject(1)}}}, store: store}
	assert.True(t, e.HasPermission(context.TODO(), "/project/1/repository", "pull"))
	assert.False(t, e.HasPermission(context.TODO(), "/project/1/repository", "push"))
	assert.False(t, e.HasPermission(context.TODO(), "/project/2/repository", "pull"), "another project is out of the role's reach")
}
