// Code generated by mockery v2.46.2. DO NOT EDIT.

package systemartifact

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	model "github.com/goharbor/harbor/src/pkg/systemartifact/model"

	systemartifact "github.com/goharbor/harbor/src/pkg/systemartifact"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Cleanup provides a mock function with given fields: ctx
func (_m *Manager) Cleanup(ctx context.Context) (int64, int64, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Cleanup")
	}

	var r0 int64
	var r1 int64
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context) (int64, int64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) int64); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Get(1).(int64)
	}

	if rf, ok := ret.Get(2).(func(context.Context) error); ok {
		r2 = rf(ctx)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Create provides a mock function with given fields: ctx, artifactRecord, reader
func (_m *Manager) Create(ctx context.Context, artifactRecord *model.SystemArtifact, reader io.Reader) (int64, error) {
	ret := _m.Called(ctx, artifactRecord, reader)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.SystemArtifact, io.Reader) (int64, error)); ok {
		return rf(ctx, artifactRecord, reader)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.SystemArtifact, io.Reader) int64); ok {
		r0 = rf(ctx, artifactRecord, reader)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.SystemArtifact, io.Reader) error); ok {
		r1 = rf(ctx, artifactRecord, reader)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, vendor, repository, digest
func (_m *Manager) Delete(ctx context.Context, vendor string, repository string, digest string) error {
	ret := _m.Called(ctx, vendor, repository, digest)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, vendor, repository, digest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, vendor, repository, digest
func (_m *Manager) Exists(ctx context.Context, vendor string, repository string, digest string) (bool, error) {
	ret := _m.Called(ctx, vendor, repository, digest)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (bool, error)); ok {
		return rf(ctx, vendor, repository, digest)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) bool); ok {
		r0 = rf(ctx, vendor, repository, digest)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, vendor, repository, digest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCleanupCriteria provides a mock function with given fields: vendor, artifactType
func (_m *Manager) GetCleanupCriteria(vendor string, artifactType string) systemartifact.Selector {
	ret := _m.Called(vendor, artifactType)

	if len(ret) == 0 {
		panic("no return value specified for GetCleanupCriteria")
	}

	var r0 systemartifact.Selector
	if rf, ok := ret.Get(0).(func(string, string) systemartifact.Selector); ok {
		r0 = rf(vendor, artifactType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(systemartifact.Selector)
		}
	}

	return r0
}

// GetStorageSize provides a mock function with given fields: ctx
func (_m *Manager) GetStorageSize(ctx context.Context) (int64, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetStorageSize")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (int64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSystemArtifactProjectNames provides a mock function with given fields:
func (_m *Manager) GetSystemArtifactProjectNames() []string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSystemArtifactProjectNames")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Read provides a mock function with given fields: ctx, vendor, repository, digest
func (_m *Manager) Read(ctx context.Context, vendor string, repository string, digest string) (io.ReadCloser, error) {
	ret := _m.Called(ctx, vendor, repository, digest)

	if len(ret) == 0 {
		panic("no return value specified for Read")
	}

	var r0 io.ReadCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (io.ReadCloser, error)); ok {
		return rf(ctx, vendor, repository, digest)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) io.ReadCloser); ok {
		r0 = rf(ctx, vendor, repository, digest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, vendor, repository, digest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegisterCleanupCriteria provides a mock function with given fields: vendor, artifactType, criteria
func (_m *Manager) RegisterCleanupCriteria(vendor string, artifactType string, criteria systemartifact.Selector) {
	_m.Called(vendor, artifactType, criteria)
}

// NewManager creates a new instance of Manager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *Manager {
	mock := &Manager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
