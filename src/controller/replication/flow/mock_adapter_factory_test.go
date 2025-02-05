// Code generated by mockery v2.51.0. DO NOT EDIT.

package flow

import (
	adapter "github.com/goharbor/harbor/src/pkg/reg/adapter"
	mock "github.com/stretchr/testify/mock"

	model "github.com/goharbor/harbor/src/pkg/reg/model"
)

// mockFactory is an autogenerated mock type for the Factory type
type mockFactory struct {
	mock.Mock
}

// AdapterPattern provides a mock function with no fields
func (_m *mockFactory) AdapterPattern() *model.AdapterPattern {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for AdapterPattern")
	}

	var r0 *model.AdapterPattern
	if rf, ok := ret.Get(0).(func() *model.AdapterPattern); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AdapterPattern)
		}
	}

	return r0
}

// Create provides a mock function with given fields: _a0
func (_m *mockFactory) Create(_a0 *model.Registry) (adapter.Adapter, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 adapter.Adapter
	var r1 error
	if rf, ok := ret.Get(0).(func(*model.Registry) (adapter.Adapter, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*model.Registry) adapter.Adapter); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(adapter.Adapter)
		}
	}

	if rf, ok := ret.Get(1).(func(*model.Registry) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// newMockFactory creates a new instance of mockFactory. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockFactory(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockFactory {
	mock := &mockFactory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
