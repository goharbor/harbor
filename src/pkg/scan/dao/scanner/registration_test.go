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

// TestAccessCredNotFilterable is the regression for the scanner access_cred filter oracle:
// access_cred stores the scanner adapter's Authorization secret and must never be usable as a
// q= filter, or list/count become a blind boolean oracle over the credential. It drives the same
// DAO path (ListRegistrations / GetTotalOfRegistrations) that ListScannerCandidatesOfProject uses.
func (suite *RegistrationDAOTestSuite) TestAccessCredNotFilterable() {
	// Seed a registration carrying a known secret.
	credUUID := uuid.New().String()
	_, err := AddRegistration(suite.Context(), &Registration{
		UUID:             credUUID,
		Name:             "forUT-cred",
		Description:      "registration with credential",
		URL:              "https://cred.scanner.com",
		Auth:             "Basic",
		AccessCredential: "Basic YWRtaW46U3VwM3JTM2NyZXQh",
	})
	require.NoError(suite.T(), err)
	defer func() {
		require.NoError(suite.T(), DeleteRegistration(suite.Context(), credUUID))
	}()

	baselineList, err := ListRegistrations(suite.Context(), nil)
	require.NoError(suite.T(), err)
	baselineTotal, err := GetTotalOfRegistrations(suite.Context(), nil)
	require.NoError(suite.T(), err)

	// A deliberately non-matching access_cred filter. If the column is filterable (the bug), this
	// predicate reaches SQL, drops the seeded row, and the count difference leaks the secret. With
	// filter:"false" the predicate is silently dropped and neither the list nor the count changes.
	query := &q.Query{Keywords: map[string]any{
		"access_cred": &q.FuzzyMatchValue{Value: "ZZZ-does-not-match"},
	}}

	l, err := ListRegistrations(suite.Context(), query)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(baselineList), len(l),
		"access_cred must not be usable as a q filter (ORM credential oracle)")

	total, err := GetTotalOfRegistrations(suite.Context(), query)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), baselineTotal, total,
		"GetTotalOfRegistrations must ignore an access_cred filter (X-Total-Count oracle)")

	// Positive control: a non-sensitive column stays filterable, so legitimate scanner filtering
	// (name/url/description) is unaffected by the fix.
	nameQuery := &q.Query{Keywords: map[string]any{
		"name": &q.FuzzyMatchValue{Value: "forUT-cred"},
	}}
	l, err = ListRegistrations(suite.Context(), nameQuery)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(l), "name filter must still work after the fix")
}
