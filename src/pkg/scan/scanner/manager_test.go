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

package scanner

import (
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/scan/scanner/dao/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// BasicManagerTestSuite tests the basic manager
type BasicManagerTestSuite struct {
	suite.Suite

	mgr        Manager
	sampleUUID string
}

// TestBasicManager is the entry of BasicManagerTestSuite
func TestBasicManager(t *testing.T) {
	suite.Run(t, new(BasicManagerTestSuite))
}

// SetupSuite prepares env for test suite
func (suite *BasicManagerTestSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()

	suite.mgr = New()

	r := &scanner.Registration{
		Name:        "forUT",
		Description: "sample registration",
		URL:         "https://sample.scanner.com",
		Adapter:     "Clair",
		Version:     "0.1.0",
		Vendor:      "Harbor",
	}

	uid, err := suite.mgr.Create(r)
	require.NoError(suite.T(), err)
	suite.sampleUUID = uid
}

// TearDownSuite clears env for test suite
func (suite *BasicManagerTestSuite) TearDownSuite() {
	err := suite.mgr.Delete(suite.sampleUUID)
	require.NoError(suite.T(), err, "delete registration")
}

// TestList tests list registrations
func (suite *BasicManagerTestSuite) TestList() {
	m := make(map[string]string, 1)
	m["name"] = "forUT"

	l, err := suite.mgr.List(&q.Query{
		PageNumber: 1,
		PageSize:   10,
		Keywords:   m,
	})

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))
}

// TestGet tests get registration
func (suite *BasicManagerTestSuite) TestGet() {
	r, err := suite.mgr.Get(suite.sampleUUID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "forUT", r.Name)
}

// TestUpdate tests update registration
func (suite *BasicManagerTestSuite) TestUpdate() {
	r, err := suite.mgr.Get(suite.sampleUUID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)

	r.URL = "https://updated.com"
	err = suite.mgr.Update(r)
	require.NoError(suite.T(), err)

	r, err = suite.mgr.Get(suite.sampleUUID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "https://updated.com", r.URL)
}

// TestDefault tests get/set default registration
func (suite *BasicManagerTestSuite) TestDefault() {
	err := suite.mgr.SetAsDefault(suite.sampleUUID)
	require.NoError(suite.T(), err)

	dr, err := suite.mgr.GetDefault()
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), dr)
	assert.Equal(suite.T(), true, dr.IsDefault)
}
