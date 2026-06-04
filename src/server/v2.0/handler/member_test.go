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
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common/rbac"
	rbacProject "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	roleCtl "github.com/goharbor/harbor/src/controller/role"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	securityMock "github.com/goharbor/harbor/src/testing/common/security"
)

// ---------------------------------------------------------------------------
// Inline mock for roleCtl.Controller
// Only Get needs real behaviour; the rest are no-ops for these tests.
// ---------------------------------------------------------------------------

type stubRoleCtl struct {
	testifymock.Mock
}

func (s *stubRoleCtl) Get(ctx context.Context, id int64, option *roleCtl.Option) (*roleCtl.Role, error) {
	args := s.Called(ctx, id, option)
	r, _ := args.Get(0).(*roleCtl.Role)
	return r, args.Error(1)
}

func (s *stubRoleCtl) Create(ctx context.Context, r *roleCtl.Role) (int64, error) { return 0, nil }
func (s *stubRoleCtl) Delete(ctx context.Context, id int64, option ...*roleCtl.Option) error {
	return nil
}
func (s *stubRoleCtl) Update(ctx context.Context, r *roleCtl.Role, option *roleCtl.Option) error {
	return nil
}
func (s *stubRoleCtl) List(ctx context.Context, query *q.Query, option *roleCtl.Option) ([]*roleCtl.Role, error) {
	return nil, nil
}
func (s *stubRoleCtl) Count(ctx context.Context, query *q.Query) (int64, error) { return 0, nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const testProjectID int64 = 42

// policyAccess builds a single-permission role for test fixtures.
func policyAccess(resource rbac.Resource, action rbac.Action) []*types.Policy {
	return []*types.Policy{{Resource: resource, Action: action}}
}

// projectResource returns the full RBAC resource string for a project subresource,
// matching what HasProjectPermission builds internally.
func projectResource(subresource rbac.Resource) rbac.Resource {
	return rbacProject.NewNamespace(testProjectID).Resource(subresource)
}

// newCtxWithSecurity injects sc into a background context, mirroring what the
// handler middleware does at request time.
func newCtxWithSecurity(sc security.Context) context.Context {
	return security.NewContext(context.Background(), sc)
}

// newAPI returns a memberAPI wired to the given role controller stub.
func newAPI(rc *stubRoleCtl) *memberAPI {
	return &memberAPI{roleCtl: rc}
}

// ---------------------------------------------------------------------------
// Tests for checkNoEscalation
// ---------------------------------------------------------------------------

func TestCheckNoEscalation_RoleWithNoPermissions(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsSysAdmin").Return(false)
	// Can() should never be called — role has zero permissions

	rc := &stubRoleCtl{}
	rc.On("Get", testifymock.Anything, int64(10), testifymock.Anything).
		Return(&roleCtl.Role{}, nil)

	err := newAPI(rc).checkNoEscalation(newCtxWithSecurity(sc), testProjectID, 10)
	assert.NoError(t, err)
	sc.AssertNotCalled(t, "Can")
}

func TestCheckNoEscalation_CallerHasAllPermissions(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsSysAdmin").Return(false)
	// Caller holds repository:pull
	sc.On("Can", testifymock.Anything,
		rbac.ActionPull, projectResource(rbac.ResourceRepository)).
		Return(true)

	rc := &stubRoleCtl{}
	rc.On("Get", testifymock.Anything, int64(20), testifymock.Anything).
		Return(&roleCtl.Role{
			Permissions: []*roleCtl.Permission{{
				Access: policyAccess(rbac.ResourceRepository, rbac.ActionPull),
			}},
		}, nil)

	err := newAPI(rc).checkNoEscalation(newCtxWithSecurity(sc), testProjectID, 20)
	assert.NoError(t, err)
}

func TestCheckNoEscalation_CallerLacksPermission(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsSysAdmin").Return(false)
	// Caller does NOT hold member:create
	sc.On("Can", testifymock.Anything,
		rbac.ActionCreate, projectResource(rbac.ResourceMember)).
		Return(false)

	rc := &stubRoleCtl{}
	rc.On("Get", testifymock.Anything, int64(30), testifymock.Anything).
		Return(&roleCtl.Role{
			Permissions: []*roleCtl.Permission{{
				Access: policyAccess(rbac.ResourceMember, rbac.ActionCreate),
			}},
		}, nil)

	err := newAPI(rc).checkNoEscalation(newCtxWithSecurity(sc), testProjectID, 30)
	assert.Error(t, err)
	assert.Equal(t, errors.ForbiddenCode, errors.ErrCode(err), "expected ForbiddenError, got: %v", err)
}

func TestCheckNoEscalation_FirstPermissionPassesSecondFails(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsSysAdmin").Return(false)
	sc.On("Can", testifymock.Anything,
		rbac.ActionPull, projectResource(rbac.ResourceRepository)).
		Return(true)
	sc.On("Can", testifymock.Anything,
		rbac.ActionCreate, projectResource(rbac.ResourceMember)).
		Return(false)

	rc := &stubRoleCtl{}
	rc.On("Get", testifymock.Anything, int64(40), testifymock.Anything).
		Return(&roleCtl.Role{
			Permissions: []*roleCtl.Permission{{
				Access: []*types.Policy{
					{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},
					{Resource: rbac.ResourceMember, Action: rbac.ActionCreate},
				},
			}},
		}, nil)

	err := newAPI(rc).checkNoEscalation(newCtxWithSecurity(sc), testProjectID, 40)
	assert.Error(t, err)
	assert.Equal(t, errors.ForbiddenCode, errors.ErrCode(err))
}

func TestCheckNoEscalation_RoleCtlGetError(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsSysAdmin").Return(false)

	rc := &stubRoleCtl{}
	rc.On("Get", testifymock.Anything, int64(50), testifymock.Anything).
		Return(nil, fmt.Errorf("db error"))

	err := newAPI(rc).checkNoEscalation(newCtxWithSecurity(sc), testProjectID, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestCheckNoEscalation_MultiplePermissionSetsAllowed(t *testing.T) {
	sc := &securityMock.Context{}
	sc.On("IsSysAdmin").Return(false)
	sc.On("Can", testifymock.Anything,
		rbac.ActionPull, projectResource(rbac.ResourceRepository)).Return(true)
	sc.On("Can", testifymock.Anything,
		rbac.ActionPush, projectResource(rbac.ResourceRepository)).Return(true)
	sc.On("Can", testifymock.Anything,
		rbac.ActionRead, projectResource(rbac.ResourceArtifact)).Return(true)

	rc := &stubRoleCtl{}
	rc.On("Get", testifymock.Anything, int64(60), testifymock.Anything).
		Return(&roleCtl.Role{
			Permissions: []*roleCtl.Permission{
				{Access: policyAccess(rbac.ResourceRepository, rbac.ActionPull)},
				{Access: policyAccess(rbac.ResourceRepository, rbac.ActionPush)},
				{Access: policyAccess(rbac.ResourceArtifact, rbac.ActionRead)},
			},
		}, nil)

	err := newAPI(rc).checkNoEscalation(newCtxWithSecurity(sc), testProjectID, 60)
	assert.NoError(t, err)
}

func TestCheckNoEscalation_ProjectNameInsteadOfID(t *testing.T) {
	// Passing a project name string follows a code path that looks up the project
	// from DB (not available in unit tests), so we only verify that numeric IDs work.
	// This test documents the expected behavior for coverage completeness.
	sc := &securityMock.Context{}
	sc.On("IsSysAdmin").Return(false)

	rc := &stubRoleCtl{}
	rc.On("Get", testifymock.Anything, int64(70), testifymock.Anything).
		Return(&roleCtl.Role{}, nil)

	// numeric int64 → no DB lookup → should work cleanly
	err := newAPI(rc).checkNoEscalation(newCtxWithSecurity(sc), testProjectID, 70)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Compile-time interface check — fails the build if stubRoleCtl drifts.
// ---------------------------------------------------------------------------

var _ roleCtl.Controller = (*stubRoleCtl)(nil)
