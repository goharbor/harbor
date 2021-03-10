package policy

import (
	"context"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/notification/policy/dao"
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
	_, err := m.mgr.Create(context.Background(), &model.Policy{})
	m.Nil(err)
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
	err := m.mgr.Update(context.Background(), &model.Policy{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGet() {
	m.dao.On("Get", mock.Anything, mock.Anything).Return(&model.Policy{
		Name: "test_policy",
	}, nil)
	policy, err := m.mgr.Get(context.Background(), 1)
	m.Nil(err)
	m.Equal("test_policy", policy.Name)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestList() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Policy{
		{
			ID:   1,
			Name: "policy",
		},
	}, nil)
	rpers, err := m.mgr.List(context.Background(), nil)
	m.Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGetByNameAndProjectID404() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Policy{}, nil)
	_, err := m.mgr.GetByNameAndProjectID(context.Background(), "test_policy", 1)
	m.NotNil(err)
	m.True(errors.IsNotFoundErr(err))
}

func (m *managerTestSuite) TestGetByNameAndProjectID() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Policy{
		{
			ID:        1,
			Name:      "test_policy",
			ProjectID: 1,
		},
	}, nil)
	policy, err := m.mgr.GetByNameAndProjectID(context.Background(), "test_policy", 1)
	m.Nil(err)
	m.Equal("test_policy", policy.Name)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGetRelatedPolices() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Policy{
		{
			ID:           1,
			Name:         "policy",
			ProjectID:    1,
			Enabled:      true,
			EventTypesDB: "[\"PULL_IMAGE\",\"PUSH_CHART\"]",
		},
		{
			ID:           2,
			Name:         "policy",
			ProjectID:    1,
			Enabled:      true,
			EventTypesDB: "[\"PULL_IMAGE\",\"PUSH_CHART\"]",
		},
	}, nil)
	rpers, err := m.mgr.GetRelatedPolices(context.Background(), 1, "PULL_IMAGE")
	m.Nil(err)
	m.Equal(2, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
