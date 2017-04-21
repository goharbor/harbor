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
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/models"
)

var (
	project    = "library"
	name       = "library/repository-test"
	repository = &models.RepoRecord{
		Name:        name,
		OwnerName:   "admin",
		ProjectName: project,
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

func TestGetTotalOfPublicRepositories(t *testing.T) {
	total, err := GetTotalOfPublicRepositories("")
	if err != nil {
		t.Fatalf("failed to get total of public repositoreis: %v", err)
	}

	if err := addRepository(repository); err != nil {
		t.Fatalf("failed to add repository %s: %v", name, err)
	}
	defer func() {
		if err := deleteRepository(name); err != nil {
			t.Fatalf("failed to delete repository %s: %v", name, err)
		}
	}()

	n, err := GetTotalOfPublicRepositories("")
	if err != nil {
		t.Fatalf("failed to get total of public repositoreis: %v", err)
	}

	if n != total+1 {
		t.Errorf("unexpected total: %d != %d", n, total+1)
	}
}

func TestGetTotalOfUserRelevantRepositories(t *testing.T) {
	total, err := GetTotalOfUserRelevantRepositories(1, "")
	if err != nil {
		t.Fatalf("failed to get total of repositoreis for user %d: %v", 1, err)
	}

	if err := addRepository(repository); err != nil {
		t.Fatalf("failed to add repository %s: %v", name, err)
	}
	defer func() {
		if err := deleteRepository(name); err != nil {
			t.Fatalf("failed to delete repository %s: %v", name, err)
		}
	}()

	users, err := GetUserByProject(1, models.User{})
	if err != nil {
		t.Fatalf("failed to list members of project %d: %v", 1, err)
	}
	exist := false
	for _, user := range users {
		if user.UserID == 1 {
			exist = true
			break
		}
	}
	if !exist {
		if err = AddProjectMember(1, 1, models.DEVELOPER); err != nil {
			t.Fatalf("failed to add user %d to be member of project %d: %v", 1, 1, err)
		}
		defer func() {
			if err = DeleteProjectMember(1, 1); err != nil {
				t.Fatalf("failed to delete user %d from member of project %d: %v", 1, 1, err)
			}
		}()
	}

	n, err := GetTotalOfUserRelevantRepositories(1, "")
	if err != nil {
		t.Fatalf("failed to get total of public repositoreis for user %d: %v", 1, err)
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

	admin, err := GetUser(models.User{Username: "admin"})
	require.NoError(err)

	user := models.User{
		Username: "user",
		Password: "user",
		Email:    "user@test.com",
	}
	userID, err := Register(user)
	require.NoError(err)
	user.UserID = int(userID)

	//
	// public project with 1 repository
	// non-public project with 2 repositories visible by admin
	// non-public project with 1 repository visible by admin and user
	// deleted public project with 1 repository
	//

	project1 := models.Project{
		OwnerID:      admin.UserID,
		Name:         "project1",
		CreationTime: time.Now(),
		OwnerName:    admin.Username,
		Public:       0,
	}
	project1.ProjectID, err = AddProject(project1)
	require.NoError(err)

	project2 := models.Project{
		OwnerID:      user.UserID,
		Name:         "project2",
		CreationTime: time.Now(),
		OwnerName:    user.Username,
		Public:       0,
	}
	project2.ProjectID, err = AddProject(project2)
	require.NoError(err)

	err = AddRepository(*repository)
	require.NoError(err)

	repository1 := &models.RepoRecord{
		Name:        fmt.Sprintf("%v/repository1", project1.Name),
		OwnerName:   admin.Username,
		ProjectName: project1.Name,
	}
	err = AddRepository(*repository1)
	require.NoError(err)
	require.NoError(IncreasePullCount(repository1.Name))
	repository1, err = GetRepositoryByName(repository1.Name)
	require.NoError(err)

	repository2 := &models.RepoRecord{
		Name:        fmt.Sprintf("%v/repository2", project1.Name),
		OwnerName:   admin.Username,
		ProjectName: project1.Name,
	}
	err = AddRepository(*repository2)
	require.NoError(err)
	require.NoError(IncreasePullCount(repository2.Name))
	require.NoError(IncreasePullCount(repository2.Name))
	repository2, err = GetRepositoryByName(repository2.Name)
	require.NoError(err)

	repository3 := &models.RepoRecord{
		Name:        fmt.Sprintf("%v/repository3", project2.Name),
		OwnerName:   admin.Username,
		ProjectName: project2.Name,
	}
	err = AddRepository(*repository3)
	require.NoError(err)
	require.NoError(IncreasePullCount(repository3.Name))
	require.NoError(IncreasePullCount(repository3.Name))
	require.NoError(IncreasePullCount(repository3.Name))
	repository3, err = GetRepositoryByName(repository3.Name)
	require.NoError(err)

	deletedPublicProject := models.Project{
		OwnerID:      admin.UserID,
		Name:         "public-deleted",
		CreationTime: time.Now(),
		OwnerName:    admin.Username,
		Public:       1,
	}
	deletedPublicProject.ProjectID, err = AddProject(deletedPublicProject)
	require.NoError(err)
	deletedPublicRepository1 := &models.RepoRecord{
		Name:        fmt.Sprintf("%v/repository1", deletedPublicProject.Name),
		OwnerName:   admin.Username,
		ProjectName: deletedPublicProject.Name,
	}
	err = AddRepository(*deletedPublicRepository1)
	require.NoError(err)
	err = DeleteProject(deletedPublicProject.ProjectID)
	require.NoError(err)

	var topRepos []*models.RepoRecord

	// anonymous should retrieve public non-deleted repositories
	topRepos, err = GetTopRepos(NonExistUserID, 100)
	require.NoError(err)
	require.Len(topRepos, 1)
	require.Equal(topRepos[0].Name, repository.Name)

	// admin should retrieve all repositories
	topRepos, err = GetTopRepos(admin.UserID, 100)
	require.NoError(err)
	require.Len(topRepos, 4)

	// user should retrieve visible repositories
	topRepos, err = GetTopRepos(user.UserID, 100)
	require.NoError(err)
	require.Len(topRepos, 2)

	// limit by count
	topRepos, err = GetTopRepos(admin.UserID, 3)
	require.NoError(err)
	require.Len(topRepos, 3)
}

func TestGetTotalOfRepositoriesByProject(t *testing.T) {
	var projectID int64 = 1
	repoName := "library/total_count"

	total, err := GetTotalOfRepositoriesByProject(projectID, repoName)
	if err != nil {
		t.Errorf("failed to get total of repositoreis of project %d: %v", projectID, err)
		return
	}

	if err := addRepository(&models.RepoRecord{
		Name:        repoName,
		OwnerName:   "admin",
		ProjectName: "library",
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

	n, err := GetTotalOfRepositoriesByProject(projectID, repoName)
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
		Name:        repoName,
		OwnerName:   "admin",
		ProjectName: "library",
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
