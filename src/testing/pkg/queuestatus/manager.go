// Code generated by mockery v2.22.1. DO NOT EDIT.

package queuestatus

import (
	context "context"

	model "github.com/goharbor/harbor/src/pkg/queuestatus/model"
	mock "github.com/stretchr/testify/mock"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// AllJobTypeStatus provides a mock function with given fields: ctx
func (_m *Manager) AllJobTypeStatus(ctx context.Context) (map[string]bool, error) {
	ret := _m.Called(ctx)

	var r0 map[string]bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (map[string]bool, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) map[string]bool); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]bool)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateOrUpdate provides a mock function with given fields: ctx, status
func (_m *Manager) CreateOrUpdate(ctx context.Context, status *model.JobQueueStatus) (int64, error) {
	ret := _m.Called(ctx, status)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.JobQueueStatus) (int64, error)); ok {
		return rf(ctx, status)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.JobQueueStatus) int64); ok {
		r0 = rf(ctx, status)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.JobQueueStatus) error); ok {
		r1 = rf(ctx, status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, jobType
func (_m *Manager) Get(ctx context.Context, jobType string) (*model.JobQueueStatus, error) {
	ret := _m.Called(ctx, jobType)

	var r0 *model.JobQueueStatus
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.JobQueueStatus, error)); ok {
		return rf(ctx, jobType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.JobQueueStatus); ok {
		r0 = rf(ctx, jobType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.JobQueueStatus)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, jobType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx
func (_m *Manager) List(ctx context.Context) ([]*model.JobQueueStatus, error) {
	ret := _m.Called(ctx)

	var r0 []*model.JobQueueStatus
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*model.JobQueueStatus, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*model.JobQueueStatus); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.JobQueueStatus)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStatus provides a mock function with given fields: ctx, jobType, paused
func (_m *Manager) UpdateStatus(ctx context.Context, jobType string, paused bool) error {
	ret := _m.Called(ctx, jobType, paused)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) error); ok {
		r0 = rf(ctx, jobType, paused)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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
