// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/models"
)

var (
	project    = "library"
	name       = "library/repository-test"
	repository = &models.RepoRecord{
		Name:      name,
		ProjectID: 1,
	}
)

func TestGetRepositoryByProjectName(t *testing.T) {
	if err := addRepository(repository); err != nil {
		t.Fatalf("failed to add repository %s: %v", name, err)
	}
	defer func() {
		if err := deleteRepository(name); err != nil {
			t.Fatalf("failed to delete repository %s: %v", name, err)
		}
	}()

	repositories, err := GetRepositoryByProjectName(project)
	if err != nil {
		t.Fatalf("failed to get repositories of project %s: %v",
			project, err)
	}

	if len(repositories) == 0 {
		t.Fatal("unexpected length of repositories: 0, at least 1")
	}

	exist := false
	for _, repo := range repositories {
		if repo.Name == name {
			exist = true
			break
		}
	}
	if !exist {
		t.Errorf("there is no repository whose name is %s", name)
	}
}

func TestGetTotalOfRepositories(t *testing.T) {
	total, err := GetTotalOfRepositories("")
	if err != nil {
		t.Fatalf("failed to get total of repositoreis: %v", err)
	}

	if err := addRepository(repository); err != nil {
		t.Fatalf("failed to add repository %s: %v", name, err)
	}
	defer func() {
		if err := deleteRepository(name); err != nil {
			t.Fatalf("failed to delete repository %s: %v", name, err)
		}
	}()

	n, err := GetTotalOfRepositories("")
	if err != nil {
		t.Fatalf("failed to get total of repositoreis: %v", err)
	}

	if n != total+1 {
		t.Errorf("unexpected total: %d != %d", n, total+1)
	}
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
		Public:  0,
	}
	project1.ProjectID, err = AddProject(project1)
	require.NoError(err)
	projectIDs = append(projectIDs, project1.ProjectID)

	project2 := models.Project{
		OwnerID: 1,
		Name:    "project2",
		Public:  0,
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

func TestGetTotalOfRepositoriesByProject(t *testing.T) {
	var projectID int64 = 1
	repoName := "library/total_count"

	total, err := GetTotalOfRepositoriesByProject([]int64{projectID}, repoName)
	if err != nil {
		t.Errorf("failed to get total of repositoreis of project %d: %v", projectID, err)
		return
	}

	if err := addRepository(&models.RepoRecord{
		Name:      repoName,
		ProjectID: projectID,
	}); err != nil {
		t.Errorf("failed to add repository %s: %v", repoName, err)
		return
	}
	defer func() {
		if err := deleteRepository(repoName); err != nil {
			t.Errorf("failed to delete repository %s: %v", name, err)
			return
		}
	}()

	n, err := GetTotalOfRepositoriesByProject([]int64{projectID}, repoName)
	if err != nil {
		t.Errorf("failed to get total of repositoreis of project %d: %v", projectID, err)
		return
	}

	if n != total+1 {
		t.Errorf("unexpected total: %d != %d", n, total+1)
	}
}

func TestGetRepositoriesByProject(t *testing.T) {
	var projectID int64 = 1
	repoName := "library/repository"

	if err := addRepository(&models.RepoRecord{
		Name:      repoName,
		ProjectID: projectID,
	}); err != nil {
		t.Errorf("failed to add repository %s: %v", repoName, err)
		return
	}
	defer func() {
		if err := deleteRepository(repoName); err != nil {
			t.Errorf("failed to delete repository %s: %v", name, err)
			return
		}
	}()

	repositories, err := GetRepositoriesByProject(projectID, repoName, 10, 0)
	if err != nil {
		t.Errorf("failed to get repositoreis of project %d: %v", projectID, err)
		return
	}

	t.Log(repositories)

	for _, repository := range repositories {
		if repository.Name == repoName {
			return
		}
	}

	t.Errorf("repository %s not found", repoName)
}

func addRepository(repository *models.RepoRecord) error {
	return AddRepository(*repository)
}

func deleteRepository(name string) error {
	return DeleteRepository(name)
}
