package cache

import (
	"errors"
	"github.com/goharbor/harbor/src/pkg/art"
)

// Cache ...
type Cache interface {
	// Set add a immutable to the project immutable list
	Set(pid int64, imc IMCandidate) error

	// Stat check whether the tag is immutable
	Stat(pid int64, repository string, tag string) (bool, error)

	// SetMultiple a list of immutable tags in project
	SetMultiple(pid int64, icands []IMCandidate) error

	// Clear remove the tag from the project immutable list
	Clear(pid int64, imc IMCandidate) error

	// Flush remove all of immutable tags of a specific project
	Flush(pid int64) error
}

// IMCandidate ...
type IMCandidate struct {
	art.Candidate
	Immutable bool
}

// ErrTagUnknown ...
var ErrTagUnknown = errors.New("tag unknown")

// ErrRepoUnknown ...
var ErrRepoUnknown = errors.New("repository unknown")
