package robot

import (
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/robot/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type managerTestingSuite struct {
	suite.Suite
	t            *testing.T
	assert       *assert.Assertions
	require      *require.Assertions
	mockRobotDao *dao.RobotAccountDao
	mgr          Manager
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
	m.mockRobotDao = &dao.RobotAccountDao{}
	m.mgr = &defaultRobotManager{
		dao: m.mockRobotDao,
	}
}

func TestManagerTestingSuite(t *testing.T) {
	suite.Run(t, &managerTestingSuite{})
}

func (m *managerTestingSuite) TestCreateRobotAccount() {
	m.mockRobotDao.On("CreateRobotAccount", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := m.mgr.CreateRobotAccount(&model.Robot{})
	m.mockRobotDao.AssertCalled(m.t, "CreateRobotAccount", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}

func (m *managerTestingSuite) TestUpdateRobotAccount() {
	m.mockRobotDao.On("UpdateRobotAccount", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.UpdateRobotAccount(&model.Robot{})
	m.mockRobotDao.AssertCalled(m.t, "UpdateRobotAccount", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestDeleteRobotAccount() {
	m.mockRobotDao.On("DeleteRobotAccount", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.DeleteRobotAccount(int64(1))
	m.mockRobotDao.AssertCalled(m.t, "DeleteRobotAccount", mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestGetRobotAccount() {
	m.mockRobotDao.On("GetRobotAccount", mock.Anything, mock.Anything).Return(&model.Robot{
		ID:        1,
		ProjectID: 1,
		Disabled:  true,
		ExpiresAt: 150000,
	}, nil)
	ir, err := m.mgr.GetRobotAccount(1)
	m.mockRobotDao.AssertCalled(m.t, "GetRobotAccount", mock.Anything, mock.Anything)
	m.require.Nil(err)
	m.require.NotNil(ir)
	m.assert.Equal(int64(1), ir.ID)
}

func (m *managerTestingSuite) ListRobotAccount() {
	m.mockRobotDao.On("ListRobotAccount", mock.Anything, mock.Anything).Return([]model.Robot{
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

	keywords := make(map[string]interface{})
	keywords["ProjectID"] = int64(1)
	query := &q.Query{
		Keywords: keywords,
	}
	rs, err := m.mgr.ListRobotAccount(query)
	m.mockRobotDao.AssertCalled(m.t, "ListRobotAccount", mock.Anything, mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(len(rs), 2)
	m.assert.Equal(rs[0].Disabled, false)
	m.assert.Equal(rs[1].ExpiresAt, 54321)

}
