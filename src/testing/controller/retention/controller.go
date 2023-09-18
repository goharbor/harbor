// Code generated by mockery v2.22.1. DO NOT EDIT.

package retention

import (
	context "context"

	pkgretention "github.com/goharbor/harbor/src/pkg/retention"
	mock "github.com/stretchr/testify/mock"

	policy "github.com/goharbor/harbor/src/pkg/retention/policy"

	q "github.com/goharbor/harbor/src/lib/q"
)

// Controller is an autogenerated mock type for the Controller type
type Controller struct {
	mock.Mock
}

// CreateRetention provides a mock function with given fields: ctx, p
func (_m *Controller) CreateRetention(ctx context.Context, p *policy.Metadata) (int64, error) {
	ret := _m.Called(ctx, p)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *policy.Metadata) (int64, error)); ok {
		return rf(ctx, p)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *policy.Metadata) int64); ok {
		r0 = rf(ctx, p)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *policy.Metadata) error); ok {
		r1 = rf(ctx, p)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteRetention provides a mock function with given fields: ctx, id
func (_m *Controller) DeleteRetention(ctx context.Context, id int64) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteRetentionByProject provides a mock function with given fields: ctx, projectID
func (_m *Controller) DeleteRetentionByProject(ctx context.Context, projectID int64) error {
	ret := _m.Called(ctx, projectID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, projectID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetRetention provides a mock function with given fields: ctx, id
func (_m *Controller) GetRetention(ctx context.Context, id int64) (*policy.Metadata, error) {
	ret := _m.Called(ctx, id)

	var r0 *policy.Metadata
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*policy.Metadata, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *policy.Metadata); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*policy.Metadata)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRetentionExec provides a mock function with given fields: ctx, eid
func (_m *Controller) GetRetentionExec(ctx context.Context, eid int64) (*pkgretention.Execution, error) {
	ret := _m.Called(ctx, eid)

	var r0 *pkgretention.Execution
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*pkgretention.Execution, error)); ok {
		return rf(ctx, eid)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *pkgretention.Execution); ok {
		r0 = rf(ctx, eid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkgretention.Execution)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, eid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRetentionExecTask provides a mock function with given fields: ctx, taskID
func (_m *Controller) GetRetentionExecTask(ctx context.Context, taskID int64) (*pkgretention.Task, error) {
	ret := _m.Called(ctx, taskID)

	var r0 *pkgretention.Task
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*pkgretention.Task, error)); ok {
		return rf(ctx, taskID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *pkgretention.Task); ok {
		r0 = rf(ctx, taskID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pkgretention.Task)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, taskID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRetentionExecTaskLog provides a mock function with given fields: ctx, taskID
func (_m *Controller) GetRetentionExecTaskLog(ctx context.Context, taskID int64) ([]byte, error) {
	ret := _m.Called(ctx, taskID)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) ([]byte, error)); ok {
		return rf(ctx, taskID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) []byte); ok {
		r0 = rf(ctx, taskID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, taskID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTotalOfRetentionExecTasks provides a mock function with given fields: ctx, executionID
func (_m *Controller) GetTotalOfRetentionExecTasks(ctx context.Context, executionID int64) (int64, error) {
	ret := _m.Called(ctx, executionID)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (int64, error)); ok {
		return rf(ctx, executionID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) int64); ok {
		r0 = rf(ctx, executionID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, executionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTotalOfRetentionExecs provides a mock function with given fields: ctx, policyID
func (_m *Controller) GetTotalOfRetentionExecs(ctx context.Context, policyID int64) (int64, error) {
	ret := _m.Called(ctx, policyID)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (int64, error)); ok {
		return rf(ctx, policyID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) int64); ok {
		r0 = rf(ctx, policyID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, policyID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListRetentionExecTasks provides a mock function with given fields: ctx, executionID, query
func (_m *Controller) ListRetentionExecTasks(ctx context.Context, executionID int64, query *q.Query) ([]*pkgretention.Task, error) {
	ret := _m.Called(ctx, executionID, query)

	var r0 []*pkgretention.Task
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) ([]*pkgretention.Task, error)); ok {
		return rf(ctx, executionID, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) []*pkgretention.Task); ok {
		r0 = rf(ctx, executionID, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*pkgretention.Task)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, *q.Query) error); ok {
		r1 = rf(ctx, executionID, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListRetentionExecs provides a mock function with given fields: ctx, policyID, query
func (_m *Controller) ListRetentionExecs(ctx context.Context, policyID int64, query *q.Query) ([]*pkgretention.Execution, error) {
	ret := _m.Called(ctx, policyID, query)

	var r0 []*pkgretention.Execution
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) ([]*pkgretention.Execution, error)); ok {
		return rf(ctx, policyID, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) []*pkgretention.Execution); ok {
		r0 = rf(ctx, policyID, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*pkgretention.Execution)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, *q.Query) error); ok {
		r1 = rf(ctx, policyID, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OperateRetentionExec provides a mock function with given fields: ctx, eid, action
func (_m *Controller) OperateRetentionExec(ctx context.Context, eid int64, action string) error {
	ret := _m.Called(ctx, eid, action)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) error); ok {
		r0 = rf(ctx, eid, action)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TriggerRetentionExec provides a mock function with given fields: ctx, policyID, trigger, dryRun
func (_m *Controller) TriggerRetentionExec(ctx context.Context, policyID int64, trigger string, dryRun bool) (int64, error) {
	ret := _m.Called(ctx, policyID, trigger, dryRun)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string, bool) (int64, error)); ok {
		return rf(ctx, policyID, trigger, dryRun)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, string, bool) int64); ok {
		r0 = rf(ctx, policyID, trigger, dryRun)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, string, bool) error); ok {
		r1 = rf(ctx, policyID, trigger, dryRun)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateRetention provides a mock function with given fields: ctx, p
func (_m *Controller) UpdateRetention(ctx context.Context, p *policy.Metadata) error {
	ret := _m.Called(ctx, p)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *policy.Metadata) error); ok {
		r0 = rf(ctx, p)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewController interface {
	mock.TestingT
	Cleanup(func())
}

// NewController creates a new instance of Controller. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewController(t mockConstructorTestingTNewController) *Controller {
	mock := &Controller{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
