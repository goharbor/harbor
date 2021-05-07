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
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/log"
	libOrm "github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/stretchr/testify/assert"
)

var testCtx context.Context

func execUpdate(o orm.Ormer, sql string, params ...interface{}) error {
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(params...)
	if err != nil {
		return err
	}
	return nil
}

func cleanByUser(username string) {
	var err error

	o := GetOrmer()
	o.Begin()

	err = execUpdate(o, `delete
		from project_member
		where entity_id = (
			select user_id
			from harbor_user
			where username = ?
		) `, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete
		from project_member
		where project_id = (
			select project_id
			from project
			where name = ?
		)`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from project where name = ?`, projectName)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}

	err = execUpdate(o, `delete from harbor_user where username = ?`, username)
	if err != nil {
		o.Rollback()
		log.Error(err)
	}
	err = execUpdate(o, `delete from replication_policy where id < 99`)
	if err != nil {
		log.Error(err)
	}
	err = execUpdate(o, `delete from registry where id < 99`)
	if err != nil {
		log.Error(err)
	}
	o.Commit()
}

const username string = "Tester01"
const password string = "Abc12345"
const projectName string = "test_project"
const repositoryName string = "test_repository"

func TestMain(m *testing.M) {
	databases := []string{"postgresql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)
		result := 1
		switch database {
		case "postgresql":
			PrepareTestForPostgresSQL()
		default:
			log.Fatalf("invalid database: %s", database)
		}
		testCtx = libOrm.Context()
		result = testForAll(m)

		if result != 0 {
			os.Exit(result)
		}
	}
}

func testForAll(m *testing.M) int {
	cleanByUser(username)
	// TODO: remove the code for populating data after the record is not needed.
	ctx := libOrm.Context()
	_, err := user.Mgr.Create(ctx, &models.User{
		Username: username,
		Email:    "tester01@vmware.com",
		Password: password,
		Realname: "tester01",
		Comment:  "register",
	})
	if err != nil {
		log.Errorf("Error occurred when creating user: %v", err)
		return 1
	}

	rc := m.Run()
	clearAll()
	return rc
}

func clearAll() {
	tables := []string{"project_member",
		"project_metadata", "repository", "replication_policy",
		"registry", "project", "harbor_user"}
	for _, t := range tables {
		if err := ClearTable(t); err != nil {
			log.Errorf("Failed to clear table: %s,error: %v", t, err)
		}
	}
}

var currentUser *models.User

func TestGetUser(t *testing.T) {
	queryUser := models.User{
		Username: username,
		Email:    "tester01@vmware.com",
	}
	var err error
	currentUser, err = GetUser(queryUser)
	if err != nil {
		t.Errorf("Error occurred in GetUser: %v", err)
	}
	if currentUser == nil {
		t.Errorf("No user found queried by user query: %+v", queryUser)
	}
	if currentUser.Email != "tester01@vmware.com" {
		t.Errorf("the user's email does not match, expected: tester01@vmware.com, actual: %s", currentUser.Email)
	}

	queryUser = models.User{}
	_, err = GetUser(queryUser)
	assert.NotNil(t, err)
}

func TestAddProject(t *testing.T) {

	project := models.Project{
		OwnerID:      currentUser.UserID,
		Name:         projectName,
		CreationTime: time.Now(),
		OwnerName:    currentUser.Username,
	}

	_, err := AddProject(project)
	if err != nil {
		t.Errorf("Error occurred in AddProject: %v", err)
	}

	newProject, err := GetProjectByName(projectName)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}
	if newProject == nil {
		t.Errorf("No project found queried by project name: %v", projectName)
	}
}

var currentProject *models.Project

func TestGetProject(t *testing.T) {
	var err error
	currentProject, err = GetProjectByName(projectName)
	if err != nil {
		t.Errorf("Error occurred in GetProjectByName: %v", err)
	}
	if currentProject == nil {
		t.Errorf("No project found queried by project name: %v", projectName)
	}
	if currentProject.Name != projectName {
		t.Errorf("Project name does not match, expected: %s, actual: %s", projectName, currentProject.Name)
	}
}

func TestGetProjectById(t *testing.T) {
	id := currentProject.ProjectID
	p, err := GetProjectByID(id)
	if err != nil {
		t.Errorf("Error in GetProjectById: %v, id: %d", err, id)
	}
	if p.Name != currentProject.Name {
		t.Errorf("project name does not match, expected: %s, actual: %s", currentProject.Name, p.Name)
	}
}

func TestGetTotalOfProjects(t *testing.T) {
	total, err := GetTotalOfProjects(nil)
	if err != nil {
		t.Fatalf("failed to get total of projects: %v", err)
	}

	if total != 2 {
		t.Errorf("unexpected total: %d != 2", total)
	}
}

func TestGetProjects(t *testing.T) {
	projects, err := GetProjects(nil)
	if err != nil {
		t.Errorf("Error occurred in GetProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("Expected length of projects is 2, but actual: %d, the projects: %+v", len(projects), projects)
	}
	if projects[1].Name != projectName {
		t.Errorf("Expected project name in the list: %s, actual: %s", projectName, projects[1].Name)
	}
}

var targetID, policyID, policyID2, policyID3, jobID, jobID2, jobID3 int64

func TestGetOrmer(t *testing.T) {
	o := GetOrmer()
	if o == nil {
		t.Errorf("Error get ormer.")
	}
}

func TestAddRepository(t *testing.T) {
	repoRecord := models.RepoRecord{
		Name:        currentProject.Name + "/" + repositoryName,
		ProjectID:   currentProject.ProjectID,
		Description: "testing repo",
		PullCount:   0,
		StarCount:   0,
	}

	err := AddRepository(repoRecord)
	if err != nil {
		t.Errorf("Error occurred in AddRepository: %v", err)
	}

	newRepoRecord, err := GetRepositoryByName(currentProject.Name + "/" + repositoryName)
	if err != nil {
		t.Errorf("Error occurred in GetRepositoryByName: %v", err)
	}
	if newRepoRecord == nil {
		t.Errorf("No repository found queried by repository name: %v", currentProject.Name+"/"+repositoryName)
	}
}

var currentRepository *models.RepoRecord

func TestGetRepositoryByName(t *testing.T) {
	var err error
	currentRepository, err = GetRepositoryByName(currentProject.Name + "/" + repositoryName)
	if err != nil {
		t.Errorf("Error occurred in GetRepositoryByName: %v", err)
	}
	if currentRepository == nil {
		t.Errorf("No repository found queried by repository name: %v", currentProject.Name+"/"+repositoryName)
	}
	if currentRepository.Name != currentProject.Name+"/"+repositoryName {
		t.Errorf("Repository name does not match, expected: %s, actual: %s", currentProject.Name+"/"+repositoryName, currentProject.Name)
	}
}

func TestDeleteRepository(t *testing.T) {
	err := DeleteRepository(currentRepository.Name)
	if err != nil {
		t.Errorf("Error occurred in DeleteRepository: %v", err)
	}
	repository, err := GetRepositoryByName(currentRepository.Name)
	if err != nil {
		t.Errorf("Error occurred in GetRepositoryByName: %v", err)
	}
	if repository != nil {
		t.Errorf("repository is not nil after deletion, repository: %+v", repository)
	}
}

func TestIsDupRecError(t *testing.T) {
	assert.True(t, IsDupRecErr(fmt.Errorf("pq: duplicate key value violates unique constraint \"properties_k_key\"")))
	assert.False(t, IsDupRecErr(fmt.Errorf("other error")))
}
