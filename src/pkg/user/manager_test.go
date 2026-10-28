package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/user/dao"
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
			return u.UserID == 9 && u.SysAdminFlag && u.SysAdminFlagSource == sysAdminFlagSourceManual
		}), "sysadmin_flag", "sysadmin_flag_source").Return(nil)
	err := m.mgr.SetSysAdminFlag(context.Background(), id, true)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *mgrTestSuite) TestSyncSysAdminFlagFromAuth_PromotesWhenNotManual() {
	id := 10
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*models.User{
		{UserID: id, SysAdminFlag: false, SysAdminFlagSource: ""},
	}, nil)
	m.dao.On("Update", mock.Anything, testifymock.MatchedBy(
		func(u *models.User) bool {
			return u.UserID == id && u.SysAdminFlag && u.SysAdminFlagSource == sysAdminFlagSourceSync
		}), "sysadmin_flag", "sysadmin_flag_source").Return(nil)
	err := m.mgr.SyncSysAdminFlagFromAuth(context.Background(), id, true)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *mgrTestSuite) TestSyncSysAdminFlagFromAuth_RevokesWhenSyncOwnedAndNoLongerInGroup() {
	id := 11
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*models.User{
		{UserID: id, SysAdminFlag: true, SysAdminFlagSource: sysAdminFlagSourceSync},
	}, nil)
	m.dao.On("Update", mock.Anything, testifymock.MatchedBy(
		func(u *models.User) bool {
			return u.UserID == id && !u.SysAdminFlag && u.SysAdminFlagSource == sysAdminFlagSourceSync
		}), "sysadmin_flag", "sysadmin_flag_source").Return(nil)
	err := m.mgr.SyncSysAdminFlagFromAuth(context.Background(), id, false)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *mgrTestSuite) TestSyncSysAdminFlagFromAuth_NeverOverridesManualGrant() {
	id := 12
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*models.User{
		{UserID: id, SysAdminFlag: true, SysAdminFlagSource: sysAdminFlagSourceManual},
	}, nil)
	// no Update call expected: a manual decision must never be touched by auth sync,
	// regardless of the live admin-group membership
	err := m.mgr.SyncSysAdminFlagFromAuth(context.Background(), id, false)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *mgrTestSuite) TestSyncSysAdminFlagFromAuth_NeverOverridesManualRevocation() {
	id := 13
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*models.User{
		{UserID: id, SysAdminFlag: false, SysAdminFlagSource: sysAdminFlagSourceManual},
	}, nil)
	// still in the admin group, but an operator explicitly revoked it manually - must stick
	err := m.mgr.SyncSysAdminFlagFromAuth(context.Background(), id, true)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *mgrTestSuite) TestSyncSysAdminFlagFromAuth_NoOpWhenAlreadyInSync() {
	id := 14
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*models.User{
		{UserID: id, SysAdminFlag: true, SysAdminFlagSource: sysAdminFlagSourceSync},
	}, nil)
	// already true and already sync-owned: no Update call expected
	err := m.mgr.SyncSysAdminFlagFromAuth(context.Background(), id, true)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *mgrTestSuite) TestSyncSysAdminFlagFromAuth_StampsSourceEvenWhenValueUnchanged() {
	id := 15
	// an ordinary non-admin user: value already matches, but source was never confirmed live
	// (empty). The write must still happen so the source becomes "sync", letting the Users
	// list report a confirmed "No" instead of "Unknown" for them going forward.
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*models.User{
		{UserID: id, SysAdminFlag: false, SysAdminFlagSource: ""},
	}, nil)
	m.dao.On("Update", mock.Anything, testifymock.MatchedBy(
		func(u *models.User) bool {
			return u.UserID == id && !u.SysAdminFlag && u.SysAdminFlagSource == sysAdminFlagSourceSync
		}), "sysadmin_flag", "sysadmin_flag_source").Return(nil)
	err := m.mgr.SyncSysAdminFlagFromAuth(context.Background(), id, false)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *mgrTestSuite) TestUserDeleteGDPR() {
	existingUser := &models.User{
		UserID:   123,
		Username: "existing",
		Email:    "existing@mytest.com",
		Realname: "RealName",
	}
	m.dao.On("List", mock.Anything, testifymock.MatchedBy(
		func(query *q.Query) bool {
			return query.Keywords["user_id"] == 123
		})).Return(
		[]*models.User{existingUser}, nil)

	m.dao.On("Update", mock.Anything, testifymock.MatchedBy(
		func(u *models.User) bool {
			return u.UserID == 123 &&
				u.Email == fmt.Sprintf("%s#%d", m.mgr.GenerateCheckSum("existing@mytest.com"), existingUser.UserID) &&
				u.Username == fmt.Sprintf("%s#%d", m.mgr.GenerateCheckSum("existing"), existingUser.UserID) &&
				u.Realname == fmt.Sprintf("%s#%d", m.mgr.GenerateCheckSum("RealName"), existingUser.UserID) &&
				u.Deleted == true
		}),
		"username",
		"email",
		"realname",
		"deleted",
	).Return(nil)

	err := m.mgr.DeleteGDPR(context.Background(), 123)
	m.Nil(err)
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
	assert.Equal(t, utils.PBKDF2SHA256, u.PasswordVersion)
	assert.Equal(t, utils.Encrypt(p, u.Salt, utils.PBKDF2SHA256), u.Password)
}

func (m *mgrTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything, testifymock.Anything).Return(3, nil)
	u := &models.User{
		Username: "test",
		Email:    "test@example.com",
		Realname: "test",
	}
	id, err := m.mgr.Create(context.Background(), u)
	m.Nil(err)
	m.Equal(3, id)
	m.Equal(u.Username, "test")

	u2 := &models.User{
		Username: "test,test",
		Email:    "test@example.com",
		Realname: "test",
	}

	id, err = m.mgr.Create(context.Background(), u2)
	m.Nil(err)
	m.Equal(3, id)
	m.Equal(u2.Username, "test_test", "username should be sanitized")
}
