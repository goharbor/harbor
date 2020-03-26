package match

import (
	"github.com/goharbor/harbor/src/lib/selector"
)

// ImmutableTagMatcher ...
type ImmutableTagMatcher interface {
	// Match whether the candidate is in the immutable list
	Match(pid int64, c selector.Candidate) (bool, error)
}
