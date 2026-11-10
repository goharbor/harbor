package user

import (
	"context"
	"fmt"
	"testing"
	"time"

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
			return u.UserID == 9 && u.SysAdminFlag
		}), "sysadmin_flag").Return(nil)
	err := m.mgr.SetSysAdminFlag(context.Background(), id, true)
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

func (m *mgrTestSuite) TestMatchLocalPasswordUsesVerifyCache() {
	salt := "test-salt"
	stored := &models.User{
		UserID:          7,
		Username:        "alice",
		Salt:            salt,
		Password:        utils.Encrypt("Harbor12345", salt, utils.PBKDF2SHA256),
		PasswordVersion: utils.PBKDF2SHA256,
	}
	// MatchLocalPassword clears Password on the returned entry, so hand out a
	// fresh copy of the DB row on every call, as the real DAO does.
	m.dao.On("List", mock.Anything, mock.Anything).Return(
		func(context.Context, *q.Query) []*models.User {
			u := *stored
			return []*models.User{&u}
		}, nil)

	cache := newVerifyCache(time.Minute, 10)
	mgr := &manager{dao: m.dao, verifyCache: cache}
	ctx := context.Background()

	u, err := mgr.MatchLocalPassword(ctx, "alice", "Harbor12345")
	m.Require().Nil(err)
	m.Require().NotNil(u)
	m.Equal(7, u.UserID)
	m.Empty(u.Password)
	m.Len(cache.entries, 1, "a successful verification should be cached")

	// Served from the cache this time; result must be identical.
	u, err = mgr.MatchLocalPassword(ctx, "alice", "Harbor12345")
	m.Require().Nil(err)
	m.Require().NotNil(u)
	m.Equal(7, u.UserID)

	// A wrong password is never accepted and never cached.
	u, err = mgr.MatchLocalPassword(ctx, "alice", "wrong")
	m.Nil(err)
	m.Nil(u)
	m.Len(cache.entries, 1)

	// After a password change the stored salt and hash differ, so the old
	// password must stop working immediately despite the cached entry.
	stored.Salt = "new-salt"
	stored.Password = utils.Encrypt("NewPassw0rd", stored.Salt, utils.PBKDF2SHA256)
	u, err = mgr.MatchLocalPassword(ctx, "alice", "Harbor12345")
	m.Nil(err)
	m.Nil(u)
	u, err = mgr.MatchLocalPassword(ctx, "alice", "NewPassw0rd")
	m.Require().Nil(err)
	m.Require().NotNil(u)
}

func (m *mgrTestSuite) TestMatchLocalPasswordUsesRedisVerifyCache() {
	salt := "test-salt"
	stored := &models.User{
		UserID:          8,
		Username:        "bob",
		Salt:            salt,
		Password:        utils.Encrypt("Harbor12345", salt, utils.PBKDF2SHA256),
		PasswordVersion: utils.PBKDF2SHA256,
	}
	m.dao.On("List", mock.Anything, mock.Anything).Return(
		func(context.Context, *q.Query) []*models.User {
			u := *stored
			return []*models.User{&u}
		}, nil)

	redis := newFakeRedis()
	podA := &manager{dao: m.dao, verifyCache: newTestRedisVerifyCache(redis, "secret-key-16chr")}
	podB := &manager{dao: m.dao, verifyCache: newTestRedisVerifyCache(redis, "secret-key-16chr")}
	ctx := context.Background()

	u, err := podA.MatchLocalPassword(ctx, "bob", "Harbor12345")
	m.Require().Nil(err)
	m.Require().NotNil(u)
	m.Equal(8, u.UserID)
	m.Len(redis.entries, 1, "a successful verification should be cached in Redis")

	// Another core instance reuses the entry.
	u, err = podB.MatchLocalPassword(ctx, "bob", "Harbor12345")
	m.Require().Nil(err)
	m.Require().NotNil(u)
	m.Equal(1, redis.saves, "a cache hit must not write to Redis again")

	// Wrong passwords are rejected and never cached.
	u, err = podB.MatchLocalPassword(ctx, "bob", "wrong")
	m.Nil(err)
	m.Nil(u)
	m.Len(redis.entries, 1)

	// A password change invalidates the cached entry on every instance.
	stored.Salt = "new-salt"
	stored.Password = utils.Encrypt("NewPassw0rd", stored.Salt, utils.PBKDF2SHA256)
	u, err = podB.MatchLocalPassword(ctx, "bob", "Harbor12345")
	m.Nil(err)
	m.Nil(u)

	// If Redis fails, verification still works through PBKDF2.
	redis.failing = true
	u, err = podA.MatchLocalPassword(ctx, "bob", "NewPassw0rd")
	m.Require().Nil(err)
	m.Require().NotNil(u)
	u, err = podA.MatchLocalPassword(ctx, "bob", "Harbor12345")
	m.Nil(err)
	m.Nil(u)
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
