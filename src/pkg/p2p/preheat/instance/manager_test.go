package instance

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/instance/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type instanceManagerSuite struct {
	suite.Suite
	manager *mocks.Manager
}

func (im *instanceManagerSuite) SetupTest() {
	im.manager = new(mocks.Manager)
}

func (im *instanceManagerSuite) TestSave() {
	im.manager.On("Save", mock.Anything).Return(int64(1), nil)
	id, err := im.manager.Save(nil)
	assert.Nil(im.T(), err)
	assert.Equal(im.T(), int64(1), id)
}

func (im *instanceManagerSuite) TestDelete() {
	im.manager.On("Delete", int64(1)).Return(nil)
	err := im.manager.Delete(1)
	assert.Nil(im.T(), err)
}

func (im *instanceManagerSuite) TestUpdate() {
	im.manager.On("Update", mock.Anything).Return(nil)
	err := im.manager.Update(nil)
	assert.Nil(im.T(), err)
}

func (im *instanceManagerSuite) TestGet() {
	ins := &models.Metadata{Name: "abc"}
	im.manager.On("Get", int64(1)).Return(ins, nil)
	res, err := im.manager.Get(1)
	assert.Nil(im.T(), err)
	assert.Equal(im.T(), ins, res)
}

func (im *instanceManagerSuite) TestList() {
	lists := []*models.Metadata{
		{Name: "abc"},
		{Name: "def"},
	}
	im.manager.On("List", mock.Anything).Return(2, lists, nil)
	total, res, err := im.manager.List(nil)
	assert.Nil(im.T(), err)
	assert.Equal(im.T(), 2, int(total))
	assert.Len(im.T(), res, 2)
	assert.Equal(im.T(), lists, res)
}

func TestInstanceManager(t *testing.T) {
	suite.Run(t, &instanceManagerSuite{})
}
