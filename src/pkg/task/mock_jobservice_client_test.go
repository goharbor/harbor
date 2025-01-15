// Code generated by mockery v2.51.0. DO NOT EDIT.

package task

import (
	job "github.com/goharbor/harbor/src/jobservice/job"
	mock "github.com/stretchr/testify/mock"

	models "github.com/goharbor/harbor/src/common/job/models"
)

// mockJobserviceClient is an autogenerated mock type for the Client type
type mockJobserviceClient struct {
	mock.Mock
}

// GetExecutions provides a mock function with given fields: uuid
func (_m *mockJobserviceClient) GetExecutions(uuid string) ([]job.Stats, error) {
	ret := _m.Called(uuid)

	if len(ret) == 0 {
		panic("no return value specified for GetExecutions")
	}

	var r0 []job.Stats
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]job.Stats, error)); ok {
		return rf(uuid)
	}
	if rf, ok := ret.Get(0).(func(string) []job.Stats); ok {
		r0 = rf(uuid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]job.Stats)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uuid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetJobLog provides a mock function with given fields: uuid
func (_m *mockJobserviceClient) GetJobLog(uuid string) ([]byte, error) {
	ret := _m.Called(uuid)

	if len(ret) == 0 {
		panic("no return value specified for GetJobLog")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]byte, error)); ok {
		return rf(uuid)
	}
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(uuid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uuid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetJobServiceConfig provides a mock function with no fields
func (_m *mockJobserviceClient) GetJobServiceConfig() (*job.Config, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetJobServiceConfig")
	}

	var r0 *job.Config
	var r1 error
	if rf, ok := ret.Get(0).(func() (*job.Config, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *job.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*job.Config)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PostAction provides a mock function with given fields: uuid, action
func (_m *mockJobserviceClient) PostAction(uuid string, action string) error {
	ret := _m.Called(uuid, action)

	if len(ret) == 0 {
		panic("no return value specified for PostAction")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(uuid, action)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitJob provides a mock function with given fields: _a0
func (_m *mockJobserviceClient) SubmitJob(_a0 *models.JobData) (string, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for SubmitJob")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(*models.JobData) (string, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*models.JobData) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(*models.JobData) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// newMockJobserviceClient creates a new instance of mockJobserviceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockJobserviceClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockJobserviceClient {
	mock := &mockJobserviceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
