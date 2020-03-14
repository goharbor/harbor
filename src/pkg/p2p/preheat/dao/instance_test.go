package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	defaultInstance = &models.Instance{
		ID:             1,
		Name:           "dragonfly-cn-1",
		Description:    "fake dragonfly server",
		Provider:       "dragonfly",
		Endpoint:       "https://cn-1.dragonfly.com",
		AuthMode:       "basic",
		AuthData:       "{\"username\": \"admin\", \"password\": \"123456\"}",
		Status:         "healthy",
		Enabled:        true,
		SetupTimestamp: 1582721396,
		Extensions:     "",
	}
)

type instanceSuite struct {
	suite.Suite
}

func (is *instanceSuite) SetupTest() {
	t := is.T()
	_, err := AddInstance(defaultInstance)
	assert.Nil(t, err)
}

func (is *instanceSuite) TearDownTest() {
	t := is.T()
	err := DeleteInstance(defaultInstance.ID)
	assert.Nil(t, err)
}

func (is *instanceSuite) TestGetInstance() {
	t := is.T()
	i, err := GetInstance(defaultInstance.ID)
	assert.Nil(t, err)
	assert.Equal(t, defaultInstance.Name, i.Name)

	// not exist
	i, err = GetInstance(0)
	assert.Nil(t, i)
}

func (is *instanceSuite) TestUpdateInstance() {
	t := is.T()
	i, err := GetInstance(defaultInstance.ID)
	assert.Nil(t, err)
	assert.NotNil(t, i)

	i.Enabled = false
	err = UpdateInstance(i)
	assert.Nil(t, err)

	i, err = GetInstance(defaultInstance.ID)
	assert.Nil(t, err)
	assert.NotNil(t, i)
	assert.False(t, i.Enabled)
}

func (is *instanceSuite) TestListInstances() {
	t := is.T()
	// add more instances
	testInstance1 := &models.Instance{
		ID:             2,
		Name:           "kraken-us-1",
		Description:    "fake kraken server",
		Provider:       "kraken",
		Endpoint:       "https://us-1.kraken.com",
		AuthMode:       "none",
		AuthData:       "",
		Status:         "success",
		Enabled:        true,
		SetupTimestamp: 0,
		Extensions:     "",
	}
	_, err := AddInstance(testInstance1)
	assert.Nil(t, err)

	// without limit should return all instances
	total, instances, err := ListInstances(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, int(total))
	assert.Len(t, instances, 2)

	// limit 1
	total, instances, err = ListInstances(&ListInstanceQuery{PageSize: 1, Page: 1})
	assert.Nil(t, err)
	assert.Equal(t, 2, int(total))
	assert.Len(t, instances, 1)
	assert.Equal(t, defaultInstance.ID, instances[0].ID)

	// keyword search
	total, instances, err = ListInstances(&ListInstanceQuery{Keyword: "kraken"})
	assert.Nil(t, err)
	assert.Equal(t, 1, int(total))
	assert.Len(t, instances, 1)
	assert.Equal(t, testInstance1.Name, instances[0].Name)

	// clean data
	err = DeleteInstance(testInstance1.ID)
	assert.Nil(t, err)
}

func TestInstance(t *testing.T) {
	suite.Run(t, &instanceSuite{})
}
