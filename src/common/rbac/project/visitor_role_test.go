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

package project

import (
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/stretchr/testify/suite"
)

type VisitorRoleTestSuite struct {
	suite.Suite
}

func (suite *VisitorRoleTestSuite) TestGetRoleName() {
	projectAdmin := visitorRole{roleID: common.RoleProjectAdmin}
	suite.Equal(projectAdmin.GetRoleName(), "projectAdmin")

	developer := visitorRole{roleID: common.RoleDeveloper}
	suite.Equal(developer.GetRoleName(), "developer")

	guest := visitorRole{roleID: common.RoleGuest}
	suite.Equal(guest.GetRoleName(), "guest")

	unknown := visitorRole{roleID: 404}
	suite.Equal(unknown.GetRoleName(), "")
}

func TestVisitorRoleTestSuite(t *testing.T) {
	suite.Run(t, new(VisitorRoleTestSuite))
}
