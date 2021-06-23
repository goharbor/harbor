package user

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
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

func (m *mgrTestSuite) TestOnboard() {
	existingUser := &models.User{
		UserID:   123,
		Username: "existing",
		Email:    "existing@mytest.com",
		Realname: "existing",
	}
	newID := 124
	m.dao.On("Create", mock.Anything, testifymock.MatchedBy(
		func(u *models.User) bool {
			return u.Username == "existing"
		})).Return(0, errors.ConflictError(nil).WithMessage("username exists"))
	m.dao.On("Create", mock.Anything, testifymock.MatchedBy(
		func(u *models.User) bool {
			return u.Username != "existing" && u.Username != "dup-but-not-existing"
		})).Return(newID, nil)
	m.dao.On("List", mock.Anything, testifymock.MatchedBy(
		func(query *q.Query) bool {
			return query.Keywords["username"] == "existing"
		})).Return([]*models.User{existingUser}, nil)
	m.dao.On("List", mock.Anything, testifymock.MatchedBy(
		func(query *q.Query) bool {
			return query.Keywords["username"] != "existing"
		})).Return([]*models.User{}, nil)

	{
		newUser := &models.User{
			Username: "newUser",
			Email:    "newUser@mytest.com",
			Realname: "newUser",
		}
		err := m.mgr.Onboard(context.Background(), newUser)
		m.Nil(err)
		m.Equal(newID, newUser.UserID)
		m.Equal(newUser.Username, newUser.Username)
	}
	{
		newUser := &models.User{
			Username: "existing",
			Email:    "existing@mytest.com",
			Realname: "existing",
		}
		err := m.mgr.Onboard(context.Background(), newUser)
		m.Nil(err)
		m.Equal(existingUser.Username, newUser.Username)
		m.Equal(existingUser.Email, newUser.Email)
		m.Equal(existingUser.UserID, newUser.UserID)
	}
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
