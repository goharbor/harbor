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

	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/q"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/search"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	repositorytesting "github.com/goharbor/harbor/src/testing/controller/repository"
	"github.com/goharbor/harbor/src/testing/mock"
)

func TestFilterRepositories_ProjectPublic(t *testing.T) {
	cases := []struct {
		name       string
		metadata   map[string]string
		wantPublic bool
	}{
		{"public project", map[string]string{"public": "true"}, true},
		{"private project", map[string]string{"public": "false"}, false},
		{"auth_only project", map[string]string{"public": "auth_only"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoCtl := &repositorytesting.Controller{}
			artifactCtl := &artifacttesting.Controller{}

			proj := &project.Project{
				ProjectID: 1,
				Name:      "testproj",
				Metadata:  tc.metadata,
			}

			repo := &repomodel.RepoRecord{
				RepositoryID: 10,
				Name:         "testproj/redis",
				ProjectID:    1,
				PullCount:    5,
			}

			mock.OnAnything(repoCtl, "List").Return([]*repomodel.RepoRecord{repo}, nil)
			mock.OnAnything(artifactCtl, "Count").Return(int64(3), nil)

			s := &searchAPI{
				repositoryCtl: repoCtl,
				artifactCtl:   artifactCtl,
			}

			result, err := s.filterRepositories(context.Background(), []*project.Project{proj}, "")
			require.NoError(t, err)
			require.Len(t, result, 1)
			assert.Equal(t, tc.wantPublic, result[0].ProjectPublic)
		})
	}
}

func TestFilterRepositories_AuthOnlyProjectName(t *testing.T) {
	repoCtl := &repositorytesting.Controller{}
	artifactCtl := &artifacttesting.Controller{}

	proj := &project.Project{
		ProjectID: 2,
		Name:      "internal-proj",
		Metadata:  map[string]string{proModels.ProMetaPublic: proModels.ProjectAuthOnly},
	}

	repo := &repomodel.RepoRecord{
		RepositoryID: 20,
		Name:         "internal-proj/alpine",
		ProjectID:    2,
	}

	mock.OnAnything(repoCtl, "List").Return([]*repomodel.RepoRecord{repo}, nil)
	mock.OnAnything(artifactCtl, "Count").Return(int64(1), nil)

	s := &searchAPI{
		repositoryCtl: repoCtl,
		artifactCtl:   artifactCtl,
	}

	result, err := s.filterRepositories(context.Background(), []*project.Project{proj}, "")
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.False(t, result[0].ProjectPublic, "auth_only project should not be reported as public in search results")
	assert.Equal(t, "internal-proj", result[0].ProjectName)
}

// TestSearch_MemberQueryAuthOnly verifies that Search includes auth_only
// projects for an authenticated, non-admin local user by setting
// WithAuthOnly on the member query keyword, mirroring the visibility rule
// used by ListProjects.
func TestSearch_MemberQueryAuthOnly(t *testing.T) {
	projectCtl := &projecttesting.Controller{}
	repoCtl := &repositorytesting.Controller{}
	artifactCtl := &artifacttesting.Controller{}

	projectCtl.On("List", tmock.Anything, tmock.MatchedBy(func(query *q.Query) bool {
		member, ok := query.Keywords["member"].(*project.MemberQuery)
		return ok && member.WithPublic && member.WithAuthOnly
	})).Return([]*project.Project{}, nil).Once()

	mock.OnAnything(repoCtl, "List").Return([]*repomodel.RepoRecord{}, nil)

	s := &searchAPI{
		projectCtl:    projectCtl,
		repositoryCtl: repoCtl,
		artifactCtl:   artifactCtl,
	}

	user := &commonmodels.User{UserID: 1, Username: "member-user"}
	secCtx := local.NewSecurityContext(user)
	ctx := security.NewContext(context.Background(), secCtx)

	resp := s.Search(ctx, operation.SearchParams{})
	require.NotNil(t, resp)

	projectCtl.AssertExpectations(t)
}

// TestSearch_NonLocalAuthenticated verifies that Search falls back to the
// strict "public" keyword (no auth_only inclusion) when the security
// context is not an authenticated local user.
func TestSearch_NonLocalAuthenticated(t *testing.T) {
	projectCtl := &projecttesting.Controller{}
	repoCtl := &repositorytesting.Controller{}
	artifactCtl := &artifacttesting.Controller{}

	projectCtl.On("List", tmock.Anything, tmock.MatchedBy(func(query *q.Query) bool {
		public, ok := query.Keywords["public"].(bool)
		_, hasMember := query.Keywords["member"]
		return ok && public && !hasMember
	})).Return([]*project.Project{}, nil).Once()

	mock.OnAnything(repoCtl, "List").Return([]*repomodel.RepoRecord{}, nil)

	s := &searchAPI{
		projectCtl:    projectCtl,
		repositoryCtl: repoCtl,
		artifactCtl:   artifactCtl,
	}

	// an unauthenticated local security context is neither sys admin nor an
	// authenticated local user, so Search should take the "public only" path.
	secCtx := local.NewSecurityContext(nil)
	ctx := security.NewContext(context.Background(), secCtx)

	resp := s.Search(ctx, operation.SearchParams{})
	require.NotNil(t, resp)

	projectCtl.AssertExpectations(t)
}
