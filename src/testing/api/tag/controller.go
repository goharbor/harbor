package tag

import (
	"context"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/stretchr/testify/mock"
)

// FakeController is a fake artifact controller that implement src/api/tag.Controller interface
type FakeController struct {
	mock.Mock
}

// Ensure ...
func (f *FakeController) Ensure(ctx context.Context, repositoryID, artifactID int64, name string) error {
	args := f.Called()
	return args.Error(0)
}

// Count ...
func (f *FakeController) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// List ...
func (f *FakeController) List(ctx context.Context, query *q.Query, option *tag.Option) ([]*tag.Tag, error) {
	args := f.Called()
	var tags []*tag.Tag
	if args.Get(0) != nil {
		tags = args.Get(0).([]*tag.Tag)
	}
	return tags, args.Error(1)
}

// Get ...
func (f *FakeController) Get(ctx context.Context, id int64, option *tag.Option) (*tag.Tag, error) {
	args := f.Called()
	var tg *tag.Tag
	if args.Get(0) != nil {
		tg = args.Get(0).(*tag.Tag)
	}
	return tg, args.Error(1)
}

// Create ...
func (f *FakeController) Create(ctx context.Context, tag *tag.Tag) (id int64, err error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}

// Update ...
func (f *FakeController) Update(ctx context.Context, tag *tag.Tag, props ...string) (err error) {
	args := f.Called()
	return args.Error(0)
}

// Delete ...
func (f *FakeController) Delete(ctx context.Context, id int64) (err error) {
	args := f.Called()
	return args.Error(0)
}

// DeleteTags ...
func (f *FakeController) DeleteTags(ctx context.Context, ids []int64) (err error) {
	args := f.Called()
	return args.Error(0)
}
