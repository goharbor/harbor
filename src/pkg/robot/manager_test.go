package robot

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/robot/dao"
	"github.com/stretchr/testify/suite"
	"testing"
)

type managerTestSuite struct {
	suite.Suite
	mgr *manager
	dao *dao.DAO
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &dao.DAO{}
	m.mgr = &manager{
		dao: m.dao,
	}
}

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	_, err := m.mgr.Create(context.Background(), &model.Robot{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestCount() {
	m.dao.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	n, err := m.mgr.Count(context.Background(), nil)
	m.Nil(err)
	m.Equal(int64(1), n)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDelete() {
	m.dao.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.Delete(context.Background(), 1)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDeleteByProjectID() {
	m.dao.On("DeleteByProjectID", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.DeleteByProjectID(context.Background(), 1)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdate() {
	m.dao.On("Update", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.Update(context.Background(), &model.Robot{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestList() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Robot{
		{
			ID:   1,
			Name: "robot",
		},
	}, nil)
	rpers, err := m.mgr.List(context.Background(), nil)
	m.Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
