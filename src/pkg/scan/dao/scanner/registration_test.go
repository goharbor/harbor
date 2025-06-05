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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/q"
	htesting "github.com/goharbor/harbor/src/testing"
)

// RegistrationDAOTestSuite is test suite of testing registration DAO
type RegistrationDAOTestSuite struct {
	htesting.Suite

	registrationID string
}

// TestRegistrationDAO is entry of test cases
func TestRegistrationDAO(t *testing.T) {
	suite.Run(t, new(RegistrationDAOTestSuite))
}

// SetupSuite prepare testing env for the suite
func (suite *RegistrationDAOTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
}

// SetupTest prepare stuff for test cases
func (suite *RegistrationDAOTestSuite) SetupTest() {
	suite.registrationID = uuid.New().String()
	r := &Registration{
		UUID:        suite.registrationID,
		Name:        "forUT",
		Description: "sample registration",
		URL:         "https://sample.scanner.com",
	}

	_, err := AddRegistration(suite.Context(), r)
	require.NoError(suite.T(), err, "add new registration")

}

// TearDownTest clears all the stuff of test cases
func (suite *RegistrationDAOTestSuite) TearDownTest() {
	err := DeleteRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err, "clear registration")
}

// TestGet tests get registration
func (suite *RegistrationDAOTestSuite) TestGet() {
	// Found
	r, err := GetRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), r.Name, "forUT")

	// Not found
	re, err := GetRegistration(suite.Context(), "not_found")
	require.NoError(suite.T(), err)
	require.Nil(suite.T(), re)
}

// TestUpdate tests update registration
func (suite *RegistrationDAOTestSuite) TestUpdate() {
	r, err := GetRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)

	r.Disabled = true
	r.IsDefault = true
	r.URL = "http://updated.registration.com"

	err = UpdateRegistration(suite.Context(), r)
	require.NoError(suite.T(), err, "update registration")

	r, err = GetRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)

	assert.Equal(suite.T(), true, r.Disabled)
	assert.Equal(suite.T(), true, r.IsDefault)
	assert.Equal(suite.T(), "http://updated.registration.com", r.URL)
}

// TestList tests list registrations
func (suite *RegistrationDAOTestSuite) TestList() {
	// no query
	l, err := ListRegistrations(suite.Context(), nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	// with query and found items
	keywords := make(map[string]any)
	keywords["description"] = &q.FuzzyMatchValue{Value: "sample"}
	l, err = ListRegistrations(suite.Context(), &q.Query{
		PageSize:   5,
		PageNumber: 1,
		Keywords:   keywords,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	// With query and not found items
	keywords["description"] = &q.FuzzyMatchValue{Value: "not_exist"}
	l, err = ListRegistrations(suite.Context(), &q.Query{
		Keywords: keywords,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, len(l))

	// Exact match
	exactKeywords := make(map[string]any)
	exactKeywords["name"] = "forUT"
	l, err = ListRegistrations(suite.Context(), &q.Query{
		Keywords: exactKeywords,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	exactKeywords["name"] = "forU"
	l, err = ListRegistrations(suite.Context(), &q.Query{
		Keywords: exactKeywords,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, len(l))
}

// TestDefault tests set/get default
func (suite *RegistrationDAOTestSuite) TestDefault() {
	dr, err := GetDefaultRegistration(suite.Context())
	require.NoError(suite.T(), err, "not found")
	require.Nil(suite.T(), dr)

	err = SetDefaultRegistration(suite.Context(), suite.registrationID)
	require.NoError(suite.T(), err)

	dr, err = GetDefaultRegistration(suite.Context())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), dr)

	dr.Disabled = true
	err = UpdateRegistration(suite.Context(), dr, "disabled")
	require.NoError(suite.T(), err)

	err = SetDefaultRegistration(suite.Context(), suite.registrationID)
	require.Error(suite.T(), err)
}
