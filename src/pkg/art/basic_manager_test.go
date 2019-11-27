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

package art

import (
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestManagerSuite is a test suite for testing manager
type TestManagerSuite struct {
	suite.Suite

	m *basicManager
}

// TestManager is the entry point of TestManagerSuite
func TestManager(t *testing.T) {
	suite.Run(t, &TestManagerSuite{})
}

// SetupSuite prepares env for test suite
func (suite *TestManagerSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()

	suite.m = &basicManager{}
}

// TestManagerSuiteList ...
func (suite *TestManagerSuite) TestManagerSuiteList() {
	kws := make(map[string]interface{})
	kws["digest"] = "fake-digest"

	l, err := suite.m.List(&q.Query{
		Keywords: kws,
	})

	require.NoError(suite.T(), err)
	suite.Equal(0, len(l))
}
