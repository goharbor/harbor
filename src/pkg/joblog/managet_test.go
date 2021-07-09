package joblog

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/joblog/models"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/joblog/dao"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type managerTestingSuite struct {
	suite.Suite
	t             *testing.T
	assert        *assert.Assertions
	require       *require.Assertions
	mockJobLogDao *dao.DAO
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
	m.mockJobLogDao = &dao.DAO{}
	Mgr = &manager{
		dao: m.mockJobLogDao,
	}
}

func TestManagerTestingSuite(t *testing.T) {
	suite.Run(t, &managerTestingSuite{})
}

func (m *managerTestingSuite) TestCreate() {
	m.mockJobLogDao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := Mgr.Create(context.Background(), &models.JobLog{})
	m.mockJobLogDao.AssertCalled(m.t, "Create", mock.Anything, mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}

func (m *managerTestingSuite) TestGet() {
	m.mockJobLogDao.On("Get", mock.Anything, mock.Anything).Return(&models.JobLog{
		LogID:   1,
		UUID:    "1234",
		Content: "test get",
	}, nil)
	ir, err := Mgr.Get(context.Background(), "1234")
	m.mockJobLogDao.AssertCalled(m.t, "Get", mock.Anything, mock.Anything)
	m.require.Nil(err)
	m.require.NotNil(ir)
	m.assert.Equal(1, ir.LogID)
}

func (m *managerTestingSuite) TestDeleteBefore() {
	m.mockJobLogDao.On("DeleteBefore", mock.Anything, mock.Anything).Return(int64(1), nil)
	_, err := Mgr.DeleteBefore(context.Background(), time.Now())
	m.mockJobLogDao.AssertCalled(m.t, "DeleteBefore", mock.Anything, mock.Anything)
	m.require.Nil(err)
}
