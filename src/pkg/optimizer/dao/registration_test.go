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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/q"
	htesting "github.com/goharbor/harbor/src/testing"
)

// RegistrationDAOTestSuite is the test suite of the optimizer registration DAO.
// It requires a real PostgreSQL instance — see the DAO test recipe in
// .vscode/CLAUDE.md for the throwaway container setup.
type RegistrationDAOTestSuite struct {
	htesting.Suite

	registrationID string
}

// TestRegistrationDAO is the entry of test cases
func TestRegistrationDAO(t *testing.T) {
	suite.Run(t, new(RegistrationDAOTestSuite))
}

// SetupSuite prepares the testing env for the suite
func (suite *RegistrationDAOTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.ClearTables = []string{"optimizer_registration"}
}

// SetupTest prepares stuff for test cases
func (suite *RegistrationDAOTestSuite) SetupTest() {
	suite.registrationID = uuid.New().String()
	r := &Registration{
		UUID:        suite.registrationID,
		Name:        "forUT",
		Description: "sample optimizer registration",
		URL:         "https://sample.optimizer.com",
	}

	_, err := AddRegistration(suite.Context(), r)
	require.NoError(suite.T(), err, "add new registration")
}

// TearDownTest clears the test case env
func (suite *RegistrationDAOTestSuite) TearDownTest() {
	err := DeleteRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err, "clear registration")
}

// TestGet tests get registration
func (suite *RegistrationDAOTestSuite) TestGet() {
	r, err := GetRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "forUT", r.Name)

	// not existing one
	r, err = GetRegistration(suite.Context(), "not-existing")
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), r)
}

// TestUpdate tests update registration
func (suite *RegistrationDAOTestSuite) TestUpdate() {
	r, err := GetRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)

	r.Disabled = true
	r.IsDefault = true
	r.URL = "http://updated.optimizer.com"

	err = UpdateRegistration(suite.Context(), r)
	require.NoError(suite.T(), err)

	r, err = GetRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)

	assert.Equal(suite.T(), true, r.Disabled)
	assert.Equal(suite.T(), true, r.IsDefault)
	assert.Equal(suite.T(), "http://updated.optimizer.com", r.URL)
}

// TestList tests list registrations
func (suite *RegistrationDAOTestSuite) TestList() {
	// no query
	l, err := ListRegistrations(suite.Context(), nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	// with query and found
	keywords := make(map[string]any)
	keywords["description"] = &q.FuzzyMatchValue{Value: "sample"}
	l, err = ListRegistrations(suite.Context(), &q.Query{
		PageSize:   5,
		PageNumber: 1,
		Keywords:   keywords,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	// no found
	keywords = make(map[string]any)
	keywords["uuid"] = "not-existing"
	l, err = ListRegistrations(suite.Context(), &q.Query{
		PageSize:   10,
		PageNumber: 1,
		Keywords:   keywords,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, len(l))
}

// TestDefault tests the transactional single-default logic
func (suite *RegistrationDAOTestSuite) TestDefault() {
	err := SetDefaultRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)

	dr, err := GetDefaultRegistration(suite.Context())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), dr)
	assert.Equal(suite.T(), suite.registrationID, dr.UUID)

	// add a second one and set it default: the first must be unset
	secondID := uuid.New().String()
	_, err = AddRegistration(suite.Context(), &Registration{
		UUID: secondID,
		Name: "forUT2",
		URL:  "https://second.optimizer.com",
	})
	require.NoError(suite.T(), err)
	defer func() {
		_ = DeleteRegistration(suite.Context(), secondID)
	}()

	err = SetDefaultRegistration(suite.Context(), secondID)
	require.NoError(suite.T(), err)

	dr, err = GetDefaultRegistration(suite.Context())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), dr)
	assert.Equal(suite.T(), secondID, dr.UUID)

	first, err := GetRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), first.IsDefault)

	// setting a not-existing one as default fails
	err = SetDefaultRegistration(suite.Context(), "not-existing")
	require.Error(suite.T(), err)
}

// TestCount tests count of registrations
func (suite *RegistrationDAOTestSuite) TestCount() {
	total, err := GetTotalOfRegistrations(suite.Context(), nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
}
