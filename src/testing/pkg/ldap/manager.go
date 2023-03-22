// Code generated by mockery v2.22.1. DO NOT EDIT.

package ldap

import (
	context "context"

	ldap "github.com/goharbor/harbor/src/pkg/ldap"
	mock "github.com/stretchr/testify/mock"

	model "github.com/goharbor/harbor/src/pkg/ldap/model"

	models "github.com/goharbor/harbor/src/lib/config/models"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// ImportUser provides a mock function with given fields: ctx, sess, ldapImportUsers
func (_m *Manager) ImportUser(ctx context.Context, sess *ldap.Session, ldapImportUsers []string) ([]model.FailedImportUser, error) {
	ret := _m.Called(ctx, sess, ldapImportUsers)

	var r0 []model.FailedImportUser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *ldap.Session, []string) ([]model.FailedImportUser, error)); ok {
		return rf(ctx, sess, ldapImportUsers)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *ldap.Session, []string) []model.FailedImportUser); ok {
		r0 = rf(ctx, sess, ldapImportUsers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.FailedImportUser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *ldap.Session, []string) error); ok {
		r1 = rf(ctx, sess, ldapImportUsers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ping provides a mock function with given fields: ctx, cfg
func (_m *Manager) Ping(ctx context.Context, cfg models.LdapConf) (bool, error) {
	ret := _m.Called(ctx, cfg)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, models.LdapConf) (bool, error)); ok {
		return rf(ctx, cfg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, models.LdapConf) bool); ok {
		r0 = rf(ctx, cfg)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, models.LdapConf) error); ok {
		r1 = rf(ctx, cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchGroup provides a mock function with given fields: ctx, sess, groupName, groupDN
func (_m *Manager) SearchGroup(ctx context.Context, sess *ldap.Session, groupName string, groupDN string) ([]model.Group, error) {
	ret := _m.Called(ctx, sess, groupName, groupDN)

	var r0 []model.Group
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *ldap.Session, string, string) ([]model.Group, error)); ok {
		return rf(ctx, sess, groupName, groupDN)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *ldap.Session, string, string) []model.Group); ok {
		r0 = rf(ctx, sess, groupName, groupDN)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Group)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *ldap.Session, string, string) error); ok {
		r1 = rf(ctx, sess, groupName, groupDN)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchUser provides a mock function with given fields: ctx, sess, username
func (_m *Manager) SearchUser(ctx context.Context, sess *ldap.Session, username string) ([]model.User, error) {
	ret := _m.Called(ctx, sess, username)

	var r0 []model.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *ldap.Session, string) ([]model.User, error)); ok {
		return rf(ctx, sess, username)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *ldap.Session, string) []model.User); ok {
		r0 = rf(ctx, sess, username)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *ldap.Session, string) error); ok {
		r1 = rf(ctx, sess, username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewManager creates a new instance of Manager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewManager(t mockConstructorTestingTNewManager) *Manager {
	mock := &Manager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
