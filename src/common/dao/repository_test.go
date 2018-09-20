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

package dao

import (
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	project    = "library"
	name       = "library/repository-test"
	repository = &models.RepoRecord{
		Name:      name,
		ProjectID: 1,
	}
)

func TestGetTotalOfRepositories(t *testing.T) {
	total, err := GetTotalOfRepositories()
	require.Nil(t, err)

	err = addRepository(repository)
	require.Nil(t, err)
	defer deleteRepository(name)

	n, err := GetTotalOfRepositories()
	require.Nil(t, err)
	assert.Equal(t, total+1, n)
}

func TestGetRepositories(t *testing.T) {
	// no query
	repositories, err := GetRepositories()
	require.Nil(t, err)
	n := len(repositories)

	err = addRepository(repository)
	require.Nil(t, err)
	defer deleteRepository(name)

	repositories, err = GetRepositories()
	require.Nil(t, err)
	assert.Equal(t, n+1, len(repositories))

	// query by name
	repositories, err = GetRepositories(&models.RepositoryQuery{
		Name: name,
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(repositories))
	assert.Equal(t, name, repositories[0].Name)

	// query by project name
	repositories, err = GetRepositories(&models.RepositoryQuery{
		ProjectName: project,
	})
	require.Nil(t, err)
	found := false
	for _, repository := range repositories {
		if repository.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// query by project ID
	repositories, err = GetRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{1},
	})
	require.Nil(t, err)
	found = false
	for _, repository := range repositories {
		if repository.Name == name {
			found = true
			break
		}
	}
	assert.True(t, found)

	// query by label ID
	labelID, err := AddLabel(&models.Label{
		Name: "label_for_test",
	})
	require.Nil(t, err)
	defer DeleteLabel(labelID)

	r, err := GetRepositoryByName(name)
	require.Nil(t, err)

	rlID, err := AddResourceLabel(&models.ResourceLabel{
		LabelID:      labelID,
		ResourceID:   r.RepositoryID,
		ResourceType: common.ResourceTypeRepository,
	})
	require.Nil(t, err)
	defer DeleteResourceLabel(rlID)

	repositories, err = GetRepositories(&models.RepositoryQuery{
		LabelID: labelID,
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(repositories))
	assert.Equal(t, name, repositories[0].Name)
}

func TestGetTopRepos(t *testing.T) {
	var err error
	require := require.New(t)

	require.NoError(GetOrmer().Begin())
	defer func() {
		require.NoError(GetOrmer().Rollback())
	}()

	projectIDs := []int64{}

	project1 := models.Project{
		OwnerID: 1,
		Name:    "project1",
	}
	project1.ProjectID, err = AddProject(project1)
	require.NoError(err)
	projectIDs = append(projectIDs, project1.ProjectID)

	project2 := models.Project{
		OwnerID: 1,
		Name:    "project2",
	}
	project2.ProjectID, err = AddProject(project2)
	require.NoError(err)
	projectIDs = append(projectIDs, project2.ProjectID)

	repository1 := &models.RepoRecord{
		Name:      fmt.Sprintf("%v/repository1", project1.Name),
		ProjectID: project1.ProjectID,
	}
	err = AddRepository(*repository1)
	require.NoError(err)
	require.NoError(IncreasePullCount(repository1.Name))

	repository2 := &models.RepoRecord{
		Name:      fmt.Sprintf("%v/repository2", project1.Name),
		ProjectID: project1.ProjectID,
	}
	err = AddRepository(*repository2)
	require.NoError(err)
	require.NoError(IncreasePullCount(repository2.Name))
	require.NoError(IncreasePullCount(repository2.Name))

	repository3 := &models.RepoRecord{
		Name:      fmt.Sprintf("%v/repository3", project2.Name),
		ProjectID: project2.ProjectID,
	}
	err = AddRepository(*repository3)
	require.NoError(err)
	require.NoError(IncreasePullCount(repository3.Name))
	require.NoError(IncreasePullCount(repository3.Name))
	require.NoError(IncreasePullCount(repository3.Name))

	topRepos, err := GetTopRepos(projectIDs, 100)
	require.NoError(err)
	require.Len(topRepos, 3)
	require.Equal(topRepos[0].Name, repository3.Name)
}

func addRepository(repository *models.RepoRecord) error {
	return AddRepository(*repository)
}

func deleteRepository(name string) error {
	return DeleteRepository(name)
}

func clearRepositoryData() error {
	return ClearTable(models.RepoTable)
}
