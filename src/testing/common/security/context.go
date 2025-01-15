// Code generated by mockery v2.51.0. DO NOT EDIT.

package security

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/goharbor/harbor/src/pkg/permission/types"
)

// Context is an autogenerated mock type for the Context type
type Context struct {
	mock.Mock
}

// Can provides a mock function with given fields: ctx, action, resource
func (_m *Context) Can(ctx context.Context, action types.Action, resource types.Resource) bool {
	ret := _m.Called(ctx, action, resource)

	if len(ret) == 0 {
		panic("no return value specified for Can")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, types.Action, types.Resource) bool); ok {
		r0 = rf(ctx, action, resource)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetUsername provides a mock function with no fields
func (_m *Context) GetUsername() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetUsername")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// IsAuthenticated provides a mock function with no fields
func (_m *Context) IsAuthenticated() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsAuthenticated")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsSolutionUser provides a mock function with no fields
func (_m *Context) IsSolutionUser() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsSolutionUser")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsSysAdmin provides a mock function with no fields
func (_m *Context) IsSysAdmin() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsSysAdmin")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Name provides a mock function with no fields
func (_m *Context) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NewContext creates a new instance of Context. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *Context {
	mock := &Context{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
