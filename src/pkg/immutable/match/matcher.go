package match

import (
	"context"
	"github.com/goharbor/harbor/src/lib/selector"
)

// ImmutableTagMatcher ...
type ImmutableTagMatcher interface {
	// Match whether the candidate is in the immutable list
	Match(ctx context.Context, pid int64, c selector.Candidate) (bool, error)
}
