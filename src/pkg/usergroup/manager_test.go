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

package usergroup

import (
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ManagerTestSuite struct {
	htesting.Suite
	mgr Manager
}

func (s *ManagerTestSuite) SetupSuite() {
	s.Suite.SetupSuite()
	s.Suite.ClearTables = []string{"user_group"}
	s.mgr = newManager()
}

func (s *ManagerTestSuite) TestOnboardGroup() {
	ctx := s.Context()
	ug := &model.UserGroup{
		GroupName:   "harbor_dev",
		GroupType:   1,
		LdapGroupDN: "cn=harbor_dev,ou=groups,dc=example,dc=com",
	}
	err := s.mgr.Onboard(ctx, ug)
	s.Nil(err)
	ugs, err := s.mgr.List(ctx, q.New(q.KeyWords{"GroupType": 1, "LdapGroupDN": "cn=harbor_dev,ou=groups,dc=example,dc=com"}))
	s.Nil(err)
	s.True(len(ugs) > 0)
}

func (s *ManagerTestSuite) TestOnboardGroupWithDuplicatedName() {
	ctx := s.Context()
	ugs := []*model.UserGroup{
		{
			GroupName:   "harbor_dev",
			GroupType:   1,
			LdapGroupDN: "cn=harbor_dev,ou=groups,dc=example,dc=com",
		},
		{
			GroupName:   "harbor_dev",
			GroupType:   1,
			LdapGroupDN: "cn=harbor_dev,ou=groups,dc=example2,dc=com",
		},
		{
			GroupName:   "harbor_dev",
			GroupType:   1,
			LdapGroupDN: "cn=harbor_dev,ou=groups,dc=example3,dc=com,dc=verylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcname",
		},
	}
	for _, ug := range ugs {
		err := s.mgr.Onboard(ctx, ug)
		s.Nil(err)
	}
	// both user group should be onboard to user group
	ugs, err := s.mgr.List(ctx, q.New(q.KeyWords{"GroupType": 1, "LdapGroupDN": "cn=harbor_dev,ou=groups,dc=example,dc=com"}))
	s.Nil(err)
	s.True(len(ugs) > 0)

	ugs, err = s.mgr.List(ctx, q.New(q.KeyWords{"GroupType": 1, "LdapGroupDN": "cn=harbor_dev,ou=groups,dc=example2,dc=com"}))
	s.Nil(err)
	s.True(len(ugs) > 0)
	s.Equal("cn=harbor_dev,ou=groups,dc=example2,dc=com", ugs[0].GroupName)

	ugs, err = s.mgr.List(ctx, q.New(q.KeyWords{"GroupType": 1, "LdapGroupDN": "cn=harbor_dev,ou=groups,dc=example3,dc=com,dc=verylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcname"}))
	s.Nil(err)
	s.True(len(ugs) > 0)
	s.Equal("cn=harbor_dev,ou=groups,dc=example3,dc=com,dc=verylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcnameverylongdcna", ugs[0].GroupName)

}

func (s *ManagerTestSuite) TestPopulateGroup() {
	ctx := s.Context()
	ugs := []model.UserGroup{
		{
			GroupName:   "harbor_dev",
			GroupType:   1,
			LdapGroupDN: "cn=harbor_dev,ou=groups,dc=example,dc=com",
		},
		{
			GroupName: "myhttp_group",
			GroupType: 2,
		},
	}
	ids, err := s.mgr.Populate(ctx, ugs)
	s.Nil(err)
	s.True(len(ids) > 0)
	for _, i := range ids {
		s.True(i > 0)
	}
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}
