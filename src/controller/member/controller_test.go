//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package member

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	comModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/member"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/pkg/usergroup"
	modelGroup "github.com/goharbor/harbor/src/pkg/usergroup/model"
	"github.com/goharbor/harbor/src/testing/mock"
	mockMember "github.com/goharbor/harbor/src/testing/pkg/member"
	mockProject "github.com/goharbor/harbor/src/testing/pkg/project"
	mockUser "github.com/goharbor/harbor/src/testing/pkg/user"
	mockUsergroup "github.com/goharbor/harbor/src/testing/pkg/usergroup"
)

type MemberControllerTestSuite struct {
	suite.Suite
	userManager   user.Manager
	memberManager member.Manager
	projectMgr    project.Manager
	groupManager  usergroup.Manager
	controller    *controller
}

func (suite *MemberControllerTestSuite) SetupSuite() {
	suite.userManager = &mockUser.Manager{}
	suite.memberManager = &mockMember.Manager{}
	suite.projectMgr = &mockProject.Manager{}
	suite.groupManager = &mockUsergroup.Manager{}
	suite.controller = &controller{
		userManager:  suite.userManager,
		mgr:          suite.memberManager,
		projectMgr:   suite.projectMgr,
		groupManager: suite.groupManager,
	}
}

func (suite *MemberControllerTestSuite) TearDownSuite() {

}

func (suite *MemberControllerTestSuite) TestAddProjectMemberWithUser() {
	mock.OnAnything(suite.projectMgr, "Get").Return(&models.Project{
		ProjectID: 1,
	}, nil)
	suite.userManager.(*mockUser.Manager).On("Get", mock.Anything, 1).Return(nil, fmt.Errorf("user not found"))
	_, err := suite.controller.Create(nil, 1, Request{MemberUser: User{UserID: 1}})
	suite.Error(err)
	_, err = suite.controller.Create(nil, 1, Request{MemberUser: User{UserID: 1}})
	suite.Error(err)
	suite.userManager.(*mockUser.Manager).On("Get", mock.Anything, 2).Return(&comModels.User{UserID: 2, Username: "mike"}, nil)
	suite.memberManager.(*mockMember.Manager).On("Add", mock.Anything, 1, 2).Return(nil)
	mock.OnAnything(suite.memberManager, "List").Return(nil, nil)
	mock.OnAnything(suite.memberManager, "AddProjectMember").Return(0, nil)
	_, err = suite.controller.Create(nil, 1, Request{MemberUser: User{UserID: 2}, Role: 1})
	suite.NoError(err)
}

func (suite *MemberControllerTestSuite) TestAddProjectMemberWithUserGroup() {
	mock.OnAnything(suite.projectMgr, "Get").Return(&models.Project{
		ProjectID: 1,
	}, nil)
	suite.groupManager.(*mockUsergroup.Manager).On("Get", mock.Anything, 1).Return(nil, fmt.Errorf("user group not found"))
	_, err := suite.controller.Create(nil, 1, Request{MemberGroup: UserGroup{ID: 1}})
	suite.Error(err)
	suite.groupManager.(*mockUsergroup.Manager).On("Get", mock.Anything, 1).Return(nil, fmt.Errorf("group not found"))
	_, err = suite.controller.Create(nil, 1, Request{MemberGroup: UserGroup{ID: 1}})
	suite.Error(err)
	suite.groupManager.(*mockUsergroup.Manager).On("Get", mock.Anything, 2).Return(&modelGroup.UserGroup{ID: 2, GroupName: "group1"}, nil)
	suite.memberManager.(*mockMember.Manager).On("Add", mock.Anything, 1, 2).Return(nil)
	mock.OnAnything(suite.memberManager, "List").Return(nil, nil)
	mock.OnAnything(suite.memberManager, "AddProjectMember").Return(0, nil)
	_, err = suite.controller.Create(nil, 1, Request{MemberGroup: UserGroup{ID: 2}, Role: 1})
	suite.NoError(err)
}

func (suite *MemberControllerTestSuite) TestIsProjectAdmin() {
	mock.OnAnything(suite.projectMgr, "ListAdminRolesOfUser").Return([]models.Member{models.Member{ID: 2, ProjectID: 2}}, nil)
	ok, err := suite.controller.IsProjectAdmin(context.Background(), comModels.User{UserID: 1})
	suite.NoError(err)
	suite.True(ok)
}

func TestMemberControllerTestSuite(t *testing.T) {
	suite.Run(t, &MemberControllerTestSuite{})
}
