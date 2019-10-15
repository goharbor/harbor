package match

import (
	"github.com/goharbor/harbor/src/pkg/art"
)

// ImmutableTagMatcher ...
type ImmutableTagMatcher interface {
	// Match whether the candidate is in the immutable list
	Match(c art.Candidate) (bool, error)
}
