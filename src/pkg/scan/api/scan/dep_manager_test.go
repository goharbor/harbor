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

package scan

import (
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DepManagerTestSuite is a test suite for dep manager.
type DepManagerTestSuite struct {
	suite.Suite

	m DepManager
}

// TestDepManager is the entry point of DepManagerTestSuite.
func TestDepManager(t *testing.T) {
	suite.Run(t, new(DepManagerTestSuite))
}

// SetupSuite ...
func (suite *DepManagerTestSuite) SetupSuite() {
	suite.m = &basicDepManager{}
	dao.PrepareTestForPostgresSQL()
}

// TestDepManagerUUID ...
func (suite *DepManagerTestSuite) TestDepManagerUUID() {
	theUUID, err := suite.m.UUID()
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), theUUID)
}

// TestDepManagerMakeRobotAccount ...
func (suite *DepManagerTestSuite) TestDepManagerMakeRobotAccount() {
	tk, err := suite.m.MakeRobotAccount(1, 1800)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), tk)
}
