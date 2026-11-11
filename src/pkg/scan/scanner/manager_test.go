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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	htesting "github.com/goharbor/harbor/src/testing"

	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

// BasicManagerTestSuite tests the basic manager
type BasicManagerTestSuite struct {
	htesting.Suite

	mgr        Manager
	sampleUUID string
}

// TestBasicManager is the entry of BasicManagerTestSuite
func TestBasicManager(t *testing.T) {
	suite.Run(t, new(BasicManagerTestSuite))
}

// SetupSuite prepares env for test suite
func (suite *BasicManagerTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()

	suite.mgr = New()

	r := &scanner.Registration{
		Name:        "forUT",
		Description: "sample registration",
		URL:         "https://sample.scanner.com",
	}

	uid, err := suite.mgr.Create(suite.Context(), r)
	require.NoError(suite.T(), err)
	suite.sampleUUID = uid
}

// TearDownSuite clears env for test suite
func (suite *BasicManagerTestSuite) TearDownSuite() {
	err := suite.mgr.Delete(suite.Context(), suite.sampleUUID)
	require.NoError(suite.T(), err, "delete registration")
}

// TestList tests list registrations
func (suite *BasicManagerTestSuite) TestList() {
	m := make(map[string]any, 1)
	m["name"] = "forUT"

	l, err := suite.mgr.List(suite.Context(), &q.Query{
		PageNumber: 1,
		PageSize:   10,
		Keywords:   m,
	})

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))
}

// TestGet tests get registration
func (suite *BasicManagerTestSuite) TestGet() {
	r, err := suite.mgr.Get(suite.Context(), suite.sampleUUID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "forUT", r.Name)
}

// TestUpdate tests update registration
func (suite *BasicManagerTestSuite) TestUpdate() {
	r, err := suite.mgr.Get(suite.Context(), suite.sampleUUID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)

	r.URL = "https://updated.com"
	err = suite.mgr.Update(suite.Context(), r)
	require.NoError(suite.T(), err)

	r, err = suite.mgr.Get(suite.Context(), suite.sampleUUID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "https://updated.com", r.URL)
}

// TestDefault tests get/set default registration
func (suite *BasicManagerTestSuite) TestDefault() {
	err := suite.mgr.SetAsDefault(suite.Context(), suite.sampleUUID)
	require.NoError(suite.T(), err)

	dr, err := suite.mgr.GetDefault(suite.Context())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), dr)
	assert.Equal(suite.T(), true, dr.IsDefault)
}

// TestGetDefaultScanner tests the get default scanner
func (suite *BasicManagerTestSuite) TestGetDefaultScanner() {
	ctx := suite.Context()
	suite.mgr.SetAsDefault(ctx, suite.sampleUUID)
	scanner, err := suite.mgr.DefaultScannerUUID(ctx)
	suite.NoError(err)
	suite.Equal(suite.sampleUUID, scanner)
}

// TestCreateWithCredential tests creating a registration with credentials that should be encrypted
func (suite *BasicManagerTestSuite) TestCreateWithCredential() {
	ctx := suite.Context()
	r := &scanner.Registration{
		Name:             "test-scanner-with-cred",
		URL:              "https://scanner-with-cred.example.com",
		Auth:             "Basic",
		AccessCredential: "username:password",
	}

	uid, err := suite.mgr.Create(ctx, r)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), uid)

	defer func() {
		suite.mgr.Delete(ctx, uid)
	}()

	retrieved, err := suite.mgr.Get(ctx, uid)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)
	assert.Equal(suite.T(), "username:password", retrieved.AccessCredential)
}

// TestUpdateWithCredential tests updating a registration with credentials
func (suite *BasicManagerTestSuite) TestUpdateWithCredential() {
	ctx := suite.Context()
	r := &scanner.Registration{
		Name:             "test-scanner-update-cred",
		URL:              "https://scanner-update-cred.example.com",
		Auth:             "Basic",
		AccessCredential: "old-cred",
	}

	uid, err := suite.mgr.Create(ctx, r)
	require.NoError(suite.T(), err)
	defer func() {
		suite.mgr.Delete(ctx, uid)
	}()

	retrieved, err := suite.mgr.Get(ctx, uid)
	require.NoError(suite.T(), err)
	retrieved.AccessCredential = "new-cred"

	err = suite.mgr.Update(ctx, retrieved)
	require.NoError(suite.T(), err)

	updated, err := suite.mgr.Get(ctx, uid)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "new-cred", updated.AccessCredential)
}

// TestListWithCredentials tests listing registrations with credentials
func (suite *BasicManagerTestSuite) TestListWithCredentials() {
	ctx := suite.Context()
	r := &scanner.Registration{
		Name:             "test-scanner-list-cred",
		URL:              "https://scanner-list-cred.example.com",
		Auth:             "Bearer",
		AccessCredential: "bearer-token-123",
	}

	uid, err := suite.mgr.Create(ctx, r)
	require.NoError(suite.T(), err)
	defer func() {
		suite.mgr.Delete(ctx, uid)
	}()

	m := make(map[string]any, 1)
	m["name"] = "test-scanner-list-cred"

	list, err := suite.mgr.List(ctx, &q.Query{
		PageNumber: 1,
		PageSize:   10,
		Keywords:   m,
	})
	require.NoError(suite.T(), err)
	require.Len(suite.T(), list, 1)
	assert.Equal(suite.T(), "bearer-token-123", list[0].AccessCredential)
}

// TestGetDefaultWithCredential tests getting default registration with credentials
func (suite *BasicManagerTestSuite) TestGetDefaultWithCredential() {
	ctx := suite.Context()
	r := &scanner.Registration{
		Name:             "test-scanner-default-cred",
		URL:              "https://scanner-default-cred.example.com",
		Auth:             "Basic",
		AccessCredential: "default-cred",
	}

	uid, err := suite.mgr.Create(ctx, r)
	require.NoError(suite.T(), err)
	defer func() {
		suite.mgr.Delete(ctx, uid)
	}()

	err = suite.mgr.SetAsDefault(ctx, uid)
	require.NoError(suite.T(), err)

	defaultReg, err := suite.mgr.GetDefault(ctx)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), defaultReg)
	assert.Equal(suite.T(), "default-cred", defaultReg.AccessCredential)
}

// TestEncryptDecryptCredentialRoundTrip tests that encrypt/decrypt works correctly
func (suite *BasicManagerTestSuite) TestEncryptDecryptCredentialRoundTrip() {
	ctx := suite.Context()
	originalCred := "sensitive-api-key-12345"
	r := &scanner.Registration{
		Name:             "test-scanner-roundtrip",
		URL:              "https://scanner-roundtrip.example.com",
		Auth:             "Bearer",
		AccessCredential: originalCred,
	}

	uid, err := suite.mgr.Create(ctx, r)
	require.NoError(suite.T(), err)
	defer func() {
		suite.mgr.Delete(ctx, uid)
	}()

	retrieved, err := suite.mgr.Get(ctx, uid)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)
	assert.Equal(suite.T(), originalCred, retrieved.AccessCredential, "Credential should be decrypted correctly")
}

// TestEmptyCredential tests that empty credentials are handled correctly
func (suite *BasicManagerTestSuite) TestEmptyCredential() {
	ctx := suite.Context()
	r := &scanner.Registration{
		Name:             "test-scanner-empty-cred",
		URL:              "https://scanner-empty-cred.example.com",
		AccessCredential: "",
	}

	uid, err := suite.mgr.Create(ctx, r)
	require.NoError(suite.T(), err)
	defer func() {
		suite.mgr.Delete(ctx, uid)
	}()

	retrieved, err := suite.mgr.Get(ctx, uid)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)
	assert.Equal(suite.T(), "", retrieved.AccessCredential)
}

func TestEncryptCredentialWithEmptyCredential(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	r := &scanner.Registration{
		Name:             "test-scanner",
		AccessCredential: "",
	}

	err := encryptCredential(r)
	require.NoError(t, err)
	assert.Equal(t, "", r.AccessCredential)
}

func TestEncryptCredentialWithData(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	r := &scanner.Registration{
		Name:             "test-scanner",
		AccessCredential: "username:password",
	}

	err := encryptCredential(r)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(r.AccessCredential, utils.EncryptHeaderV1), "Encrypted data should have encryption header")
	assert.NotEqual(t, "username:password", r.AccessCredential, "Credential should be encrypted")
}

func TestDecryptCredentialWithEmptyCredential(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	r := &scanner.Registration{
		Name:             "test-scanner",
		AccessCredential: "",
	}

	err := decryptCredential(r)
	require.NoError(t, err)
	assert.Equal(t, "", r.AccessCredential)
}

func TestDecryptCredentialWithEncryptedData(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	originalCred := "username:password"
	r := &scanner.Registration{
		Name:             "test-scanner",
		AccessCredential: originalCred,
	}

	err := encryptCredential(r)
	require.NoError(t, err)

	err = decryptCredential(r)
	require.NoError(t, err)
	assert.Equal(t, originalCred, r.AccessCredential)
}

func TestEncryptDecryptCredentialRoundTrip(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	originalCred := "sensitive-api-key-12345"
	r := &scanner.Registration{
		Name:             "test-scanner",
		AccessCredential: originalCred,
	}

	err := encryptCredential(r)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(r.AccessCredential, utils.EncryptHeaderV1))

	err = decryptCredential(r)
	require.NoError(t, err)
	assert.Equal(t, originalCred, r.AccessCredential)
}
