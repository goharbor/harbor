package user

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/user/dao"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mgrTestSuite struct {
	suite.Suite
	mgr Manager
	dao *dao.DAO
}

func (m *mgrTestSuite) SetupTest() {
	m.dao = &dao.DAO{}
	m.mgr = &manager{
		dao: m.dao,
	}
}

func (m *mgrTestSuite) TestCount() {
	m.dao.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	n, err := m.mgr.Count(context.Background(), nil)
	m.Nil(err)
	m.Equal(int64(1), n)
	m.dao.AssertExpectations(m.T())
}

func (m *mgrTestSuite) TestSetAdminFlag() {
	id := 9
	m.dao.On("Update", mock.Anything, testifymock.MatchedBy(
		func(u *models.User) bool {
			return u.UserID == 9 && u.SysAdminFlag
		}), "sysadmin_flag").Return(nil)
	err := m.mgr.SetSysAdminFlag(context.Background(), id, true)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &mgrTestSuite{})
}

func TestInjectPasswd(t *testing.T) {
	u := &models.User{
		UserID: 9,
	}
	p := "pass"
	injectPasswd(u, p)
	assert.Equal(t, "sha256", u.PasswordVersion)
	assert.Equal(t, utils.Encrypt(p, u.Salt, "sha256"), u.Password)
}
