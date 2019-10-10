package robot

import (
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type mockRobotDao struct {
	mock.Mock
}

func (m *mockRobotDao) CreateRobotAccount(r *model.Robot) (int64, error) {
	args := m.Called(r)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockRobotDao) UpdateRobotAccount(r *model.Robot) error {
	args := m.Called(r)
	return args.Error(1)
}

func (m *mockRobotDao) DeleteRobotAccount(id int64) error {
	args := m.Called(id)
	return args.Error(1)
}

func (m *mockRobotDao) GetRobotAccount(id int64) (*model.Robot, error) {
	args := m.Called(id)
	var r *model.Robot
	if args.Get(0) != nil {
		r = args.Get(0).(*model.Robot)
	}
	return r, args.Error(1)
}

func (m *mockRobotDao) ListRobotAccounts(query *q.Query) ([]*model.Robot, error) {
	args := m.Called()
	var rs []*model.Robot
	if args.Get(0) != nil {
		rs = args.Get(0).([]*model.Robot)
	}
	return rs, args.Error(1)
}

type managerTestingSuite struct {
	suite.Suite
	t            *testing.T
	assert       *assert.Assertions
	require      *require.Assertions
	mockRobotDao *mockRobotDao
}

func (m *managerTestingSuite) SetupSuite() {
	m.t = m.T()
	m.assert = assert.New(m.t)
	m.require = require.New(m.t)

	err := os.Setenv("RUN_MODE", "TEST")
	m.require.Nil(err)
}

func (m *managerTestingSuite) TearDownSuite() {
	err := os.Unsetenv("RUN_MODE")
	m.require.Nil(err)
}

func (m *managerTestingSuite) SetupTest() {
	m.mockRobotDao = &mockRobotDao{}
	Mgr = &defaultRobotManager{
		dao: m.mockRobotDao,
	}
}

func TestManagerTestingSuite(t *testing.T) {
	suite.Run(t, &managerTestingSuite{})
}

func (m *managerTestingSuite) TestCreateRobotAccount() {
	m.mockRobotDao.On("CreateRobotAccount", mock.Anything).Return(1, nil)
	id, err := Mgr.CreateRobotAccount(&model.Robot{})
	m.mockRobotDao.AssertCalled(m.t, "CreateRobotAccount", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}

func (m *managerTestingSuite) TestUpdateRobotAccount() {
	m.mockRobotDao.On("UpdateRobotAccount", mock.Anything).Return(1, nil)
	err := Mgr.UpdateRobotAccount(&model.Robot{})
	m.mockRobotDao.AssertCalled(m.t, "UpdateRobotAccount", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestDeleteRobotAccount() {
	m.mockRobotDao.On("DeleteRobotAccount", mock.Anything).Return(1, nil)
	err := Mgr.DeleteRobotAccount(int64(1))
	m.mockRobotDao.AssertCalled(m.t, "DeleteRobotAccount", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestGetRobotAccount() {
	m.mockRobotDao.On("GetRobotAccount", mock.Anything).Return(&model.Robot{
		ID:        1,
		ProjectID: 1,
		Disabled:  true,
		ExpiresAt: 150000,
	}, nil)
	ir, err := Mgr.GetRobotAccount(1)
	m.mockRobotDao.AssertCalled(m.t, "GetRobotAccount", mock.Anything)
	m.require.Nil(err)
	m.require.NotNil(ir)
	m.assert.Equal(int64(1), ir.ID)
}

func (m *managerTestingSuite) ListRobotAccount() {
	m.mockRobotDao.On("ListRobotAccount", mock.Anything).Return([]model.Robot{
		{
			ID:        1,
			ProjectID: 1,
			Disabled:  false,
			ExpiresAt: 12345,
		},
		{
			ID:        2,
			ProjectID: 1,
			Disabled:  false,
			ExpiresAt: 54321,
		}}, nil)

	rs, err := Mgr.ListRobotAccount(int64(1))
	m.mockRobotDao.AssertCalled(m.t, "ListRobotAccount", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(len(rs), 2)
	m.assert.Equal(rs[0].Disabled, false)
	m.assert.Equal(rs[1].ExpiresAt, 54321)

}
