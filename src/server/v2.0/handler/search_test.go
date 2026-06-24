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
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/project"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	repositorytesting "github.com/goharbor/harbor/src/testing/controller/repository"
	"github.com/goharbor/harbor/src/testing/mock"
)

func TestFilterRepositories_ProjectPublic(t *testing.T) {
	cases := []struct {
		name           string
		metadata       map[string]string
		wantPublic     bool
	}{
		{"public project", map[string]string{"public": "true"}, true},
		{"private project", map[string]string{"public": "false"}, false},
		{"auth_only project", map[string]string{"public": "auth_only"}, true},
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
	assert.True(t, result[0].ProjectPublic, "auth_only project should have ProjectPublic=true in search results")
	assert.Equal(t, "internal-proj", result[0].ProjectName)
}
