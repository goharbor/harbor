// Code generated by mockery v2.22.1. DO NOT EDIT.

package artifact

import (
	context "context"

	artifact "github.com/goharbor/harbor/src/controller/artifact"

	mock "github.com/stretchr/testify/mock"

	processor "github.com/goharbor/harbor/src/controller/artifact/processor"

	q "github.com/goharbor/harbor/src/lib/q"

	time "time"
)

// Controller is an autogenerated mock type for the Controller type
type Controller struct {
	mock.Mock
}

// AddLabel provides a mock function with given fields: ctx, artifactID, labelID
func (_m *Controller) AddLabel(ctx context.Context, artifactID int64, labelID int64) error {
	ret := _m.Called(ctx, artifactID, labelID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) error); ok {
		r0 = rf(ctx, artifactID, labelID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Copy provides a mock function with given fields: ctx, srcRepo, reference, dstRepo
func (_m *Controller) Copy(ctx context.Context, srcRepo string, reference string, dstRepo string) (int64, error) {
	ret := _m.Called(ctx, srcRepo, reference, dstRepo)

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (int64, error)); ok {
		return rf(ctx, srcRepo, reference, dstRepo)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) int64); ok {
		r0 = rf(ctx, srcRepo, reference, dstRepo)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, srcRepo, reference, dstRepo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Count provides a mock function with given fields: ctx, query
func (_m *Controller) Count(ctx context.Context, query *q.Query) (int64, error) {
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

// Delete provides a mock function with given fields: ctx, id
func (_m *Controller) Delete(ctx context.Context, id int64) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Ensure provides a mock function with given fields: ctx, repository, digest, option
func (_m *Controller) Ensure(ctx context.Context, repository string, digest string, option *artifact.ArtOption) (bool, int64, error) {
	ret := _m.Called(ctx, repository, digest, option)

	var r0 bool
	var r1 int64
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *artifact.ArtOption) (bool, int64, error)); ok {
		return rf(ctx, repository, digest, option)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *artifact.ArtOption) bool); ok {
		r0 = rf(ctx, repository, digest, option)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, *artifact.ArtOption) int64); ok {
		r1 = rf(ctx, repository, digest, option)
	} else {
		r1 = ret.Get(1).(int64)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, string, *artifact.ArtOption) error); ok {
		r2 = rf(ctx, repository, digest, option)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Get provides a mock function with given fields: ctx, id, option
func (_m *Controller) Get(ctx context.Context, id int64, option *artifact.Option) (*artifact.Artifact, error) {
	ret := _m.Called(ctx, id, option)

	var r0 *artifact.Artifact
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, *artifact.Option) (*artifact.Artifact, error)); ok {
		return rf(ctx, id, option)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, *artifact.Option) *artifact.Artifact); ok {
		r0 = rf(ctx, id, option)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*artifact.Artifact)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, *artifact.Option) error); ok {
		r1 = rf(ctx, id, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAddition provides a mock function with given fields: ctx, artifactID, additionType
func (_m *Controller) GetAddition(ctx context.Context, artifactID int64, additionType string) (*processor.Addition, error) {
	ret := _m.Called(ctx, artifactID, additionType)

	var r0 *processor.Addition
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) (*processor.Addition, error)); ok {
		return rf(ctx, artifactID, additionType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) *processor.Addition); ok {
		r0 = rf(ctx, artifactID, additionType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*processor.Addition)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, string) error); ok {
		r1 = rf(ctx, artifactID, additionType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByReference provides a mock function with given fields: ctx, repository, reference, option
func (_m *Controller) GetByReference(ctx context.Context, repository string, reference string, option *artifact.Option) (*artifact.Artifact, error) {
	ret := _m.Called(ctx, repository, reference, option)

	var r0 *artifact.Artifact
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *artifact.Option) (*artifact.Artifact, error)); ok {
		return rf(ctx, repository, reference, option)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *artifact.Option) *artifact.Artifact); ok {
		r0 = rf(ctx, repository, reference, option)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*artifact.Artifact)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, *artifact.Option) error); ok {
		r1 = rf(ctx, repository, reference, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, query, option
func (_m *Controller) List(ctx context.Context, query *q.Query, option *artifact.Option) ([]*artifact.Artifact, error) {
	ret := _m.Called(ctx, query, option)

	var r0 []*artifact.Artifact
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query, *artifact.Option) ([]*artifact.Artifact, error)); ok {
		return rf(ctx, query, option)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *q.Query, *artifact.Option) []*artifact.Artifact); ok {
		r0 = rf(ctx, query, option)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*artifact.Artifact)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *q.Query, *artifact.Option) error); ok {
		r1 = rf(ctx, query, option)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveLabel provides a mock function with given fields: ctx, artifactID, labelID
func (_m *Controller) RemoveLabel(ctx context.Context, artifactID int64, labelID int64) error {
	ret := _m.Called(ctx, artifactID, labelID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) error); ok {
		r0 = rf(ctx, artifactID, labelID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdatePullTime provides a mock function with given fields: ctx, artifactID, tagID, _a3
func (_m *Controller) UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, _a3 time.Time) error {
	ret := _m.Called(ctx, artifactID, tagID, _a3)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64, time.Time) error); ok {
		r0 = rf(ctx, artifactID, tagID, _a3)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Walk provides a mock function with given fields: ctx, root, walkFn, option
func (_m *Controller) Walk(ctx context.Context, root *artifact.Artifact, walkFn func(*artifact.Artifact) error, option *artifact.Option) error {
	ret := _m.Called(ctx, root, walkFn, option)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *artifact.Artifact, func(*artifact.Artifact) error, *artifact.Option) error); ok {
		r0 = rf(ctx, root, walkFn, option)
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
