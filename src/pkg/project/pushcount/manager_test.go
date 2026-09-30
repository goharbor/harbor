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

package pushcount

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/orm"
	htesting "github.com/goharbor/harbor/src/testing"
)

type ManagerTestSuite struct {
	htesting.Suite
	mgr Manager
}

func (suite *ManagerTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.mgr = New()
}

func (suite *ManagerTestSuite) TestAddAndGet() {
	// use a dedicated project so that parallel test packages cannot touch this counter
	suite.WithProject(func(projectID int64, _ string) {
		ctx := orm.Context()
		defer suite.ExecSQL("DELETE FROM project_push_count WHERE project_id = ?", projectID)

		count, err := suite.mgr.Get(ctx, projectID)
		suite.Require().NoError(err)
		suite.Equal(int64(0), count)

		suite.Require().NoError(suite.mgr.Add(ctx, projectID, 1))
		suite.Require().NoError(suite.mgr.Add(ctx, projectID, 2))

		count, err = suite.mgr.Get(ctx, projectID)
		suite.Require().NoError(err)
		suite.Equal(int64(3), count)
	})
}

func (suite *ManagerTestSuite) TestAddUnknownProject() {
	suite.Error(suite.mgr.Add(orm.Context(), 999999, 1))
}

func (suite *ManagerTestSuite) TestNoOrmInContext() {
	suite.Error(suite.mgr.Add(context.Background(), 1, 1))
	_, err := suite.mgr.Get(context.Background(), 1)
	suite.Error(err)
}

func TestManager(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}
