// Code generated by mockery v2.46.2. DO NOT EDIT.

package jobmonitor

import (
	context "context"

	jobmonitor "github.com/goharbor/harbor/src/pkg/jobmonitor"
	mock "github.com/stretchr/testify/mock"
)

// QueueManager is an autogenerated mock type for the QueueManager type
type QueueManager struct {
	mock.Mock
}

// List provides a mock function with given fields: ctx, monitClient
func (_m *QueueManager) List(ctx context.Context, monitClient jobmonitor.JobServiceMonitorClient) ([]*jobmonitor.Queue, error) {
	ret := _m.Called(ctx, monitClient)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []*jobmonitor.Queue
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, jobmonitor.JobServiceMonitorClient) ([]*jobmonitor.Queue, error)); ok {
		return rf(ctx, monitClient)
	}
	if rf, ok := ret.Get(0).(func(context.Context, jobmonitor.JobServiceMonitorClient) []*jobmonitor.Queue); ok {
		r0 = rf(ctx, monitClient)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*jobmonitor.Queue)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, jobmonitor.JobServiceMonitorClient) error); ok {
		r1 = rf(ctx, monitClient)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewQueueManager creates a new instance of QueueManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewQueueManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *QueueManager {
	mock := &QueueManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
