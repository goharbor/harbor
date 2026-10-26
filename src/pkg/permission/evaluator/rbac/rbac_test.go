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

	"github.com/goharbor/harbor/src/pkg/permission/types"
)

type fakeRole struct{ name, resource string }

func (r fakeRole) GetRoleName() string { return r.name }
func (r fakeRole) GetPolicies() []*types.Policy {
	return []*types.Policy{{Resource: types.Resource(r.resource), Action: "pull"}}
}

type fakeUser struct {
	name string
	role fakeRole
}

func (u fakeUser) GetUserName() string          { return u.name }
func (u fakeUser) GetPolicies() []*types.Policy { return nil }
func (u fakeUser) GetRoles() []types.RBACRole   { return []types.RBACRole{u.role} }

func TestEvaluatorUsesHarborKeyMatch2(t *testing.T) {
	store.Purge()
	e := New(fakeUser{name: "alice", role: fakeRole{name: "dev", resource: "/project/:pid/repository"}})
	assert.True(t, e.HasPermission(context.TODO(), "/project/1/repository", "pull"))

	_, cached := store.entries.Load("/project/:pid/repository")
	assert.True(t, cached, "enforcer did not call Harbor's keyMatch2")
}

func TestEvaluatorNamesWithCSVSpecialChars(t *testing.T) {
	for _, name := range []string{`team"ops`, "a,b", "a#b"} {
		e := New(fakeUser{name: "user " + name, role: fakeRole{name: "role " + name, resource: "/project/1/repository"}})
		assert.True(t, e.HasPermission(context.TODO(), "/project/1/repository", "pull"), name)
		assert.False(t, e.HasPermission(context.TODO(), "/project/1/repository", "push"), name)
	}
}
