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

package local

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/stretchr/testify/assert"
)

func TestPATSecurityContext_Can_WithPullScope(t *testing.T) {
	scope := `[{"project_id":1,"project_name":"test-project","access":[{"resource":"/project/1/repository","actions":["pull"]}]}]`
	ctx := NewPATSecurityContext(&models.User{UserID: 1}, scope)

	ns := rbac_project.NewNamespace(1)
	resource := ns.Resource(rbac.ResourceRepository)

	assert.True(t, ctx.Can(nil, rbac.ActionPull, resource), "should allow pull")
	assert.False(t, ctx.Can(nil, rbac.ActionPush, resource), "should not allow push")
}

func TestPATSecurityContext_Can_WithPushAndPullScope(t *testing.T) {
	scope := `[{"project_id":1,"project_name":"test-project","access":[{"resource":"/project/1/repository","actions":["pull","push"]}]}]`
	ctx := NewPATSecurityContext(&models.User{UserID: 1}, scope)

	ns := rbac_project.NewNamespace(1)
	resource := ns.Resource(rbac.ResourceRepository)

	assert.True(t, ctx.Can(nil, rbac.ActionPull, resource), "should allow pull")
	assert.True(t, ctx.Can(nil, rbac.ActionPush, resource), "should allow push")
}

func TestPATSecurityContext_Can_WithWildcardScope(t *testing.T) {
	scope := `[{"project_id":1,"project_name":"test-project","access":[{"resource":"/project/1/repository","actions":["*"]}]}]`
	ctx := NewPATSecurityContext(&models.User{UserID: 1}, scope)

	ns := rbac_project.NewNamespace(1)
	resource := ns.Resource(rbac.ResourceRepository)

	assert.True(t, ctx.Can(nil, rbac.ActionPull, resource), "should allow pull with wildcard")
	assert.True(t, ctx.Can(nil, rbac.ActionPush, resource), "should allow push with wildcard")
	assert.True(t, ctx.Can(nil, rbac.ActionDelete, resource), "should allow delete with wildcard")
}

func TestPATSecurityContext_Can_EmptyScope(t *testing.T) {
	ctx := NewPATSecurityContext(&models.User{UserID: 1}, "[]")

	ns := rbac_project.NewNamespace(1)
	resource := ns.Resource(rbac.ResourceRepository)

	assert.False(t, ctx.Can(nil, rbac.ActionPull, resource), "should not allow pull with empty scope")
	assert.False(t, ctx.Can(nil, rbac.ActionPush, resource), "should not allow push with empty scope")
}

func TestPATSecurityContext_Can_WrongProject(t *testing.T) {
	scope := `[{"project_id":1,"project_name":"test-project","access":[{"resource":"/project/1/repository","actions":["pull","push"]}]}]`
	ctx := NewPATSecurityContext(&models.User{UserID: 1}, scope)

	ns := rbac_project.NewNamespace(2)
	resource := ns.Resource(rbac.ResourceRepository)

	assert.False(t, ctx.Can(nil, rbac.ActionPull, resource), "should not allow for wrong project")
}

func TestPATSecurityContext_IsSysAdmin(t *testing.T) {
	scope := `[{"project_id":1,"project_name":"test-project","access":[{"resource":"/project/1/repository","actions":["*"]}]}]`
	ctx := NewPATSecurityContext(&models.User{UserID: 1}, scope)
	assert.False(t, ctx.IsSysAdmin())

	ctxAdmin := NewPATSecurityContext(&models.User{UserID: 1, SysAdminFlag: true}, scope)
	assert.True(t, ctxAdmin.IsSysAdmin())
}

func TestPATSecurityContext_GetUsername(t *testing.T) {
	scope := `[{"project_id":1,"project_name":"test-project","access":[{"resource":"/project/1/repository","actions":["pull"]}]}]`
	ctx := NewPATSecurityContext(&models.User{Username: "testuser"}, scope)
	assert.Equal(t, "testuser", ctx.GetUsername())
}

func TestPATSecurityContext_IsAuthenticated(t *testing.T) {
	scope := `[{"project_id":1,"project_name":"test-project","access":[{"resource":"/project/1/repository","actions":["pull"]}]}]`
	ctx := NewPATSecurityContext(&models.User{Username: "testuser"}, scope)
	assert.True(t, ctx.IsAuthenticated())

	ctxNil := NewPATSecurityContext(nil, scope)
	assert.False(t, ctxNil.IsAuthenticated())
}

func TestPATSecurityContext_InvalidScope(t *testing.T) {
	scope := `invalid json`
	ctx := NewPATSecurityContext(&models.User{UserID: 1}, scope)

	ns := rbac_project.NewNamespace(1)
	resource := ns.Resource(rbac.ResourceRepository)

	assert.False(t, ctx.Can(nil, rbac.ActionPull, resource), "should not allow with invalid scope")
}