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
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	htesting "github.com/goharbor/harbor/src/testing"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (s *DaoTestSuite) SetupSuite() {
	s.Suite.SetupSuite()
	s.Suite.ClearTables = []string{"user_group"}
	s.dao = New()
}

func (s *DaoTestSuite) TestCRUDUsergroup() {
	ctx := s.Context()
	userGroup := model.UserGroup{
		GroupName:   "harbor_dev",
		GroupType:   1,
		LdapGroupDN: "cn=harbor_dev,ou=groups,dc=example,dc=com",
	}
	id, err := s.dao.Add(ctx, userGroup)
	s.Nil(err)
	s.True(id > 0)

	ug, err2 := s.dao.Get(ctx, id)
	s.Nil(err2)
	s.Equal("harbor_dev", ug.GroupName)
	s.Equal("cn=harbor_dev,ou=groups,dc=example,dc=com", ug.LdapGroupDN)
	s.Equal(1, ug.GroupType)

	err3 := s.dao.UpdateName(ctx, id, "my_harbor_dev")
	s.Nil(err3)

	ug2, err4 := s.dao.Get(ctx, id)
	s.Nil(err4)
	s.Equal("my_harbor_dev", ug2.GroupName)
	s.Equal("cn=harbor_dev,ou=groups,dc=example,dc=com", ug2.LdapGroupDN)
	s.Equal(1, ug2.GroupType)

	err5 := s.dao.Delete(ctx, id)
	s.Nil(err5)
}

func (s *DaoTestSuite) TestSearchByName() {
	ctx := s.Context()
	for i := 0; i < 10; i++ {
		ug := model.UserGroup{
			GroupName:   "group_" + string(rune(i+'0')),
			GroupType:   1,
			LdapGroupDN: "cn=group_" + string(rune(i+'0')) + ",ou=groups,dc=example,dc=com",
		}
		_, err := s.dao.Add(ctx, ug)
		s.Nil(err)
	}
	userGroup1 := model.UserGroup{
		GroupName:   "group",
		GroupType:   1,
		LdapGroupDN: "cn=dev_team,ou=groups,dc=example,dc=com",
	}
	_, err := s.dao.Add(ctx, userGroup1)
	s.Nil(err)

	// no result
	results, err := s.dao.SearchByName(ctx, "dev", 10)
	s.Nil(err)
	s.Equal(0, len(results))

	// the total result should limit to 10
	results2, err := s.dao.SearchByName(ctx, "group", 10)
	s.Nil(err)
	s.Equal(10, len(results2))
	// the first one should be "group"
	s.Equal("group", results2[0].GroupName)
}

// TestReadOrCreateKeepsTransactionUsable is the regression test for #22716.
// Onboarding an LDAP group reads by ldap_group_dn, but user_group is unique on
// group_name, so a group already onboarded under a different DN makes the
// insert fail with a duplicate key. That error must stay contained: the caller
// keeps a usable transaction instead of a PostgreSQL 25P02 on every following
// statement.
func (s *DaoTestSuite) TestReadOrCreateKeepsTransactionUsable() {
	ctx := s.Context()

	_, err := s.dao.Add(ctx, model.UserGroup{
		GroupName:   "harbor_qa",
		GroupType:   1,
		LdapGroupDN: "cn=harbor_qa,ou=groups,dc=example,dc=com",
	})
	s.Require().NoError(err)

	err = orm.WithTransaction(func(txCtx context.Context) error {
		_, _, err := s.dao.ReadOrCreate(txCtx, &model.UserGroup{
			GroupName:   "harbor_qa",
			GroupType:   1,
			LdapGroupDN: "cn=harbor_qa,ou=other,dc=example,dc=com",
		}, "LdapGroupDN", "GroupType")
		s.Require().Error(err)

		_, err = s.dao.Add(txCtx, model.UserGroup{
			GroupName:   "harbor_ops",
			GroupType:   1,
			LdapGroupDN: "cn=harbor_ops,ou=groups,dc=example,dc=com",
		})
		return err
	})(ctx)
	s.Require().NoError(err)

	ugs, err := s.dao.Query(ctx, q.New(q.KeyWords{"GroupName": "harbor_ops"}))
	s.Require().NoError(err)
	s.Len(ugs, 1)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
