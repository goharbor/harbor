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
	"testing"

	"github.com/goharbor/harbor/src/pkg/types"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearSQLs = []string{
		"DELETE FROM quota WHERE id > 1",
		"DELETE FROM quota_usage WHERE id > 1",
	}
	suite.dao = New()
}

func (suite *DaoTestSuite) TestCreate() {
	hardLimits := types.ResourceList{types.ResourceCount: 1}
	usage := types.ResourceList{types.ResourceCount: 0}
	id, err := suite.dao.Create(suite.Context(), "project", "2", hardLimits, usage)
	suite.Nil(err)

	q, err := suite.dao.Get(suite.Context(), id)
	if suite.Nil(err) {
		hard, _ := q.GetHard()
		used, _ := q.GetUsed()

		suite.Equal(hardLimits, hard)
		suite.Equal(usage, used)
	}
}

func (suite *DaoTestSuite) TestDelete() {
	hardLimits := types.ResourceList{types.ResourceCount: 1}
	usage := types.ResourceList{types.ResourceCount: 0}

	id, err := suite.dao.Create(suite.Context(), "project", "3", hardLimits, usage)
	suite.Nil(err)

	{
		q, err := suite.dao.Get(suite.Context(), id)
		suite.Nil(err)
		suite.NotNil(q)
	}

	suite.Nil(suite.dao.Delete(suite.Context(), id))

	{
		_, err := suite.dao.Get(suite.Context(), id)
		suite.Error(err)
	}
}

func (suite *DaoTestSuite) TestUpdate() {
	hardLimits := types.ResourceList{types.ResourceCount: 1}
	usage := types.ResourceList{types.ResourceCount: 0}

	id, err := suite.dao.Create(suite.Context(), "project", "3", hardLimits, usage)
	suite.Nil(err)

	newHardLimits := types.ResourceList{types.ResourceCount: 2}
	newUsage := types.ResourceList{types.ResourceCount: 1}

	{
		q, err := suite.dao.Get(suite.Context(), id)
		if suite.Nil(err) {
			q.SetHard(newHardLimits).SetUsed(newUsage)

			suite.Nil(suite.dao.Update(suite.Context(), q))
		}
	}

	{
		q, err := suite.dao.Get(suite.Context(), id)
		if suite.Nil(err) {
			hard, _ := q.GetHard()
			used, _ := q.GetUsed()

			suite.Equal(newHardLimits, hard)
			suite.Equal(newUsage, used)
		}
	}
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
