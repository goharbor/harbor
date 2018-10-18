// Copyright 2018 Project Harbor Authors
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

package api

import (
	"os"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
)

const (
	// Prepare Test info
	TestUserName       = "testUser0001"
	TestUserPwd        = "testUser0001"
	TestUserEmail      = "testUser0001@mydomain.com"
	TestProName        = "testProject0001"
	TestTargetName     = "testTarget0001"
	TestRepoName       = "testRepo0001"
	AdminName          = "admin"
	DefaultProjectName = "library"
)

func CommonAddUser() {

	commonUser := models.User{
		Username: TestUserName,
		Password: TestUserPwd,
		Email:    TestUserEmail,
	}

	_, _ = dao.Register(commonUser)

}

func CommonGetUserID() int {
	queryUser := &models.User{
		Username: TestUserName,
	}
	commonUser, _ := dao.GetUser(*queryUser)
	return commonUser.UserID
}

func CommonDelUser() {
	queryUser := &models.User{
		Username: TestUserName,
	}
	commonUser, _ := dao.GetUser(*queryUser)
	_ = dao.DeleteUser(commonUser.UserID)

}

func CommonAddProject() {

	queryUser := &models.User{
		Username: "admin",
	}
	adminUser, _ := dao.GetUser(*queryUser)
	commonProject := &models.Project{
		Name:    TestProName,
		OwnerID: adminUser.UserID,
	}

	_, _ = dao.AddProject(*commonProject)

}

func CommonDelProject() {
	commonProject, _ := dao.GetProjectByName(TestProName)

	_ = dao.DeleteProject(commonProject.ProjectID)
}

func CommonAddTarget() {
	endPoint := os.Getenv("REGISTRY_URL")
	commonTarget := &models.RepTarget{
		URL:      endPoint,
		Name:     TestTargetName,
		Username: adminName,
		Password: adminPwd,
	}
	_, _ = dao.AddRepTarget(*commonTarget)
}

func CommonGetTarget() int {
	target, _ := dao.GetRepTargetByName(TestTargetName)
	return int(target.ID)
}

func CommonDelTarget() {
	target, _ := dao.GetRepTargetByName(TestTargetName)
	_ = dao.DeleteRepTarget(target.ID)
}

func CommonAddRepository() {
	commonRepository := &models.RepoRecord{
		RepositoryID: 1,
		Name:         TestRepoName,
		ProjectID:    1,
		PullCount:    1,
	}
	_ = dao.AddRepository(*commonRepository)
}

func CommonDelRepository() {
	_ = dao.DeleteRepository(TestRepoName)
}
