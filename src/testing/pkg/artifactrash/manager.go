package artifactrash

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
)

// FakeManager is a fake tag manager that implement the src/pkg/tag.Manager interface
type FakeManager struct {
	mock.Mock
}

// Create ...
func (f *FakeManager) Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// Delete ...
func (f *FakeManager) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}

// Filter ...
func (f *FakeManager) Filter(ctx context.Context, timeWindow int64) (arts []model.ArtifactTrash, err error) {
	args := f.Called()
	return args.Get(0).([]model.ArtifactTrash), args.Error(1)
}

// Flush ...
func (f *FakeManager) Flush(ctx context.Context, timeWindow int64) (err error) {
	args := f.Called()
	return args.Error(0)
}
