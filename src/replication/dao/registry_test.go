package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	defaultRegistry = &models.Registry{
		Name:           "daoTestDefault",
		URL:            "test.harbor.io",
		CredentialType: "basic",
		AccessKey:      "key1",
		AccessSecret:   "secret1",
		Type:           "harbor",
	}
	testRegistry1 = &models.Registry{
		Name:           "daoTest2",
		URL:            "test2.harbor.io",
		CredentialType: "basic",
		AccessKey:      "key1",
		AccessSecret:   "secret1",
		Type:           "harbor",
	}
)

type RegistrySuite struct {
	suite.Suite
	defaultID int64
}

func (suite *RegistrySuite) SetupTest() {
	assert := assert.New(suite.T())
	id, err := AddRegistry(defaultRegistry)
	assert.Nil(err)
	suite.defaultID = id
}

func (suite *RegistrySuite) TearDownTest() {
	assert := assert.New(suite.T())
	err := DeleteRegistry(suite.defaultID)
	assert.Nil(err)
}

func (suite *RegistrySuite) TestGetRegistry() {
	assert := assert.New(suite.T())

	// Get non-existed registry, should fail
	r, _ := GetRegistry(0)
	assert.Nil(r)

	// Get existed registry, should succeed
	r, err := GetRegistry(suite.defaultID)
	assert.Nil(err)
	assert.Equal(defaultRegistry.Name, r.Name)
}

func (suite *RegistrySuite) TestGetRegistryByName() {
	assert := assert.New(suite.T())

	// Get registry by empty name, should fail
	r, _ := GetRegistryByName("")
	assert.Nil(r)

	// Get non-existed registry, should fail
	r, _ = GetRegistryByName("non-exist")
	assert.Nil(r)

	// Get existed registry, should succeed
	r, err := GetRegistryByName(defaultRegistry.Name)
	assert.Nil(err)
	assert.Equal(defaultRegistry.Name, r.Name)
}

func (suite *RegistrySuite) TestGetRegistryByURL() {
	assert := assert.New(suite.T())

	// Get registry by empty url, should fail
	r, _ := GetRegistryByURL("")
	assert.Nil(r)

	// Get non-existed registry, should fail
	r, _ = GetRegistryByURL("non-exist.harbor.io")
	assert.Nil(r)

	// Get existed registry, should succeed
	r, err := GetRegistryByURL(defaultRegistry.URL)
	assert.Nil(err)
	assert.Equal(defaultRegistry.Name, r.Name)
}

func (suite *RegistrySuite) TestListRegistries() {
	assert := assert.New(suite.T())

	// Insert on more registry
	id, err := AddRegistry(testRegistry1)
	assert.Nil(err)
	assert.NotEqual(0, id)

	// List all registries, should succeed
	total, registries, err := ListRegistries()
	assert.Nil(err)
	if total < 2 {
		suite.T().Errorf("At least %d should be found in total, but got %d", 2, total)
	}

	// List default registry by normal query, should succeed
	total, registries, err = ListRegistries(&ListRegistryQuery{
		Query:  "Default",
		Offset: 0,
		Limit:  10,
	})
	assert.Nil(err)
	assert.Equal(int64(1), total)
	assert.Equal(defaultRegistry.Name, registries[0].Name)

	// List registry and limit to 1, should return one
	total, registries, err = ListRegistries(&ListRegistryQuery{
		Query:  "dao",
		Offset: 0,
		Limit:  1,
	})
	assert.Nil(err)
	assert.Equal(int64(2), total)
	assert.Equal(1, len(registries))

	// List registry and limit set to -1, should return all
	total, registries, err = ListRegistries(&ListRegistryQuery{
		Limit: -1,
	})
	assert.Nil(err)
	if total < 2 {
		suite.T().Errorf("At least %d should be found in total, but got %d", 2, total)
	}
	if len(registries) < 2 {
		suite.T().Errorf("At least %d should be returned, but got %d", 2, len(registries))
	}

	// List registry and large offset, should return empty
	total, registries, err = ListRegistries(&ListRegistryQuery{
		Offset: 10,
		Limit:  1,
	})
	assert.Nil(err)
	if total < 2 {
		suite.T().Errorf("At least %d should be found in total, but got %d", 2, total)
	}
	assert.Equal(0, len(registries))
}

func (suite *RegistrySuite) TestUpdate() {
	assert := assert.New(suite.T())

	// Get registry, should succeed
	r, err := GetRegistry(suite.defaultID)
	assert.Nil(err)
	assert.NotNil(r)

	r.AccessKey = "key2"
	err = UpdateRegistry(r)
	assert.Nil(err)

	r, err = GetRegistry(suite.defaultID)
	assert.Nil(err)
	assert.NotNil(r)
	assert.Equal("key2", r.AccessKey)
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}
