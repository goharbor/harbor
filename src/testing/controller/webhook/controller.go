// Code generated by mockery v2.35.4. DO NOT EDIT.

package webhook

import (
	context "context"

	model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	mock "github.com/stretchr/testify/mock"

	q "github.com/goharbor/harbor/src/lib/q"

	task "github.com/goharbor/harbor/src/pkg/task"

	time "time"
)

// Controller is an autogenerated mock type for the Controller type
type Controller struct {
	mock.Mock
}

// CountExecutions provides a mock function with given fields: ctx, policyID, query
func (_m *Controller) CountExecutions(ctx context.Context, policyID int64, query *q.Query) (int64, error) {
	ret := _m.Called(ctx, policyID, query)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) (int64, error)); ok {
		return rf(ctx, policyID, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) int64); ok {
		r0 = rf(ctx, policyID, query)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, *q.Query) error); ok {
		r1 = rf(ctx, policyID, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CountPolicies provides a mock function with given fields: ctx, query
func (_m *Controller) CountPolicies(ctx context.Context, query *q.Query) (int64, error) {
	ret := _m.Called(ctx, query)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) (int64, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) int64); ok {
		r0 = rf(ctx, query)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *q.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CountTasks provides a mock function with given fields: ctx, execID, query
func (_m *Controller) CountTasks(ctx context.Context, execID int64, query *q.Query) (int64, error) {
	ret := _m.Called(ctx, execID, query)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) (int64, error)); ok {
		return rf(ctx, execID, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) int64); ok {
		r0 = rf(ctx, execID, query)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, *q.Query) error); ok {
		r1 = rf(ctx, execID, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreatePolicy provides a mock function with given fields: ctx, policy
func (_m *Controller) CreatePolicy(ctx context.Context, policy *model.Policy) (int64, error) {
	ret := _m.Called(ctx, policy)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Policy) (int64, error)); ok {
		return rf(ctx, policy)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.Policy) int64); ok {
		r0 = rf(ctx, policy)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.Policy) error); ok {
		r1 = rf(ctx, policy)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeletePolicy provides a mock function with given fields: ctx, policyID
func (_m *Controller) DeletePolicy(ctx context.Context, policyID int64) error {
	ret := _m.Called(ctx, policyID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, policyID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetLastTriggerTime provides a mock function with given fields: ctx, eventType, policyID
func (_m *Controller) GetLastTriggerTime(ctx context.Context, eventType string, policyID int64) (time.Time, error) {
	ret := _m.Called(ctx, eventType, policyID)

	var r0 time.Time
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) (time.Time, error)); ok {
		return rf(ctx, eventType, policyID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) time.Time); ok {
		r0 = rf(ctx, eventType, policyID)
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, int64) error); ok {
		r1 = rf(ctx, eventType, policyID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPolicy provides a mock function with given fields: ctx, id
func (_m *Controller) GetPolicy(ctx context.Context, id int64) (*model.Policy, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*model.Policy, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *model.Policy); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRelatedPolices provides a mock function with given fields: ctx, projectID, eventType
func (_m *Controller) GetRelatedPolices(ctx context.Context, projectID int64, eventType string) ([]*model.Policy, error) {
	ret := _m.Called(ctx, projectID, eventType)

	var r0 []*model.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) ([]*model.Policy, error)); ok {
		return rf(ctx, projectID, eventType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) []*model.Policy); ok {
		r0 = rf(ctx, projectID, eventType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, string) error); ok {
		r1 = rf(ctx, projectID, eventType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTask provides a mock function with given fields: ctx, taskID
func (_m *Controller) GetTask(ctx context.Context, taskID int64) (*task.Task, error) {
	ret := _m.Called(ctx, taskID)

	var r0 *task.Task
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (*task.Task, error)); ok {
		return rf(ctx, taskID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) *task.Task); ok {
		r0 = rf(ctx, taskID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*task.Task)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, taskID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTaskLog provides a mock function with given fields: ctx, taskID
func (_m *Controller) GetTaskLog(ctx context.Context, taskID int64) ([]byte, error) {
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

// ListExecutions provides a mock function with given fields: ctx, policyID, query
func (_m *Controller) ListExecutions(ctx context.Context, policyID int64, query *q.Query) ([]*task.Execution, error) {
	ret := _m.Called(ctx, policyID, query)

	var r0 []*task.Execution
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) ([]*task.Execution, error)); ok {
		return rf(ctx, policyID, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) []*task.Execution); ok {
		r0 = rf(ctx, policyID, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*task.Execution)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, *q.Query) error); ok {
		r1 = rf(ctx, policyID, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPolicies provides a mock function with given fields: ctx, query
func (_m *Controller) ListPolicies(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	ret := _m.Called(ctx, query)

	var r0 []*model.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) ([]*model.Policy, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query) []*model.Policy); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *q.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListTasks provides a mock function with given fields: ctx, execID, query
func (_m *Controller) ListTasks(ctx context.Context, execID int64, query *q.Query) ([]*task.Task, error) {
	ret := _m.Called(ctx, execID, query)

	var r0 []*task.Task
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) ([]*task.Task, error)); ok {
		return rf(ctx, execID, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, *q.Query) []*task.Task); ok {
		r0 = rf(ctx, execID, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*task.Task)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, *q.Query) error); ok {
		r1 = rf(ctx, execID, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdatePolicy provides a mock function with given fields: ctx, policy
func (_m *Controller) UpdatePolicy(ctx context.Context, policy *model.Policy) error {
	ret := _m.Called(ctx, policy)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Policy) error); ok {
		r0 = rf(ctx, policy)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewController creates a new instance of Controller. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewController(t interface {
	mock.TestingT
	Cleanup(func())
}) *Controller {
	mock := &Controller{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
