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
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/models"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = New()
}

func (suite *DaoTestSuite) WithUser(f func(int64, string), usernames ...string) {
	var username string
	if len(usernames) > 0 {
		username = usernames[0]
	} else {
		username = suite.RandString(5)
	}

	o, err := orm.FromContext(orm.Context())
	if err != nil {
		suite.Fail("got error %v", err)
	}

	var userID int64

	email := fmt.Sprintf("%s@example.com", username)
	sql := "INSERT INTO harbor_user (username, realname, email, password) VALUES (?, ?, ?, 'Harbor12345') RETURNING user_id"
	suite.Nil(o.Raw(sql, username, username, email).QueryRow(&userID))

	defer func() {
		o.Raw("UPDATE harbor_user SET deleted=True, username=concat_ws('#', username, user_id), email=concat_ws('#', email, user_id) WHERE user_id = ?", userID).Exec()
	}()

	f(userID, username)
}

func (suite *DaoTestSuite) WithUserGroup(f func(int64, string), groupNames ...string) {
	var groupName string
	if len(groupNames) > 0 {
		groupName = groupNames[0]
	} else {
		groupName = suite.RandString(5)
	}

	o, err := orm.FromContext(orm.Context())
	if err != nil {
		suite.Fail("got error %v", err)
	}

	var groupID int64

	groupDN := fmt.Sprintf("cn=%s,dc=goharbor,dc=io", groupName)
	suite.Nil(o.Raw("INSERT INTO user_group (group_name, ldap_group_dn) VALUES (?, ?) RETURNING id", groupName, groupDN).QueryRow(&groupID))

	defer func() {
		o.Raw("DELETE FROM user_group WHERE id = ?", groupID).Exec()
	}()

	f(groupID, groupName)
}

func (suite *DaoTestSuite) TestCreate() {
	{
		project := &models.Project{
			Name:    "foobar",
			OwnerID: 1,
		}

		projectID, err := suite.dao.Create(orm.Context(), project)
		suite.Nil(err)
		suite.dao.Delete(orm.Context(), projectID)
	}

	{
		// project name duplicated
		project := &models.Project{
			Name:    "library",
			OwnerID: 1,
		}

		projectID, err := suite.dao.Create(orm.Context(), project)
		suite.Error(err)
		suite.True(errors.IsConflictErr(err))
		suite.Equal(int64(0), projectID)
	}
}

func (suite *DaoTestSuite) TestCount() {
	{
		count, err := suite.dao.Count(orm.Context(), q.New(q.KeyWords{"project_id": 1}))
		suite.Nil(err)
		suite.Equal(int64(1), count)
	}
}

func (suite *DaoTestSuite) TestDelete() {
	project := &models.Project{
		Name:    "foobar",
		OwnerID: 1,
	}

	projectID, err := suite.dao.Create(orm.Context(), project)
	suite.Nil(err)

	p1, err := suite.dao.Get(orm.Context(), projectID)
	suite.Nil(err)
	suite.Equal("foobar", p1.Name)

	suite.dao.Delete(orm.Context(), projectID)

	p2, err := suite.dao.Get(orm.Context(), projectID)
	suite.Error(err)
	suite.True(errors.IsNotFoundErr(err))
	suite.Nil(p2)
}

func (suite *DaoTestSuite) TestGet() {
	{
		project, err := suite.dao.Get(orm.Context(), 1)
		suite.Nil(err)
		suite.Equal("library", project.Name)
	}

	{
		// not found
		project, err := suite.dao.Get(orm.Context(), 10000)
		suite.Error(err)
		suite.True(errors.IsNotFoundErr(err))
		suite.Nil(project)
	}
}

func (suite *DaoTestSuite) TestGetByName() {
	{
		project, err := suite.dao.GetByName(orm.Context(), "library")
		suite.Nil(err)
		suite.Equal("library", project.Name)
	}

	{
		// not found
		project, err := suite.dao.GetByName(orm.Context(), "project10000")
		suite.Error(err)
		suite.True(errors.IsNotFoundErr(err))
		suite.Nil(project)
	}
}

func (suite *DaoTestSuite) TestList() {
	projectNames := []string{"foo1", "foo2", "foo3"}

	var projectIDs []int64
	for _, projectName := range projectNames {
		project := &models.Project{
			Name:    projectName,
			OwnerID: 1,
		}
		projectID, err := suite.dao.Create(orm.Context(), project)
		if suite.Nil(err) {
			projectIDs = append(projectIDs, projectID)
		}
	}

	defer func() {
		for _, projectID := range projectIDs {
			suite.dao.Delete(orm.Context(), projectID)
		}
	}()

	{
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"project_id__in": projectIDs}))
		suite.Nil(err)
		suite.Len(projects, len(projectNames))
	}
}

func (suite *DaoTestSuite) TestListByPublic() {
	{
		// default library project
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"public": true}))
		suite.Nil(err)
		suite.Len(projects, 1)
	}

	{
		// default library project
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"public": "true"}))
		suite.Nil(err)
		suite.Len(projects, 1)
	}

	{
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"public": false}))
		suite.Nil(err)
		suite.Len(projects, 0)
	}

	{
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"public": "false"}))
		suite.Nil(err)
		suite.Len(projects, 0)
	}
}

func (suite *DaoTestSuite) TestListByOwner() {
	{
		// default library project
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"owner": "admin"}))
		suite.Nil(err)
		suite.Len(projects, 1)
	}

	{
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"owner": "owner-not-found"}))
		suite.Nil(err)
		suite.Len(projects, 0)
	}

	{
		// single quotes in owner
		suite.WithUser(func(userID int64, username string) {
			project := &models.Project{
				Name:    "project-owner-name-include-single-quotes",
				OwnerID: int(userID),
			}
			projectID, err := suite.dao.Create(orm.Context(), project)
			suite.Nil(err)

			defer suite.dao.Delete(orm.Context(), projectID)

			projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"owner": username}))
			suite.Nil(err)
			suite.Len(projects, 1)
		}, "owner include single quotes ' in it")
	}

	{
		// sql inject
		suite.WithUser(func(userID int64, username string) {
			project := &models.Project{
				Name:    "project-sql-inject",
				OwnerID: int(userID),
			}
			projectID, err := suite.dao.Create(orm.Context(), project)
			suite.Nil(err)

			defer suite.dao.Delete(orm.Context(), projectID)

			projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"owner": username}))
			suite.Nil(err)
			suite.Len(projects, 1)
		}, "'owner' OR 1=1")
	}
}

func (suite *DaoTestSuite) TestListByMember() {
	{
		// project admin
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"member": &models.MemberQuery{UserID: 1, Role: common.RoleProjectAdmin}}))
		suite.Nil(err)
		suite.Len(projects, 1)
	}

	{
		// guest
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"member": &models.MemberQuery{UserID: 1, Role: common.RoleGuest}}))
		suite.Nil(err)
		suite.Len(projects, 0)
	}

	{
		// guest with public projects
		projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"member": &models.MemberQuery{UserID: 1, Role: common.RoleGuest, WithPublic: true}}))
		suite.Nil(err)
		suite.Len(projects, 1)
	}

	{
		suite.WithUser(func(userID int64, username string) {
			project := &models.Project{
				Name:    "project-with-user-group",
				OwnerID: int(userID),
			}
			projectID, err := suite.dao.Create(orm.Context(), project)
			suite.Nil(err)

			defer suite.dao.Delete(orm.Context(), projectID)

			suite.WithUserGroup(func(groupID int64, groupName string) {

				o, err := orm.FromContext(orm.Context())
				if err != nil {
					suite.Fail("got error %v", err)
				}

				var pid int64
				suite.Nil(o.Raw("INSERT INTO project_member (project_id, entity_id, role, entity_type) values (?, ?, ?, ?) RETURNING id", projectID, groupID, common.RoleGuest, "g").QueryRow(&pid))
				defer o.Raw("DELETE FROM project_member WHERE id = ?", pid)

				memberQuery := &models.MemberQuery{
					UserID:   1,
					Role:     common.RoleProjectAdmin,
					GroupIDs: []int{int(groupID)},
				}
				projects, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"member": memberQuery}))
				suite.Nil(err)
				suite.Len(projects, 2)
			})
		})
	}
}

func (suite *DaoTestSuite) TestListRoles() {
	{
		// only projectAdmin
		suite.WithUser(func(userID int64, username string) {
			project := &models.Project{
				Name:    utils.GenerateRandomString(),
				OwnerID: int(userID),
			}
			projectID, err := suite.dao.Create(orm.Context(), project)
			suite.Nil(err)
			defer suite.dao.Delete(orm.Context(), projectID)

			roles, err := suite.dao.ListRoles(orm.Context(), projectID, int(userID))
			suite.Nil(err)
			suite.Len(roles, 1)
			suite.Contains(roles, common.RoleProjectAdmin)
		})
	}

	{
		// projectAdmin and user groups
		suite.WithUser(func(userID int64, username string) {
			project := &models.Project{
				Name:    utils.GenerateRandomString(),
				OwnerID: int(userID),
			}
			projectID, err := suite.dao.Create(orm.Context(), project)
			suite.Nil(err)

			defer suite.dao.Delete(orm.Context(), projectID)

			suite.WithUserGroup(func(groupID int64, groupName string) {

				o, err := orm.FromContext(orm.Context())
				if err != nil {
					suite.Fail("got error %v", err)
				}

				var pid int64
				suite.Nil(o.Raw("INSERT INTO project_member (project_id, entity_id, role, entity_type) values (?, ?, ?, ?) RETURNING id", projectID, groupID, common.RoleGuest, "g").QueryRow(&pid))
				defer o.Raw("DELETE FROM project_member WHERE id = ?", pid)

				roles, err := suite.dao.ListRoles(orm.Context(), projectID, int(userID), int(groupID))
				suite.Nil(err)
				suite.Len(roles, 2)
				suite.Contains(roles, common.RoleProjectAdmin)
				suite.Contains(roles, common.RoleGuest)
			})
		})
	}
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
