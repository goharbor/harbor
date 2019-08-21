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

package project

import (
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/stretchr/testify/suite"
)

type fakeVisitorContext struct {
	username   string
	isSysAdmin bool
}

func (ctx *fakeVisitorContext) IsAuthenticated() bool {
	return ctx.username != ""
}

func (ctx *fakeVisitorContext) GetUsername() string {
	return ctx.username
}

func (ctx *fakeVisitorContext) IsSysAdmin() bool {
	return ctx.IsAuthenticated() && ctx.isSysAdmin
}

var (
	anonymousCtx     = &fakeVisitorContext{}
	authenticatedCtx = &fakeVisitorContext{username: "user"}
	sysAdminCtx      = &fakeVisitorContext{username: "admin", isSysAdmin: true}
)

type VisitorTestSuite struct {
	suite.Suite
}

func (suite *VisitorTestSuite) TestGetPolicies() {
	namespace := rbac.NewProjectNamespace(1, false)
	publicNamespace := rbac.NewProjectNamespace(1, true)

	anonymous := NewUser(anonymousCtx, namespace)
	suite.Nil(anonymous.GetPolicies())

	anonymousForPublicProject := NewUser(anonymousCtx, publicNamespace)
	suite.Equal(anonymousForPublicProject.GetPolicies(), PoliciesForPublicProject(publicNamespace))

	authenticated := NewUser(authenticatedCtx, namespace)
	suite.Nil(authenticated.GetPolicies())

	authenticatedForPublicProject := NewUser(authenticatedCtx, publicNamespace)
	suite.Equal(authenticatedForPublicProject.GetPolicies(), PoliciesForPublicProject(publicNamespace))

	systemAdmin := NewUser(sysAdminCtx, namespace)
	suite.Equal(systemAdmin.GetPolicies(), GetAllPolicies(namespace))

	systemAdminForPublicProject := NewUser(sysAdminCtx, publicNamespace)
	suite.Equal(systemAdminForPublicProject.GetPolicies(), GetAllPolicies(publicNamespace))
}

func (suite *VisitorTestSuite) TestGetRoles() {
	namespace := rbac.NewProjectNamespace(1, false)

	anonymous := NewUser(anonymousCtx, namespace)
	suite.Nil(anonymous.GetRoles())

	authenticated := NewUser(authenticatedCtx, namespace)
	suite.Empty(authenticated.GetRoles())

	authenticated = NewUser(authenticatedCtx, namespace, common.RoleProjectAdmin)
	suite.Len(authenticated.GetRoles(), 1)

	authenticated = NewUser(authenticatedCtx, namespace, common.RoleProjectAdmin, common.RoleDeveloper)
	suite.Len(authenticated.GetRoles(), 2)
}

func TestVisitorTestSuite(t *testing.T) {
	suite.Run(t, new(VisitorTestSuite))
}
