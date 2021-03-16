package immutable

import (
	"context"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/stretchr/testify/mock"
)

// FakeMatcher ...
type FakeMatcher struct {
	mock.Mock
}

// Match ...
func (f *FakeMatcher) Match(ctx context.Context, pid int64, c selector.Candidate) (bool, error) {
	args := f.Called()
	return args.Bool(0), args.Error(1)
}
