// Code generated by mockery v2.35.4. DO NOT EDIT.

package jobmonitor

import (
	context "context"

	jobmonitor "github.com/goharbor/harbor/src/pkg/jobmonitor"
	mock "github.com/stretchr/testify/mock"
)

// WorkerManager is an autogenerated mock type for the WorkerManager type
type WorkerManager struct {
	mock.Mock
}

// List provides a mock function with given fields: ctx, monitClient, poolID
func (_m *WorkerManager) List(ctx context.Context, monitClient jobmonitor.JobServiceMonitorClient, poolID string) ([]*jobmonitor.Worker, error) {
	ret := _m.Called(ctx, monitClient, poolID)

	var r0 []*jobmonitor.Worker
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, jobmonitor.JobServiceMonitorClient, string) ([]*jobmonitor.Worker, error)); ok {
		return rf(ctx, monitClient, poolID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, jobmonitor.JobServiceMonitorClient, string) []*jobmonitor.Worker); ok {
		r0 = rf(ctx, monitClient, poolID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*jobmonitor.Worker)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, jobmonitor.JobServiceMonitorClient, string) error); ok {
		r1 = rf(ctx, monitClient, poolID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewWorkerManager creates a new instance of WorkerManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWorkerManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *WorkerManager {
	mock := &WorkerManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
