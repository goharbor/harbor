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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	robotSec "github.com/goharbor/harbor/src/common/security/robot"
	ctlrobot "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib/q"
	pkgModels "github.com/goharbor/harbor/src/pkg/project/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
)

// TestListProjects_MemberQueryAuthOnly verifies that ListProjects builds the
// member subquery keywords correctly for an authenticated, non-admin local
// user under the three relevant "public" filter states: unset, explicit
// true and explicit false. It also verifies the "public" keyword is dropped
// from the query whenever the member query already encodes the auth_only
// inclusion, since keeping both would incorrectly AND-restrict the results
// (see FilterByPublic strict private semantics).
func TestListProjects_MemberQueryAuthOnly(t *testing.T) {
	cases := []struct {
		name           string
		public         *bool
		wantWithPublic bool
		wantWithAuth   bool
		wantPublicKept bool
	}{
		{"no filter", nil, true, true, false},
		{"public=true", boolPtr(true), true, false, true},
		{"public=false", boolPtr(false), false, true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtl := &projecttesting.Controller{}
			a := &projectAPI{projectCtl: mockCtl}

			mockCtl.On("Count", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
				member, ok := query.Keywords["member"].(*pkgModels.MemberQuery)
				if !ok {
					return false
				}
				if member.WithPublic != tc.wantWithPublic || member.WithAuthOnly != tc.wantWithAuth {
					return false
				}
				_, publicKept := query.Keywords["public"]
				return publicKept == tc.wantPublicKept
			})).Return(int64(0), nil).Once()

			user := &commonmodels.User{UserID: 1, Username: "member-user"}
			secCtx := local.NewSecurityContext(user)
			ctx := security.NewContext(context.Background(), secCtx)

			resp := a.ListProjects(ctx, operation.ListProjectsParams{Public: tc.public})
			require.NotNil(t, resp)

			mockCtl.AssertExpectations(t)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

// TestListProjects_MemberQueryExplicitAuthOnly verifies that an explicit
// q=public=auth_only request for a non-admin local user narrows the result
// to strictly auth_only projects: the member query still unions in the
// auth_only set (so the FilterByMember subquery has something to AND
// against), but the "public" keyword is kept rather than dropped, since
// ANDing it with FilterByPublic(auth_only) yields exactly the global
// auth_only set regardless of the member's own project memberships.
func TestListProjects_MemberQueryExplicitAuthOnly(t *testing.T) {
	mockCtl := &projecttesting.Controller{}
	a := &projectAPI{projectCtl: mockCtl}

	mockCtl.On("Count", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		member, ok := query.Keywords["member"].(*pkgModels.MemberQuery)
		if !ok || member.WithPublic || !member.WithAuthOnly {
			return false
		}
		public, publicKept := query.Keywords["public"]
		if !publicKept {
			return false
		}
		s, isStr := public.(string)
		return isStr && s == pkgModels.ProjectAuthOnly
	})).Return(int64(0), nil).Once()

	user := &commonmodels.User{UserID: 1, Username: "member-user"}
	secCtx := local.NewSecurityContext(user)
	ctx := security.NewContext(context.Background(), secCtx)

	resp := a.ListProjects(ctx, operation.ListProjectsParams{Q: strPtr("public=auth_only")})
	require.NotNil(t, resp)

	mockCtl.AssertExpectations(t)
}

// TestListProjects_RobotNamesQueryExplicitAuthOnly mirrors
// TestListProjects_MemberQueryExplicitAuthOnly for a project-scoped
// (non cover-all) system robot account.
func TestListProjects_RobotNamesQueryExplicitAuthOnly(t *testing.T) {
	mockCtl := &projecttesting.Controller{}
	a := &projectAPI{projectCtl: mockCtl}

	mockCtl.On("Count", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		names, ok := query.Keywords["names"].(*pkgModels.NamesQuery)
		if !ok || names.WithPublic || !names.WithAuthOnly {
			return false
		}
		public, publicKept := query.Keywords["public"]
		if !publicKept {
			return false
		}
		s, isStr := public.(string)
		return isStr && s == pkgModels.ProjectAuthOnly
	})).Return(int64(0), nil).Once()

	r := &ctlrobot.Robot{
		Permissions: []*ctlrobot.Permission{
			{Namespace: "myproject", Scope: ctlrobot.SCOPEPROJECT},
		},
	}
	secCtx := robotSec.NewSecurityContext(r)
	ctx := security.NewContext(context.Background(), secCtx)

	resp := a.ListProjects(ctx, operation.ListProjectsParams{Q: strPtr("public=auth_only")})
	require.NotNil(t, resp)

	mockCtl.AssertExpectations(t)
}

// TestListProjects_RobotNamesQueryAuthOnly mirrors TestListProjects_MemberQueryAuthOnly
// for a project-scoped (non cover-all) system robot account.
func TestListProjects_RobotNamesQueryAuthOnly(t *testing.T) {
	cases := []struct {
		name           string
		public         *bool
		wantWithPublic bool
		wantWithAuth   bool
		wantPublicKept bool
	}{
		{"no filter", nil, true, true, false},
		{"public=true", boolPtr(true), true, false, true},
		{"public=false", boolPtr(false), false, true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtl := &projecttesting.Controller{}
			a := &projectAPI{projectCtl: mockCtl}

			mockCtl.On("Count", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
				names, ok := query.Keywords["names"].(*pkgModels.NamesQuery)
				if !ok {
					return false
				}
				if names.WithPublic != tc.wantWithPublic || names.WithAuthOnly != tc.wantWithAuth {
					return false
				}
				_, publicKept := query.Keywords["public"]
				return publicKept == tc.wantPublicKept
			})).Return(int64(0), nil).Once()

			r := &ctlrobot.Robot{
				Permissions: []*ctlrobot.Permission{
					{Namespace: "myproject", Scope: ctlrobot.SCOPEPROJECT},
				},
			}
			secCtx := robotSec.NewSecurityContext(r)
			ctx := security.NewContext(context.Background(), secCtx)

			resp := a.ListProjects(ctx, operation.ListProjectsParams{Public: tc.public})
			require.NotNil(t, resp)

			mockCtl.AssertExpectations(t)
		})
	}
}

// TestListProjects_Anonymous verifies that anonymous requests never see
// auth_only or private projects: an explicit public=false filter short-circuits
// to an empty result without ever calling the project controller, and the
// default (no filter) case forces a strict public-only query.
func TestListProjects_Anonymous(t *testing.T) {
	t.Run("public=false returns empty without querying", func(t *testing.T) {
		mockCtl := &projecttesting.Controller{}
		a := &projectAPI{projectCtl: mockCtl}

		resp := a.ListProjects(context.Background(), operation.ListProjectsParams{Public: boolPtr(false)})
		require.NotNil(t, resp)

		mockCtl.AssertNotCalled(t, "Count", mock.Anything, mock.Anything)
		mockCtl.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
	})

	t.Run("no filter forces public=true", func(t *testing.T) {
		mockCtl := &projecttesting.Controller{}
		a := &projectAPI{projectCtl: mockCtl}

		mockCtl.On("Count", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
			public, ok := query.Keywords["public"].(bool)
			return ok && public
		})).Return(int64(0), nil).Once()

		resp := a.ListProjects(context.Background(), operation.ListProjectsParams{})
		require.NotNil(t, resp)

		mockCtl.AssertExpectations(t)
	})
}
