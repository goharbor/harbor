/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package api

import (
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"os"
)

const (
	//Prepare Test info
	TestUserName   = "testUser0001"
	TestUserPwd    = "testUser0001"
	TestUserEmail  = "testUser0001@mydomain.com"
	TestProName    = "testProject0001"
	TestTargetName = "testTarget0001"
)

func CommonAddUser() {

	commonUser := models.User{
		Username: TestUserName,
		Email:    TestUserPwd,
		Password: TestUserEmail,
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

func CommonPolicyEabled(policyID int, enabled int) {
	_ = dao.UpdateRepPolicyEnablement(int64(policyID), enabled)
}
