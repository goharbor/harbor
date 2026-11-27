package role

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/role/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/role/dao"
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
	_, err := m.mgr.Create(context.Background(), &model.Role{})
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

func (m *managerTestSuite) TestUpdate() {
	m.dao.On("Update", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.Update(context.Background(), &model.Role{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestList() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Role{
		{
			ID:   1,
			Name: "role",
		},
	}, nil)
	rpers, err := m.mgr.List(context.Background(), nil)
	m.Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestOutOfRangeID() {
	for _, id := range []int64{math.MaxInt32 + 1, math.MaxInt64, math.MinInt32 - 1} {
		_, err := m.mgr.Get(context.Background(), id)
		m.True(errors.IsNotFoundErr(err), "Get(%d): %v", id, err)

		err = m.mgr.Delete(context.Background(), id)
		m.True(errors.IsNotFoundErr(err), "Delete(%d): %v", id, err)

		err = m.mgr.Update(context.Background(), &model.Role{ID: id})
		m.True(errors.IsNotFoundErr(err), "Update(%d): %v", id, err)
	}
	m.dao.AssertNotCalled(m.T(), "Get", mock.Anything, mock.Anything)
	m.dao.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)
	m.dao.AssertNotCalled(m.T(), "Update", mock.Anything, mock.Anything)
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
