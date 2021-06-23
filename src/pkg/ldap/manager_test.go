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

package ldap

import (
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ManagerTestSuite struct {
	htesting.Suite
}

func (suite *ManagerTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.ClearSQLs = []string{"delete from harbor_user where username = 'mike02'"}
}

func (suite *ManagerTestSuite) TestPing() {
	ctx := suite.Context()
	suc, err := Mgr.Ping(ctx, ldapCfg)
	suite.Nil(err)
	suite.True(suc)
}

func (suite *ManagerTestSuite) TestSearchUser() {
	ctx := suite.Context()
	sess := NewSession(ldapCfg, groupCfg)
	users, err := Mgr.SearchUser(ctx, sess, "mike02")
	suite.Nil(err)
	suite.True(len(users) > 0)
	suite.Equal("mike02", users[0].Username)
}

func (suite *ManagerTestSuite) TestImportUser() {
	ctx := suite.Context()
	sess := NewSession(ldapCfg, groupCfg)
	failedUsers, err := Mgr.ImportUser(ctx, sess, []string{"mike03"})
	suite.Nil(err)
	suite.True(len(failedUsers) > 0)
}

func (suite *ManagerTestSuite) TestSearchGroup() {
	ctx := suite.Context()
	ugs, err := Mgr.SearchGroup(ctx, NewSession(ldapCfg, groupCfg), "harbor_admin", "")
	suite.Nil(err)
	suite.True(len(ugs) > 0)
	suite.Equal("cn=harbor_admin,ou=groups,dc=example,dc=com", ugs[0].Dn)
	ugs2, err := Mgr.SearchGroup(ctx, NewSession(ldapCfg, groupCfg), "", "cn=harbor_admin,ou=groups,dc=example,dc=com")
	suite.Nil(err)
	suite.True(len(ugs2) > 0)
	suite.Equal("harbor_admin", ugs[0].Name)
}
func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}
