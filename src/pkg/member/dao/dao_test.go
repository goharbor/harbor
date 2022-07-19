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

package dao

import (
	"database/sql"
	"testing"

	"github.com/goharbor/harbor/src/common"
	_ "github.com/goharbor/harbor/src/common/dao"
	testDao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/member/models"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/user"
	userDao "github.com/goharbor/harbor/src/pkg/user/dao"
	"github.com/goharbor/harbor/src/pkg/usergroup"
	ugModel "github.com/goharbor/harbor/src/pkg/usergroup/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type DaoTestSuite struct {
	htesting.Suite
	dao        DAO
	projectMgr project.Manager
	projectID  int64
	userMgr    user.Manager
}

func (s *DaoTestSuite) SetupSuite() {
	s.Suite.SetupSuite()
	s.Suite.ClearTables = []string{"project_member"}
	s.dao = New()
	// Extract to test utils
	initSqls := []string{
		"insert into harbor_user (username, email, password, realname)  values ('member_test_01', 'member_test_01@example.com', '123456', 'member_test_01')",
		"insert into project (name, owner_id) values ('member_test_01', 1)",
		"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_01', 1, 'CN=harbor_users,OU=sample,OU=vmware,DC=harbor,DC=com')",
		"update project set owner_id = (select user_id from harbor_user where username = 'member_test_01') where name = 'member_test_01'",
		"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select user_id from harbor_user where username = 'member_test_01'), 'u', 1)",
		"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select id from user_group where group_name = 'test_group_01'), 'g', 1)",

		"insert into harbor_user (username, email, password, realname)  values ('member_test_02', 'member_test_02@example.com', '123456', 'member_test_02')",
		"insert into project (name, owner_id) values ('member_test_02', 1)",
		"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_02', 1, 'CN=harbor_users,OU=sample,OU=vmware,DC=harbor,DC=com')",
		"update project set owner_id = (select user_id from harbor_user where username = 'member_test_02') where name = 'member_test_02'",
		"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_02') , (select user_id from harbor_user where username = 'member_test_02'), 'u', 1)",
		"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_02') , (select id from user_group where group_name = 'test_group_02'), 'g', 1)",
	}

	clearSqls := []string{
		"delete from project where name='member_test_01' or name='member_test_02'",
		"delete from harbor_user where username='member_test_01' or username='member_test_02' or username='pm_sample'",
		"delete from user_group",
		"delete from project_member where id > 1",
	}
	testDao.PrepareTestData(clearSqls, initSqls)
	s.projectMgr = pkg.ProjectMgr
	s.userMgr = user.Mgr
	ctx := s.Context()
	proj, err := s.projectMgr.Get(ctx, "member_test_01")
	s.Nil(err)
	s.NotNil(proj)
	s.projectID = proj.ProjectID
}
func (s *DaoTestSuite) TearDownSuite() {
}

func (s *DaoTestSuite) TestAddProjectMember() {
	ctx := s.Context()
	proj, err := s.projectMgr.Get(ctx, "member_test_01")
	s.Nil(err)
	s.NotNil(proj)

	member := models.Member{
		ProjectID:  proj.ProjectID,
		EntityID:   1,
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	}
	pmid, err := s.dao.AddProjectMember(ctx, member)
	s.Nil(err)
	s.True(pmid > 0)

	queryMember := models.Member{
		ProjectID: proj.ProjectID,
		ID:        pmid,
	}
	memberList, err := s.dao.GetProjectMember(ctx, queryMember, nil)
	s.Nil(err)
	s.False(len(memberList) == 0)

	_, err = s.dao.AddProjectMember(ctx, models.Member{
		ProjectID:  -1,
		EntityID:   1,
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	})

	s.NotNil(err)

	_, err = s.dao.AddProjectMember(ctx, models.Member{
		ProjectID:  1,
		EntityID:   -1,
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	})

	s.NotNil(err)
}

func (s *DaoTestSuite) TestUpdateProjectMemberRole() {
	ctx := s.Context()
	proj, err := s.projectMgr.Get(ctx, "member_test_01")
	s.Nil(err)
	s.NotNil(proj)
	user := userDao.User{
		Username: "pm_sample",
		Email:    sql.NullString{String: "pm_sample@example.com", Valid: true},
		Realname: "pm_sample",
		Password: "1234567d",
	}
	o, err := orm.FromContext(ctx)
	s.Nil(err)
	userID, err := o.Insert(&user)
	s.Nil(err)
	member := models.Member{
		ProjectID:  proj.ProjectID,
		EntityID:   int(userID),
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	}

	pmid, err := s.dao.AddProjectMember(ctx, member)
	s.Nil(err)
	s.dao.UpdateProjectMemberRole(ctx, proj.ProjectID, pmid, common.RoleDeveloper)

	queryMember := models.Member{
		ProjectID:  proj.ProjectID,
		EntityID:   int(userID),
		EntityType: common.UserMember,
	}

	memberList, err := s.dao.GetProjectMember(ctx, queryMember, nil)
	s.Nil(err)
	s.True(len(memberList) == 1, "project member should exist")
	memberItem := memberList[0]
	s.Equal(common.RoleDeveloper, memberItem.Role, "should be developer role")
	s.Equal(user.Username, memberItem.Entityname)

	memberList2, err := s.dao.SearchMemberByName(ctx, proj.ProjectID, "pm_sample")
	s.Nil(err)
	s.True(len(memberList2) > 0)

	memberList3, err := s.dao.SearchMemberByName(ctx, proj.ProjectID, "")
	s.Nil(err)
	s.True(len(memberList3) > 0, "failed to search project member")
}

func (s *DaoTestSuite) TestGetProjectMembers() {
	ctx := s.Context()

	query1 := models.Member{ProjectID: s.projectID, Entityname: "member_test_01", EntityType: common.UserMember}
	member1, err := s.dao.GetProjectMember(ctx, query1, nil)
	s.Nil(err)
	s.True(len(member1) > 0)
	s.Equal(member1[0].Entityname, "member_test_01")

	query2 := models.Member{ProjectID: s.projectID, Entityname: "test_group_01", EntityType: common.GroupMember}
	member2, err := s.dao.GetProjectMember(ctx, query2, nil)
	s.Nil(err)
	s.True(len(member2) > 0)
	s.Equal(member2[0].Entityname, "test_group_01")
}

func (s *DaoTestSuite) TestGetTotalOfProjectMembers() {
	ctx := s.Context()
	tot, err := s.dao.GetTotalOfProjectMembers(ctx, s.projectID, nil)
	s.Nil(err)
	s.Equal(2, int(tot))
}

func (s *DaoTestSuite) TestListRoles() {
	ctx := s.Context()

	// nil user
	roles, err := s.dao.ListRoles(ctx, nil, 1)
	s.Nil(err)
	s.Len(roles, 0)

	// user with empty groups
	u, err := s.userMgr.GetByName(ctx, "member_test_01")
	s.Nil(err)
	s.NotNil(u)
	user := &models.User{
		UserID:   u.UserID,
		Username: u.Username,
	}
	roles, err = s.dao.ListRoles(ctx, user, s.projectID)
	s.Nil(err)
	s.Len(roles, 1)

	// user with a group whose ID doesn't exist
	user.GroupIDs = []int{9999}
	roles, err = s.dao.ListRoles(ctx, user, s.projectID)
	s.Nil(err)
	s.Len(roles, 1)
	s.Equal(common.RoleProjectAdmin, roles[0])

	// user with a valid group
	groupID, err := usergroup.Mgr.Create(ctx, ugModel.UserGroup{
		GroupName:   "group_for_list_role",
		GroupType:   1,
		LdapGroupDN: "CN=list_role_users,OU=sample,OU=vmware,DC=harbor,DC=com",
	})

	s.Nil(err)
	defer usergroup.Mgr.Delete(ctx, groupID)

	memberID, err := s.dao.AddProjectMember(ctx, models.Member{
		ProjectID:  s.projectID,
		Role:       common.RoleDeveloper,
		EntityID:   groupID,
		EntityType: "g",
	})
	s.Nil(err)
	defer s.dao.DeleteProjectMemberByID(ctx, s.projectID, memberID)

	user.GroupIDs = []int{groupID}
	roles, err = s.dao.ListRoles(ctx, user, s.projectID)
	s.Nil(err)
	s.Len(roles, 2)
	s.Equal(common.RoleProjectAdmin, roles[0])
	s.Equal(common.RoleDeveloper, roles[1])
}

func (s *DaoTestSuite) TestDeleteProjectMember() {
	ctx := s.Context()
	var addMember = models.Member{
		ProjectID:  s.projectID,
		EntityID:   1,
		EntityType: common.UserMember,
		Role:       common.RoleDeveloper,
	}
	pmid, err := s.dao.AddProjectMember(ctx, addMember)
	s.Nil(err)
	s.True(pmid > 0)

	err = s.dao.DeleteProjectMemberByID(ctx, s.projectID, pmid)
	s.Nil(err)

	// not exist
	err = s.dao.DeleteProjectMemberByID(ctx, s.projectID, -1)
	s.Nil(err)

}

func (s *DaoTestSuite) TestDeleteProjectMemberByUserId() {
	ctx := s.Context()
	userID := 22
	var addMember = models.Member{
		ProjectID:  s.projectID,
		EntityID:   userID,
		EntityType: common.UserMember,
		Role:       common.RoleDeveloper,
	}
	pmid, err := s.dao.AddProjectMember(ctx, addMember)
	s.Nil(err)
	s.True(pmid > 0)

	err = s.dao.DeleteProjectMemberByUserID(ctx, userID)
	s.Nil(err)

	queryMember := models.Member{ProjectID: s.projectID, EntityID: userID, EntityType: common.UserMember}

	// not exist
	members, err := s.dao.GetProjectMember(ctx, queryMember, nil)
	s.True(len(members) == 0)
	s.Nil(err)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
