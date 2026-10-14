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

package optimizer

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
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	htesting "github.com/goharbor/harbor/src/testing"

	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

// BasicManagerTestSuite tests the basic manager. It requires a real PostgreSQL
// database (see .vscode/CLAUDE.md "Running DAO tests locally") since Manager is a
// thin wrapper over the package-level dao functions, which talk to the ORM directly
// rather than through an injectable interface.
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

	r := &dao.Registration{
		Name:        "forUT",
		Description: "sample registration",
		URL:         "https://sample.optimizer.com",
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

// TestCount tests counting registrations
func (suite *BasicManagerTestSuite) TestCount() {
	m := make(map[string]any, 1)
	m["name"] = "forUT"

	total, err := suite.mgr.Count(suite.Context(), &q.Query{Keywords: m})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
}

// TestGet tests get registration
func (suite *BasicManagerTestSuite) TestGet() {
	r, err := suite.mgr.Get(suite.Context(), suite.sampleUUID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "forUT", r.Name)
}

// TestGetEmptyUUID tests that Get rejects an empty UUID
func (suite *BasicManagerTestSuite) TestGetEmptyUUID() {
	_, err := suite.mgr.Get(suite.Context(), "")
	require.Error(suite.T(), err)
}

// TestCreateNil tests that Create rejects a nil registration
func (suite *BasicManagerTestSuite) TestCreateNil() {
	_, err := suite.mgr.Create(suite.Context(), nil)
	require.Error(suite.T(), err)
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

// TestUpdateNil tests that Update rejects a nil registration
func (suite *BasicManagerTestSuite) TestUpdateNil() {
	err := suite.mgr.Update(suite.Context(), nil)
	require.Error(suite.T(), err)
}

// TestDeleteEmptyUUID tests that Delete rejects an empty UUID
func (suite *BasicManagerTestSuite) TestDeleteEmptyUUID() {
	err := suite.mgr.Delete(suite.Context(), "")
	require.Error(suite.T(), err)
}

// TestSetAsDefaultEmptyUUID tests that SetAsDefault rejects an empty UUID
func (suite *BasicManagerTestSuite) TestSetAsDefaultEmptyUUID() {
	err := suite.mgr.SetAsDefault(suite.Context(), "")
	require.Error(suite.T(), err)
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

// TestGetDefaultOptimizerUUID tests DefaultOptimizerUUID
func (suite *BasicManagerTestSuite) TestGetDefaultOptimizerUUID() {
	ctx := suite.Context()
	require.NoError(suite.T(), suite.mgr.SetAsDefault(ctx, suite.sampleUUID))

	uid, err := suite.mgr.DefaultOptimizerUUID(ctx)
	suite.NoError(err)
	suite.Equal(suite.sampleUUID, uid)
}

// TestCreateWithCredential tests creating a registration with credentials that should be encrypted
func (suite *BasicManagerTestSuite) TestCreateWithCredential() {
	ctx := suite.Context()
	r := &dao.Registration{
		Name:             "test-optimizer-with-cred",
		URL:              "https://optimizer-with-cred.example.com",
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
	r := &dao.Registration{
		Name:             "test-optimizer-update-cred",
		URL:              "https://optimizer-update-cred.example.com",
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
	r := &dao.Registration{
		Name:             "test-optimizer-list-cred",
		URL:              "https://optimizer-list-cred.example.com",
		Auth:             "Bearer",
		AccessCredential: "bearer-token-123",
	}

	uid, err := suite.mgr.Create(ctx, r)
	require.NoError(suite.T(), err)
	defer func() {
		suite.mgr.Delete(ctx, uid)
	}()

	m := make(map[string]any, 1)
	m["name"] = "test-optimizer-list-cred"

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
	r := &dao.Registration{
		Name:             "test-optimizer-default-cred",
		URL:              "https://optimizer-default-cred.example.com",
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

// TestEmptyCredential tests that empty credentials are handled correctly
func (suite *BasicManagerTestSuite) TestEmptyCredential() {
	ctx := suite.Context()
	r := &dao.Registration{
		Name:             "test-optimizer-empty-cred",
		URL:              "https://optimizer-empty-cred.example.com",
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

	r := &dao.Registration{
		Name:             "test-optimizer",
		AccessCredential: "",
	}

	err := encryptCredential(r)
	require.NoError(t, err)
	assert.Equal(t, "", r.AccessCredential)
}

func TestEncryptCredentialWithData(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	r := &dao.Registration{
		Name:             "test-optimizer",
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

	r := &dao.Registration{
		Name:             "test-optimizer",
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
	r := &dao.Registration{
		Name:             "test-optimizer",
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
	r := &dao.Registration{
		Name:             "test-optimizer",
		AccessCredential: originalCred,
	}

	err := encryptCredential(r)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(r.AccessCredential, utils.EncryptHeaderV1))

	err = decryptCredential(r)
	require.NoError(t, err)
	assert.Equal(t, originalCred, r.AccessCredential)
}
