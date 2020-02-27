package match

import (
	"github.com/goharbor/harbor/src/pkg/artifactselector"
)

// ImmutableTagMatcher ...
type ImmutableTagMatcher interface {
	// Match whether the candidate is in the immutable list
	Match(pid int64, c artifactselector.Candidate) (bool, error)
}
