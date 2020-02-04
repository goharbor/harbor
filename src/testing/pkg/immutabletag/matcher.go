package immutabletag

import (
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/stretchr/testify/mock"
)

// FakeMatcher ...
type FakeMatcher struct {
	mock.Mock
}

// Match ...
func (f *FakeMatcher) Match(pid int64, c art.Candidate) (bool, error) {
	args := f.Called()
	return args.Bool(0), args.Error(1)
}
